package ai_post_editor

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/dracory/blogadmin/ai_post_editor/templates"
	"github.com/dracory/blogadmin/blogai"
	"github.com/dracory/blogadmin/shared"
	"github.com/dracory/cdn"
	"github.com/dracory/customstore"
	"github.com/dracory/hb"
	"github.com/dracory/req"
)

const (
	ACTION_REGENERATE_SECTION   = "regenerate_section"
	ACTION_REGENERATE_IMAGE     = "regenerate_image"
	ACTION_CREATE_FINAL_POST    = "create_final_post"
	ACTION_SAVE_DRAFT           = "save_draft"
	ACTION_REGENERATE_PARAGRAPH = "regenerate_paragraph"
	ACTION_LOAD_POST            = "load_post"
	ACTION_REGENERATE_SUMMARY   = "regenerate_summary"
	ACTION_REGENERATE_METAS     = "regenerate_metas"
)

// UiInterface defines the ai post editor controller's UI interface
type UiInterface interface {
	shared.UiInterface
	AiPostEditor(w http.ResponseWriter, r *http.Request)
}

// ui implements UiInterface
type ui struct {
	shared.UiBase
}

// UI creates a new ai post editor controller UI from the given config
func UI(config shared.UiConfig) UiInterface {
	return &ui{UiBase: shared.NewUiBase(config)}
}

// AiPostEditor handles the ai post editor controller requests
func (u *ui) AiPostEditor(w http.ResponseWriter, r *http.Request) {
	html := u.Handler(w, r)
	if html != "" {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(html))
	}
}

type pageData struct {
	Request    *http.Request
	BlogAiPost blogai.RecordPost
	Record     customstore.RecordInterface
}

// Handler processes the ai post editor controller request and returns HTML
func (u *ui) Handler(w http.ResponseWriter, r *http.Request) string {
	u.Logger().Info("Post Editor Handler called")

	data, errorMessage := u.prepareDataAndValidate(r)
	if errorMessage != "" {
		return shared.ErrorPopup(errorMessage).ToHTML()
	}

	action := req.GetStringTrimmed(r, "action")
	switch {
	case r.Method == http.MethodPost && action == ACTION_REGENERATE_SECTION:
		return u.onRegenerateSection(data)
	case r.Method == http.MethodPost && action == ACTION_REGENERATE_IMAGE:
		return u.onRegenerateImage(data)
	case r.Method == http.MethodPost && action == ACTION_REGENERATE_PARAGRAPH:
		return u.onRegenerateParagraph(data)
	case r.Method == http.MethodPost && action == ACTION_CREATE_FINAL_POST:
		return u.onCreateFinalPost(data)
	case r.Method == http.MethodPost && action == ACTION_SAVE_DRAFT:
		return u.onSaveDraft(data)
	case r.Method == http.MethodPost && action == ACTION_LOAD_POST:
		return u.onLoadPost(data)
	case r.Method == http.MethodPost && action == ACTION_REGENERATE_SUMMARY:
		return u.onRegenerateSummary(data)
	case r.Method == http.MethodPost && action == ACTION_REGENERATE_METAS:
		return u.onRegenerateMetas(data)
	}

	content := u.view(r, data)
	return u.Layout(w, r, "Edit & Save Blog Post", content.ToHTML(), struct {
		Styles     []string
		StyleURLs  []string
		Scripts    []string
		ScriptURLs []string
	}{
		ScriptURLs: []string{
			cdn.Sweetalert2_11(),
		},
	})
}

func (u *ui) buildPostMarkdownContent(_ *http.Request, record *blogai.RecordPost) string {
	content := "# " + record.Title + "\n\n"

	content += "## " + record.Introduction.Title + "\n\n"
	for _, paragraph := range record.Introduction.Paragraphs {
		content += paragraph + "\n\n"
	}

	for _, section := range record.Sections {
		content += "## " + section.Title + "\n\n"
		for _, paragraph := range section.Paragraphs {
			content += paragraph + "\n\n"
		}
	}

	content += "## " + record.Conclusion.Title + "\n\n"
	for _, paragraph := range record.Conclusion.Paragraphs {
		content += paragraph + "\n\n"
	}

	return content
}

func (u *ui) view(r *http.Request, data pageData) hb.TagInterface {
	linksHelper := shared.NewLinksFromRequest(r)

	header := hb.Heading1().HTML("Edit Blog Post")

	breadcrumbs := shared.Breadcrumbs([]shared.Breadcrumb{
		{
			Name: "Dashboard",
			URL:  shared.AdminHomeURL(r),
		},
		{
			Name: "Blog",
			URL:  shared.URLR(r, shared.CONTROLLER_DASHBOARD, nil),
		},
		{
			Name: "AI Tools",
			URL:  linksHelper.AiTools(nil),
		},
		{
			Name: "Post Editor",
			URL:  linksHelper.AiPostEditor(nil),
		},
		{
			Name: "Post Editor",
			URL:  linksHelper.AiPostEditor(map[string]string{"id": data.BlogAiPost.ID}),
		},
	})

	backButton := hb.A().
		Class("btn btn-secondary me-3").
		Href(linksHelper.AiPostGenerator(map[string]string{})).
		HTML("← Back to Post Generator")

	vueApp := hb.Raw(templates.Tpl("app.html", map[string]any{}))
	vueScript := hb.Script(templates.Tpl("app.js", map[string]any{
		"postJSON": data.BlogAiPost.ToJSON(),
		"id":       data.BlogAiPost.ID,
		"url":      linksHelper.AiPostEditor(map[string]string{}),
	}))
	vueStyles := hb.Style(templates.Tpl("app.css", nil))

	return hb.Div().
		Class("container min-vh-100 py-4 bg-light").
		Child(breadcrumbs).
		Child(hb.Div().
			Class("container").
			Child(header).
			Child(hb.Div().
				Class("d-flex justify-content-between mb-3").
				Child(backButton),
			).
			Child(shared.VueLoaderScript()).
			Child(vueApp).
			Child(vueScript).
			Child(vueStyles),
		)
}

func (u *ui) prepareDataAndValidate(r *http.Request) (pageData, string) {
	var (
		data pageData
		err  error
	)

	data.Request = r
	recordPostID := req.GetStringTrimmed(r, "id")
	if recordPostID == "" {
		return data, "Record Post ID is missing"
	}

	data.Record, err = u.CustomStore().RecordFindByID(recordPostID)
	if err != nil {
		u.Logger().Error("BlogAi. Post Editor. Prepare Data. Error finding record post", slog.String("error", err.Error()))
		return data, fmt.Sprintf("Failed to find record post: %s", err)
	}

	if data.Record == nil {
		u.Logger().Error("BlogAi. Post Editor. Prepare Data. Post record not found", slog.String("record_id", recordPostID))
		return data, "Post record not found"
	}

	if data.Record.Type() != blogai.POST_RECORD_TYPE {
		u.Logger().Error("BlogAi. Post Editor. Prepare Data. Invalid record type", slog.String("record_type", data.Record.Type()), slog.String("record_id", recordPostID))
		return data, "Invalid record type"
	}

	data.BlogAiPost, err = blogai.NewRecordPostFromCustomRecord(data.Record)
	if err != nil {
		u.Logger().Error("BlogAi. Post Editor. Prepare Data. Failed to parse blog record", slog.String("error", err.Error()))
		return data, fmt.Sprintf("Failed to parse blog record: %s", err)
	}

	return data, ""
}

// RecordFromJSON parses a JSON string into a blogai.RecordPost
func RecordFromJSON(jsonStr string) (*blogai.RecordPost, error) {
	type postData struct {
		Title           string                         `json:"title"`
		Subtitle        string                         `json:"subtitle,omitempty"`
		Summary         string                         `json:"summary,omitempty"`
		Introduction    blogai.PostContentIntroduction `json:"introduction"`
		Sections        []blogai.PostContentSection    `json:"sections"`
		Conclusion      blogai.PostContentConclusion   `json:"conclusion"`
		Keywords        []string                       `json:"keywords,omitempty"`
		MetaDescription string                         `json:"metaDescription,omitempty"`
		MetaKeywords    []string                       `json:"metaKeywords,omitempty"`
		MetaTitle       string                         `json:"metaTitle,omitempty"`
		Image           string                         `json:"image,omitempty"`
	}

	var record postData
	if err := json.Unmarshal([]byte(jsonStr), &record); err != nil {
		return nil, err
	}

	recordPost := blogai.RecordPost{}
	recordPost.Title = record.Title
	recordPost.Subtitle = record.Subtitle
	recordPost.Summary = record.Summary
	recordPost.Introduction = record.Introduction
	recordPost.Sections = record.Sections
	recordPost.Conclusion = record.Conclusion
	recordPost.Keywords = record.Keywords
	recordPost.MetaDescription = record.MetaDescription
	recordPost.MetaKeywords = record.MetaKeywords
	recordPost.MetaTitle = record.MetaTitle
	recordPost.Image = record.Image

	return &recordPost, nil
}
