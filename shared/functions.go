package shared

import (
	"context"
	"errors"
	"sort"

	"github.com/dracory/blogstore"
	"github.com/dracory/hb"
)

// errLlmFactoryNil is returned when LlmEngine() is called but no
// LlmFactory was provided in UiConfig.
var errLlmFactoryNil = errors.New("llm factory is not configured")

// ErrorPopup returns a SweetAlert2 error popup tag.
func ErrorPopup(errorMessage string) hb.TagInterface {
	return hb.Swal(hb.SwalOptions{
		Title:            "Error",
		Text:             errorMessage,
		Icon:             "error",
		Timer:            10000,
		TimerProgressBar: true,
	})
}

// SuccessPopup returns a SweetAlert2 success popup tag.
func SuccessPopup(successMessage string) hb.TagInterface {
	return hb.Swal(hb.SwalOptions{
		Title:            "Success",
		Text:             successMessage,
		Icon:             "success",
		Timer:            10000,
		TimerProgressBar: true,
	})
}

// SuccessPopupWithRedirect returns a SweetAlert2 success popup with an
// optional redirect. If redirectUrl is empty, no redirect is configured.
func SuccessPopupWithRedirect(successMessage string, redirectUrl string, redirectSeconds int) hb.TagInterface {
	if redirectUrl != "" {
		return hb.Swal(hb.SwalOptions{
			Title:            "Success",
			Text:             successMessage,
			Icon:             "success",
			Timer:            redirectSeconds * 1000,
			TimerProgressBar: true,
			RedirectURL:      redirectUrl,
			RedirectSeconds:  redirectSeconds,
		})
	}

	return hb.Swal(hb.SwalOptions{
		Title:            "Success",
		Text:             successMessage,
		Icon:             "success",
		Timer:            redirectSeconds * 1000,
		TimerProgressBar: true,
	})
}

// PostImageURL resolves the image URL for a post.
// It returns the first media URL (sorted by sequence) if media exists,
// using the /blog/media/<id>.<ext> route. Otherwise falls back to the
// post's ImageUrl field.
//
// This replaces the former project/internal/rules.PostImageURL helper,
// keeping blogadmin decoupled from the host project.
func PostImageURL(ctx context.Context, store blogstore.StoreInterface, post blogstore.PostInterface) string {
	if store == nil || post == nil {
		return ""
	}

	media, err := store.MediaListByEntityID(ctx, post.GetID())
	if err == nil && len(media) > 0 {
		sort.Slice(media, func(i, j int) bool {
			return media[i].GetSequence() < media[j].GetSequence()
		})
		m := media[0]
		ext := m.GetExtension()
		if ext == "" {
			ext = ".png"
		}
		return "/blog/media/" + m.GetID() + ext
	}

	return post.GetImageUrl()
}
