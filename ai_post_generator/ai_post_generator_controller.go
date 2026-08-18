package ai_post_generator

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/dracory/blogadmin/shared"
	"github.com/dracory/blogadmin/blogai"
	"github.com/dracory/cdn"
	"github.com/dracory/customstore"
	"github.com/dracory/hb"
	"github.com/dracory/req"
)

const ACTION_GENERATE_POST = "generate_post"

// UiInterface defines the ai post generator controller's UI interface
type UiInterface interface {
	shared.UiInterface
	AiPostGenerator(w http.ResponseWriter, r *http.Request)
}

// ui implements UiInterface
type ui struct {
	shared.UiBase
}

// UI creates a new ai post generator controller UI from the given config
func UI(config shared.UiConfig) UiInterface {
	return &ui{UiBase: shared.NewUiBase(config)}
}

// AiPostGenerator handles the ai post generator controller requests
func (u *ui) AiPostGenerator(w http.ResponseWriter, r *http.Request) {
	html := u.Handler(w, r)
	if html != "" {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(html))
	}
}

type pageData struct {
	Request             *http.Request
	Action              string
	ApprovedBlogAiPosts []blogai.RecordPost
}

// Handler processes the ai post generator controller request and returns HTML
func (u *ui) Handler(w http.ResponseWriter, r *http.Request) string {
	data, errorMessage := u.prepareData(r)

	if errorMessage != "" {
		return shared.ErrorPopup(errorMessage).ToHTML()
	}

	if r.Method == http.MethodPost && data.Action == ACTION_GENERATE_POST {
		return u.onGeneratePost(r)
	}

	content := u.view(r, data)
	return u.Layout(w, r, "Post Generator", content.ToHTML(), struct {
		Styles     []string
		StyleURLs  []string
		Scripts    []string
		ScriptURLs []string
	}{
		ScriptURLs: []string{
			cdn.Htmx_2_0_0(),
			cdn.Sweetalert2_11(),
		},
	})
}

func (u *ui) view(r *http.Request, data pageData) hb.TagInterface {
	linksHelper := shared.NewLinksFromRequest(r)

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
			Name: "Post Generator",
			URL:  linksHelper.AiPostGenerator(nil),
		},
	})

	card := hb.Div().
		Class("card shadow-sm w-100 mb-5")
	card = card.Child(
		hb.Div().Class("card-body text-center p-4").
			Child(
				hb.Heading1().
					HTML("Post Generator").
					Class("h3 mb-3 fw-bold"),
			).
			Child(
				hb.Paragraph().
					HTML("Select an approved title below to generate a blog post.").
					Class("text-muted mb-4"),
			).
			Child(
				hb.Div().Class("text-start").
					Child(u.tableApprovedTitles(r, data)),
			),
	)

	return hb.Div().
		Class("container").
		Class("min-vh-100 py-4").
		Child(breadcrumbs).
		Child(card)
}

func (u *ui) prepareData(r *http.Request) (data pageData, errorMessage string) {
	data.Request = r
	data.Action = req.GetStringTrimmed(r, "action")

	customStore := u.CustomStore()
	if customStore == nil {
		return data, "custom store not configured"
	}

	approvedTitleRecords, err := customStore.RecordList(customstore.NewRecordQuery().
		SetType(blogai.POST_RECORD_TYPE).
		AddPayloadSearch(`"status":"approved"`).
		AddPayloadSearch(`"status":"draft"`))

	if err != nil {
		return data, fmt.Sprintf("Error fetching approved titles: %s", err.Error())
	}

	approvedBlogAiPosts := []blogai.RecordPost{}
	for _, record := range approvedTitleRecords {
		recordPost, err := blogai.NewRecordPostFromCustomRecord(record)
		if err != nil {
			u.Logger().Warn("Failed to parse custom record into RecordPost", slog.String("error", err.Error()))
			continue
		}

		if recordPost.Status == blogai.POST_STATUS_APPROVED || recordPost.Status == blogai.POST_STATUS_DRAFT {
			approvedBlogAiPosts = append(approvedBlogAiPosts, recordPost)
		}
	}

	data.ApprovedBlogAiPosts = approvedBlogAiPosts

	return data, ""
}
