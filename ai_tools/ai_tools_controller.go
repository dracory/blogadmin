package ai_tools

import (
	"net/http"

	"github.com/dracory/blogadmin/shared"
	"github.com/dracory/cdn"
	"github.com/dracory/hb"
)

// UiInterface defines the ai tools controller's UI interface
type UiInterface interface {
	shared.UiInterface
	AiTools(w http.ResponseWriter, r *http.Request)
}

// ui implements UiInterface
type ui struct {
	shared.UiBase
}

// UI creates a new ai tools controller UI from the given config
func UI(config shared.UiConfig) UiInterface {
	return &ui{UiBase: shared.NewUiBase(config)}
}

// AiTools handles the ai tools controller requests
func (u *ui) AiTools(w http.ResponseWriter, r *http.Request) {
	html := u.Handler(w, r)
	if html != "" {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(html))
	}
}

// Handler processes the ai tools controller request and returns HTML
func (u *ui) Handler(w http.ResponseWriter, r *http.Request) string {
	if r.Method == http.MethodPost && r.FormValue("action") == "testai" {
		w.Header().Set("Content-Type", "application/json")
		model, err := u.LlmEngine()
		if err != nil {
			if _, writeErr := w.Write([]byte(hb.Swal(hb.SwalOptions{
				Title:            "Error",
				Text:             err.Error(),
				Icon:             "error",
				Timer:            15000,
				TimerProgressBar: true,
			}).ToHTML())); writeErr != nil {
				return ""
			}
			return ""
		}
		if model == nil {
			if _, writeErr := w.Write([]byte(hb.Swal(hb.SwalOptions{
				Title:            "Error",
				Text:             "model is nil",
				Icon:             "error",
				Timer:            15000,
				TimerProgressBar: true,
			}).ToHTML())); writeErr != nil {
				return ""
			}
			return ""
		}

		response, err := model.GenerateText("You are a helpful assistant.", "Tell me shortly about blogs.")
		if err != nil {
			if _, writeErr := w.Write([]byte(hb.Swal(hb.SwalOptions{
				Title:            "Error",
				Text:             err.Error(),
				Icon:             "error",
				Timer:            15000,
				TimerProgressBar: true,
			}).ToHTML())); writeErr != nil {
				return ""
			}
			return ""
		}

		w.WriteHeader(http.StatusOK)
		if _, writeErr := w.Write([]byte(hb.Swal(hb.SwalOptions{
			Title:            "Success",
			Text:             response,
			Icon:             "success",
			Timer:            15000,
			TimerProgressBar: true,
		}).ToHTML())); writeErr != nil {
			return ""
		}
		return ""
	}

	content := u.view(r)
	return u.Layout(w, r, "BlogAI", content.ToHTML(), struct {
		Styles     []string
		StyleURLs  []string
		Scripts    []string
		ScriptURLs []string
	}{
		ScriptURLs: []string{cdn.Sweetalert2_11()},
	})
}

func (u *ui) view(r *http.Request) hb.TagInterface {
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
	})

	card := hb.Div().
		Class("card shadow-sm w-100 mb-5")
	card = card.Child(
		hb.Div().Class("card-body text-center p-4").
			Child(
				hb.Heading1().
					HTML("BlogAI").
					Class("h3 mb-3 fw-bold"),
			).
			Child(
				hb.Paragraph().
					HTML("Welcome to your AI-powered blog tools.").
					Class("text-muted mb-4"),
			).
			Child(
				hb.Div().Class("d-grid gap-3 col-8 mx-auto").
					Child(
						hb.Hyperlink().
							HTML("Post Generator").
							Href(linksHelper.AiPostGenerator(nil)).
							Class("btn btn-primary btn-lg fw-semibold"),
					).
					Child(
						hb.Hyperlink().
							HTML("Title Generator").
							Href(linksHelper.AiTitleGenerator(nil)).
							Class("btn btn-success btn-lg fw-semibold"),
					).
					Child(
						hb.Hyperlink().
							Class("btn btn-warning btn-lg fw-semibold").
							HTML("Test AI is working").
							Href(linksHelper.AiTest(nil)),
					).
					Child(
						hb.Hyperlink().
							Class("btn btn-outline-secondary btn-lg fw-semibold d-inline-flex align-items-center justify-content-center").
							Child(hb.I().Class("bi bi-arrow-left-circle me-2")).
							HTML("Back to Blog Home").
							Href(linksHelper.Home(nil)),
					),
			),
	)

	return hb.Div().
		Class("container").
		Class("min-vh-100 py-4").
		Child(breadcrumbs).
		Child(card)
}
