package ai_post_content_update

import (
	"embed"
	"net/http"

	"github.com/dracory/blogadmin/shared"
)

//go:embed *.html
//go:embed *.js
var editorFiles embed.FS

const (
	actionFetchData       = "fetch-data"
	actionRegenerateBlock = "regenerate-block"
	actionSave            = "save"
)

// UiInterface defines the ai post content update controller's UI interface
type UiInterface interface {
	shared.UiInterface
	AiPostContentUpdate(w http.ResponseWriter, r *http.Request)
}

// ui implements UiInterface
type ui struct {
	shared.UiBase
}

// UI creates a new ai post content update controller UI from the given config
func UI(config shared.UiConfig) UiInterface {
	return &ui{UiBase: shared.NewUiBase(config)}
}

// AiPostContentUpdate handles the ai post content update controller requests
func (u *ui) AiPostContentUpdate(w http.ResponseWriter, r *http.Request) {
	html := u.Handler(w, r)
	if html != "" {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(html))
	}
}

// Handler processes the ai post content update controller request and returns HTML
func (u *ui) Handler(w http.ResponseWriter, r *http.Request) string {
	action := r.URL.Query().Get("action")
	if action == "" {
		action = r.PostFormValue("action")
	}

	switch action {
	case actionFetchData:
		return u.handleFetchData(r)
	case actionRegenerateBlock:
		return u.handleRegenerateBlock(r)
	case actionSave:
		return u.handleSave(w, r)
	default:
		return u.renderPage(w, r)
	}
}
