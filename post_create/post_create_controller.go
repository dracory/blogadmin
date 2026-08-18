package post_create

import (
	"log/slog"
	"net/http"

	"github.com/dracory/blogadmin/shared"
	"github.com/dracory/blogstore"
	"github.com/dracory/hb"
	"github.com/dracory/req"
)

// UiInterface defines the post create controller's UI interface
type UiInterface interface {
	shared.UiInterface
	PostCreate(w http.ResponseWriter, r *http.Request)
}

// ui implements UiInterface
type ui struct {
	shared.UiBase
}

// UI creates a new post create controller UI from the given config
func UI(config shared.UiConfig) UiInterface {
	return &ui{UiBase: shared.NewUiBase(config)}
}

// PostCreate handles the post create controller requests
func (u *ui) PostCreate(w http.ResponseWriter, r *http.Request) {
	html := u.Handler(w, r)
	if html != "" {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(html))
	}
}

type postCreateControllerData struct {
	title          string
	successMessage string
}

// Handler processes the post create controller request and returns HTML
func (u *ui) Handler(w http.ResponseWriter, r *http.Request) string {
	data, errorMessage := u.prepareDataAndValidate(r)

	if errorMessage != "" {
		return hb.Swal(hb.SwalOptions{
			Icon: "error",
			Text: errorMessage,
		}).ToHTML()
	}

	if data.successMessage != "" {
		return hb.Wrap().
			Child(hb.Swal(hb.SwalOptions{
				Icon: "success",
				Text: data.successMessage,
			})).
			Child(hb.Script("setTimeout(() => {window.location.href = window.location.href}, 2000)")).
			ToHTML()
	}

	return modalPostCreate(r, data).ToHTML()
}

func (u *ui) prepareDataAndValidate(r *http.Request) (data postCreateControllerData, errorMessage string) {
	data.title = req.GetStringTrimmed(r, "post_title")

	if r.Method != "POST" {
		return data, ""
	}

	if data.title == "" {
		return data, "post title is required"
	}

	post := blogstore.NewPost()
	post.SetTitle(data.title)

	err := u.Store().PostCreate(r.Context(), post)

	if err != nil {
		u.Logger().Error("At postCreateController > prepareDataAndValidate", slog.String("error", err.Error()))
		return data, "Creating post failed. Please contact an administrator."
	}

	data.successMessage = "post created successfully."

	return data, ""
}
