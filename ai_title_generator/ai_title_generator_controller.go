package ai_title_generator

import (
	"embed"
	"fmt"
	"net/http"
	"strings"

	"github.com/dracory/base/htmx"
	"github.com/dracory/blogadmin/shared"
	"github.com/dracory/blogadmin/blogai"
	"github.com/dracory/cdn"
	"github.com/dracory/customstore"
	"github.com/dracory/hb"
	"github.com/dracory/req"
)

//go:embed settings_modal.html
//go:embed settings_modal.js
var settingsModalFiles embed.FS

const (
	ACTION_ADD_TITLE       = "add_title"
	ACTION_GENERATE_TITLES = "generate_titles"
	ACTION_APPROVE_TITLE   = "approve_title"
	ACTION_REJECT_TITLE    = "reject_title"
	ACTION_GENERATE_POST   = "generate_post"
	ACTION_DELETE_TITLE    = "delete_title"
	ACTION_SETTINGS_FETCH  = "settings-fetch-data"
	ACTION_SETTINGS_SUBMIT = "settings-submit"
)

// UiInterface defines the ai title generator controller's UI interface
type UiInterface interface {
	shared.UiInterface
	AiTitleGenerator(w http.ResponseWriter, r *http.Request)
}

// ui implements UiInterface
type ui struct {
	shared.UiBase
}

// UI creates a new ai title generator controller UI from the given config
func UI(config shared.UiConfig) UiInterface {
	return &ui{UiBase: shared.NewUiBase(config)}
}

// AiTitleGenerator handles the ai title generator controller requests
func (u *ui) AiTitleGenerator(w http.ResponseWriter, r *http.Request) {
	html := u.Handler(w, r)
	if html != "" {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(html))
	}
}

type pageData struct {
	Request             *http.Request
	Action              string
	ExistingPostRecords []blogai.RecordPost
	HasSystemPrompt     bool
}

// Handler processes the ai title generator controller request and returns HTML
func (u *ui) Handler(w http.ResponseWriter, r *http.Request) string {
	data, errorMessage := u.prepareData(r)

	if errorMessage != "" {
		return shared.ErrorPopup(errorMessage).ToHTML()
	}

	if r.Method == http.MethodGet && data.Action == ACTION_ADD_TITLE {
		return u.onAddTitleModal(r)
	}

	if r.Method == http.MethodPost {
		switch data.Action {
		case ACTION_ADD_TITLE:
			return u.onAddTitle(r)
		case ACTION_GENERATE_TITLES:
			return u.onGenerateTitles(r)
		case ACTION_APPROVE_TITLE:
			return u.onApproveTitle(r)
		case ACTION_REJECT_TITLE:
			return u.onRejectTitle(r)
		case ACTION_DELETE_TITLE:
			return u.onDeleteTitle(r)
		case ACTION_SETTINGS_FETCH:
			return u.handleSettingsFetchData(r)
		case ACTION_SETTINGS_SUBMIT:
			return u.handleSettingsSubmit(r)
		}
	}

	content := u.view(r, data)
	return u.Layout(w, r, "AI Title Generator", content.ToHTML(), struct {
		Styles     []string
		StyleURLs  []string
		Scripts    []string
		ScriptURLs []string
	}{
		ScriptURLs: []string{
			cdn.Htmx_2_0_0(),
			cdn.Sweetalert2_11(),
		},
		Styles: []string{
			htmx.HxHideIndicatorCSS(),
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
			Name: "Title Generator",
			URL:  linksHelper.AiTitleGenerator(nil),
		},
	})

	settingsModalHTML, _ := settingsModalFiles.ReadFile("settings_modal.html")
	settingsModalJS, _ := settingsModalFiles.ReadFile("settings_modal.js")

	vueCDN := hb.Script("").Src(cdn.VueJs_3_5_32())

	initScript := hb.Script(`
		const urlTitleSettingsFetchData = '` + linksHelper.AiTitleGenerator(map[string]string{"action": ACTION_SETTINGS_FETCH}) + `';
		const urlTitleSettingsSubmit = '` + linksHelper.AiTitleGenerator(map[string]string{"action": ACTION_SETTINGS_SUBMIT}) + `';
	`)

	settingsVueContainer := hb.Div().
		Child(vueCDN).
		Child(hb.Wrap().HTML(string(settingsModalHTML))).
		Child(initScript).
		Child(hb.Script(string(settingsModalJS)))

	settingsButton := hb.Div().Class("d-inline-block").Child(settingsVueContainer)

	card := hb.Div().
		Class("card shadow-sm w-100 mb-5")
	card = card.
		Child(
			hb.Div().Class("card-body text-center p-4").
				Child(hb.Div().
					Class("d-flex justify-content-between align-items-center mb-3").
					Child(hb.Heading1().
						HTML("Title Generator").
						Class("h3 mb-0 fw-bold text-dark")).
					Child(settingsButton),
				).
				Child(
					hb.Paragraph().
						HTML("Create up to 10 fresh AI titles per run—existing titles are skipped automatically.").
						Class("text-muted mb-4"),
				).
				Child(
					func() hb.TagInterface {
						if !data.HasSystemPrompt {
							return hb.Div().
								Class("col-8 mx-auto mb-4").
								Child(hb.Div().
									Class("alert alert-info d-flex align-items-center gap-2 mb-0").
									Attr("role", "alert").
									Child(hb.I().Class("bi bi-info-circle-fill")).
									Child(hb.Span().Text("Set the Title Generator settings first, then you can generate new titles.")))
						}

						return hb.Div().
							Class("d-grid gap-3 col-8 mx-auto mb-4").
							Children([]hb.TagInterface{
								hb.Button().
									Class("btn btn-primary btn-lg fw-semibold").
									HTML(`Generate New Titles <span class="htmx-indicator spinner-border spinner-border-sm" role="status"></span>`).
									HxPost(linksHelper.AiTitleGenerator(map[string]string{"action": ACTION_GENERATE_TITLES})).
									HxTarget("body").
									HxSwap("beforeend").
									Attr("hx-indicator", "this"),
								hb.Button().
									Class("btn btn-outline-primary btn-lg fw-semibold").
									HTML(`Add Custom Title <span class="htmx-indicator spinner-border spinner-border-sm" role="status"></span>`).
									HxGet(linksHelper.AiTitleGenerator(map[string]string{"action": ACTION_ADD_TITLE})).
									HxTarget("body").
									HxSwap("beforeend").
									Attr("hx-indicator", "this"),
							})
					}(),
				).
				Child(
					hb.Div().Class("text-start").
						Child(tableGeneratedTitles(r, data)),
				).
				Child(settingsButton),
		)

	return hb.Div().
		Class("container").
		Class("min-vh-100 py-4").
		Child(breadcrumbs).
		Child(card)
}

func (u *ui) prepareData(r *http.Request) (data pageData, errorMessage string) {
	data = pageData{
		Request: r,
		Action:  req.GetStringTrimmed(r, "action"),
	}

	if u.CustomStore() == nil {
		return data, "Custom store is not initialized"
	}

	records, err := u.CustomStore().RecordList(customstore.RecordQuery().
		SetType(blogai.POST_RECORD_TYPE))
	if err != nil {
		return data, fmt.Sprintf("Failed to fetch titles: %s", err.Error())
	}

	recordPosts := []blogai.RecordPost{}
	for _, record := range records {
		recordPost, err := blogai.NewRecordPostFromCustomRecord(record)
		if err != nil {
			u.Logger().Warn("Failed to parse custom record into RecordPost: " + err.Error())
			continue
		}
		recordPosts = append(recordPosts, recordPost)
	}

	data.ExistingPostRecords = recordPosts

	// Determine if the system prompt setting is configured
	if u.SettingStore() != nil {
		value, err := u.SettingStore().Get(r.Context(), SETTING_KEY_BLOG_TOPIC, "")
		if err == nil && strings.TrimSpace(value) != "" {
			data.HasSystemPrompt = true
		}
	}

	return data, ""
}
