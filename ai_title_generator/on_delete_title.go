package ai_title_generator

import (
	"fmt"
	"net/http"

	"github.com/dracory/blogadmin/shared"
	"github.com/dracory/hb"
	"github.com/dracory/req"
)

func (u *ui) onDeleteTitle(r *http.Request) string {
	titleID := req.GetStringTrimmed(r, "record_post_id")
	if titleID == "" {
		return shared.ErrorPopup("Title ID is required").ToHTML()
	}

	err := u.CustomStore().RecordDeleteByID(titleID)
	if err != nil {
		return shared.ErrorPopup(fmt.Sprintf("Error deleting title: %s", err.Error())).ToHTML()
	}

	return hb.Swal(hb.SwalOptions{
		Title:            "Success",
		Text:             "Title deleted successfully! Reloading page...",
		Icon:             "success",
		Timer:            3000,
		TimerProgressBar: true,
		RedirectURL:      shared.NewLinksFromRequest(r).AiTitleGenerator(nil),
		RedirectSeconds:  3,
	}).ToHTML()
}
