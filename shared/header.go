package shared

import (
	"log/slog"
	"net/http"

	"github.com/dracory/blogstore"
	"github.com/dracory/hb"
	"github.com/spf13/cast"
)

// Header renders the blog admin navigation header as pills inside a card,
// matching the nav-pills pattern used across the admin packages (shopadmin,
// taskadmin, useradmin). The active tab is derived from the ?controller=
// query parameter. The Posts tab shows the post count as a badge. The AI
// Tools tab is only shown when AI features are enabled, and the File Manager
// tab only when a file manager URL has been configured.
// It is nil-safe for both store and logger.
func Header(store blogstore.StoreInterface, logger *slog.Logger, r *http.Request) hb.TagInterface {
	if store == nil {
		if logger != nil {
			logger.Error("blog store is nil")
		}
		return nil
	}

	active := "blog"
	controller := ""
	if r != nil && r.URL != nil {
		controller = r.URL.Query().Get("controller")
	}
	switch controller {
	case CONTROLLER_POST_MANAGER, CONTROLLER_POST_CREATE, CONTROLLER_POST_UPDATE,
		CONTROLLER_POST_UPDATE_V1, CONTROLLER_POST_DELETE:
		active = "posts"
	case CONTROLLER_CATEGORY_MANAGER:
		active = "categories"
	case CONTROLLER_TAG_MANAGER:
		active = "tags"
	case CONTROLLER_BLOG_SETTINGS:
		active = "settings"
	case CONTROLLER_AI_TOOLS, CONTROLLER_AI_POST_CONTENT_UPDATE,
		CONTROLLER_AI_POST_GENERATOR, CONTROLLER_AI_TITLE_GENERATOR,
		CONTROLLER_AI_POST_EDITOR, CONTROLLER_AI_TEST:
		active = "ai"
	}

	postCount, err := store.PostCount(r.Context(), blogstore.PostQueryOptions{})
	if err != nil {
		if logger != nil {
			logger.Error(err.Error())
		}
		postCount = -1
	}

	tab := func(label, key, url string, count int) hb.TagInterface {
		class := "nav-link"
		if key == active {
			class += " active"
		}
		link := hb.Hyperlink().HTML(label).Href(url).Class(class)
		if count >= 0 {
			link.Child(hb.Span().
				Class("badge bg-secondary ms-2").
				HTML(cast.ToString(count)))
		}
		return hb.LI().Class("nav-item").Child(link)
	}

	tabs := []hb.TagInterface{
		tab("Dashboard", "home", AdminHomeURL(r), -1),
		tab("Blog", "blog", URLR(r, CONTROLLER_DASHBOARD, nil), -1),
		tab("Posts", "posts", URLR(r, CONTROLLER_POST_MANAGER, nil), int(postCount)),
		tab("Categories", "categories", URLR(r, CONTROLLER_CATEGORY_MANAGER, nil), -1),
		tab("Tags", "tags", URLR(r, CONTROLLER_TAG_MANAGER, nil), -1),
	}

	// Only show the AI Tools nav link when AI features are enabled.
	if AIEnabled(r) {
		tabs = append(tabs, tab("AI Tools", "ai", URLR(r, CONTROLLER_AI_TOOLS, nil), -1))
	}

	tabs = append(tabs, tab("Settings", "settings", URLR(r, CONTROLLER_BLOG_SETTINGS, nil), -1))

	if fileManagerURL := FileManagerURL(r); fileManagerURL != "" {
		tabs = append(tabs, tab("File Manager", "files", fileManagerURL, -1))
	}

	ulNav := hb.UL().
		Class("nav nav-pills p-1 justify-content-center flex-grow-1").
		Children(tabs)

	label := hb.Span().
		Class("fw-bolder text-uppercase ms-3 me-3").
		Text("Blog").
		Child(hb.I().Class("bi bi-chevron-right ms-2"))

	return hb.Div().
		Class("card card-default mt-3 mb-4 bg-primary-subtle").
		Child(hb.Div().
			Class("card-body d-flex align-items-center").
			Style("padding: 2px;").
			Child(label).
			Child(ulNav))
}
