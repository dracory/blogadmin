package blog_settings

import (
	"embed"
	"net/http"

	"github.com/dracory/blogadmin/shared"
)

//go:embed *.html
//go:embed *.js
var settingsFiles embed.FS

const (
	actionFetchData = "fetch-data"
	actionSubmit    = "submit"
)

// UiInterface defines the blog settings controller's UI interface
type UiInterface interface {
	shared.UiInterface
	BlogSettings(w http.ResponseWriter, r *http.Request)
}

// ui implements UiInterface
type ui struct {
	shared.UiBase
}

// UI creates a new blog settings controller UI from the given config
func UI(config shared.UiConfig) UiInterface {
	return &ui{UiBase: shared.NewUiBase(config)}
}

// BlogSettings handles the blog settings controller requests
func (u *ui) BlogSettings(w http.ResponseWriter, r *http.Request) {
	html := u.Handler(w, r)
	if html != "" {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(html))
	}
}

// Handler processes the blog settings controller request and returns HTML
func (u *ui) Handler(w http.ResponseWriter, r *http.Request) string {
	action := r.URL.Query().Get("action")
	if action == "" {
		action = r.PostFormValue("action")
	}

	switch action {
	case actionFetchData:
		return u.handleFetchData(r)
	case actionSubmit:
		return u.handleSubmit(r)
	default:
		return u.renderPage(w, r)
	}
}
