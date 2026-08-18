package shared

import (
	"log/slog"
	"net/http"

	"github.com/dracory/blogstore"
	"github.com/dracory/hb"
	"github.com/spf13/cast"
)

// Header renders the blog admin navigation header.
// It is nil-safe for both store and logger.
func Header(store blogstore.StoreInterface, logger *slog.Logger, r *http.Request) hb.TagInterface {
	if store == nil {
		if logger != nil {
			logger.Error("blog store is nil")
		}
		return nil
	}

	linkHome := hb.NewHyperlink().
		HTML("Dashboard").
		Href(AdminHomeURL(r)).
		Class("nav-link")

	linkBlog := hb.NewHyperlink().
		HTML("Blog").
		Href(URLR(r, CONTROLLER_DASHBOARD, nil)).
		Class("nav-link")

	linkPosts := hb.Hyperlink().
		HTML("Posts ").
		Href(URLR(r, CONTROLLER_POST_MANAGER, nil)).
		Class("nav-link")

	linkCategories := hb.Hyperlink().
		HTML("Categories").
		Href(URLR(r, CONTROLLER_CATEGORY_MANAGER, nil)).
		Class("nav-link")

	linkTags := hb.Hyperlink().
		HTML("Tags").
		Href(URLR(r, CONTROLLER_TAG_MANAGER, nil)).
		Class("nav-link")

	linkAiTools := hb.Hyperlink().
		HTML("AI Tools").
		Href(URLR(r, CONTROLLER_AI_TOOLS, nil)).
		Class("nav-link")

	linkSettings := hb.Hyperlink().
		HTML("Settings").
		Href(URLR(r, CONTROLLER_BLOG_SETTINGS, nil)).
		Class("nav-link")

	postCount, err := store.PostCount(r.Context(), blogstore.PostQueryOptions{})
	if err != nil {
		if logger != nil {
			logger.Error(err.Error())
		}
		postCount = -1
	}

	ulNav := hb.NewUL().
		Class("nav  nav-pills justify-content-center").
		Child(hb.NewLI().
			Class("nav-item").Child(linkHome)).
		Child(hb.NewLI().
			Class("nav-item").Child(linkBlog)).
		Child(hb.LI().
			Class("nav-item").
			Child(linkPosts.
				Child(hb.Span().
					Class("badge bg-secondary ms-2").
					HTML(cast.ToString(postCount))))).
		Child(hb.LI().
			Class("nav-item").
			Child(linkCategories)).
		Child(hb.LI().
			Class("nav-item").
			Child(linkTags)).
		Child(hb.LI().
			Class("nav-item").
			Child(linkAiTools)).
		Child(hb.LI().
			Class("nav-item").
			Child(linkSettings))

	fileManagerURL := FileManagerURL(r)
	if fileManagerURL != "" {
		linkFileManager := hb.Hyperlink().
			HTML("File Manager").
			Href(fileManagerURL).
			Class("nav-link")
		ulNav.Child(hb.LI().
			Class("nav-item").
			Child(linkFileManager))
	}

	divCard := hb.NewDiv().Class("card card-default mt-3 mb-3")
	divCardBody := hb.NewDiv().Class("card-body").Style("padding: 2px;")
	return divCard.AddChild(divCardBody.AddChild(ulNav))
}
