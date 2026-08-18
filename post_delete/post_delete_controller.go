package post_delete

import (
	"log/slog"
	"net/http"

	"github.com/dracory/blogadmin/shared"
	"github.com/dracory/blogstore"
	"github.com/dracory/hb"
	"github.com/dracory/req"
)

// UiInterface defines the post delete controller's UI interface
type UiInterface interface {
	shared.UiInterface
	PostDelete(w http.ResponseWriter, r *http.Request)
}

// ui implements UiInterface
type ui struct {
	shared.UiBase
}

// UI creates a new post delete controller UI from the given config
func UI(config shared.UiConfig) UiInterface {
	return &ui{UiBase: shared.NewUiBase(config)}
}

// PostDelete handles the post delete controller requests
func (u *ui) PostDelete(w http.ResponseWriter, r *http.Request) {
	html := u.Handler(w, r)
	if html != "" {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(html))
	}
}

type postDeleteControllerData struct {
	postID         string
	post           blogstore.PostInterface
	successMessage string
}

// Handler processes the post delete controller request and returns HTML
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

	return modalPostDelete(r, data).ToHTML()
}

func (u *ui) prepareDataAndValidate(r *http.Request) (data postDeleteControllerData, errorMessage string) {
	data.postID = req.GetStringTrimmed(r, "post_id")

	if data.postID == "" {
		return data, "post id is required"
	}

	post, err := u.Store().PostFindByID(r.Context(), data.postID)

	if err != nil {
		u.Logger().Error("At postDeleteController > prepareDataAndValidate", slog.String("error", err.Error()))
		return data, "Post not found"
	}

	if post == nil {
		return data, "Post not found"
	}

	data.post = post

	if r.Method != "POST" {
		return data, ""
	}

	err = u.Store().PostTrash(r.Context(), post)

	if err != nil {
		u.Logger().Error("At postDeleteController > prepareDataAndValidate", slog.String("error", err.Error()))
		return data, "Deleting post failed. Please contact an administrator."
	}

	data.successMessage = "post deleted successfully."

	return data, ""
}
