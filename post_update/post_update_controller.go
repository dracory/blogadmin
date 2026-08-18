package post_update

import (
	"embed"
	"net/http"

	"github.com/dracory/blogadmin/shared"
	"github.com/dracory/blogstore"
	"github.com/dracory/bs"
	"github.com/dracory/cdn"
	"github.com/dracory/hb"
	"github.com/dracory/req"
)

//go:embed *.html
//go:embed *.js
var postCategoriesFiles embed.FS

// UiInterface defines the post update controller's UI interface
type UiInterface interface {
	shared.UiInterface
	PostUpdate(w http.ResponseWriter, r *http.Request)
}

// ui implements UiInterface
type ui struct {
	shared.UiBase
	fileManagerURL string
}

// UI creates a new post update controller UI from the given config.
// fileManagerURL is the optional URL to the host project's file manager
// (used by the media component's "Add from file manager" flow).
func UI(config shared.UiConfig, fileManagerURL string) UiInterface {
	return &ui{
		UiBase:         shared.NewUiBase(config),
		fileManagerURL: fileManagerURL,
	}
}

// PostUpdate handles the post update controller requests
func (u *ui) PostUpdate(w http.ResponseWriter, r *http.Request) {
	html := u.Handler(w, r)
	if html != "" {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(html))
	}
}

// Handler processes the post update controller request and returns HTML
func (u *ui) Handler(w http.ResponseWriter, r *http.Request) string {
	action := req.GetStringTrimmed(r, "action")
	postID := req.GetStringTrimmed(r, "post_id")
	view := req.GetStringTrimmedOr(r, "view", "content")

	// Handle API actions for categories and tags
	if action != "" {
		switch action {
		case "load-categories":
			return u.handleLoadCategories(w, r)
		case "add-category":
			return u.handleAddCategory(w, r)
		case "remove-category":
			return u.handleRemoveCategory(w, r)
		case "load-tags":
			return u.handleLoadTags(w, r)
		case "add-tag":
			return u.handleAddTag(w, r)
		case "remove-tag":
			return u.handleRemoveTag(w, r)
		// Details component actions
		case "load-details":
			return u.handleLoadDetails(w, r)
		case "save-details":
			return u.handleSaveDetails(w, r)
		case "regenerate-image":
			return u.handleRegenerateImage(w, r)
		// Content component actions
		case "load-content":
			return u.handleLoadContent(w, r)
		case "save-content":
			return u.handleSaveContent(w, r)
		case "blockeditor-handle":
			return u.handleBlockEditorHandle(w, r)
		// SEO component actions
		case "load-seo":
			return u.handleLoadSEO(w, r)
		case "save-seo":
			return u.handleSaveSEO(w, r)
		// Versioning component actions
		case "load-versions":
			return u.handleLoadVersions(w, r)
		case "load-version-detail":
			return u.handleLoadVersionDetail(w, r)
		case "restore-version-attributes":
			return u.handleRestoreVersionAttributes(w, r)
		// Media component actions
		case "load-media":
			return u.newPostMediaComponent().HandleLoad(w, r)
		case "upload-media":
			return u.newPostMediaComponent().HandleUpload(w, r)
		case "save-media":
			return u.newPostMediaComponent().HandleSave(w, r)
		case "delete-media":
			return u.newPostMediaComponent().HandleDelete(w, r)
		case "add-media":
			return u.newPostMediaComponent().HandleAdd(w, r)
		}
	}

	if postID == "" {
		return shared.ErrorAlert("Post ID is required")
	}

	post, err := u.Store().PostFindByID(r.Context(), postID)
	if err != nil {
		u.Logger().Error(
			"Error. postUpdateController: PostFindByID",
			"error", err.Error(),
			"post_id", postID,
		)
		return shared.ErrorAlert("Post not found")
	}

	if post == nil {
		u.Logger().Warn(
			"Warning. postUpdateController: PostFindByID",
			"error", "Post not found",
			"post_id", postID,
		)
		return shared.ErrorAlert("Post not found")
	}

	pageContent := u.page(r, post, view)

	// BlockArea JS is a host-project resource; if not available, the
	// block editor tab still renders but BlockArea features are disabled.
	// We use a placeholder URL that the host can override via FuncLayout.
	blockAreaJS := "/js/blockarea_v0200.js"

	return u.Layout(w, r, "Edit Post | Blog", pageContent.ToHTML(), struct {
		Styles     []string
		StyleURLs  []string
		Scripts    []string
		ScriptURLs []string
	}{
		ScriptURLs: []string{
			cdn.Jquery_3_7_1(),
			"https://cdn.jsdelivr.net/npm/summernote@0.8.18/dist/summernote-lite.min.js",
			"https://cdn.jsdelivr.net/npm/easymde/dist/easymde.min.js",
			"https://cdn.jsdelivr.net/npm/codemirror@5.65.5/lib/codemirror.min.js",
			"https://cdn.jsdelivr.net/npm/codemirror@5.65.5/mode/markdown/markdown.min.js",
			"https://cdn.jsdelivr.net/npm/codemirror@5.65.5/mode/xml/xml.min.js",
			cdn.Sweetalert2_10(),
			cdn.JqueryUiJs_1_13_1(),
			blockAreaJS,
			cdn.VueJs_3_5_32(),
		},
		StyleURLs: []string{
			cdn.JqueryUiCss_1_13_1(),
			"https://cdn.jsdelivr.net/npm/summernote@0.8.18/dist/summernote-lite.min.css",
			"https://cdn.jsdelivr.net/npm/easymde/dist/easymde.min.css",
			"https://cdn.jsdelivr.net/npm/codemirror@5.65.5/lib/codemirror.min.css",
		},
	})
}

func (u *ui) page(r *http.Request, post blogstore.PostInterface, view string) hb.TagInterface {
	linksHelper := shared.NewLinksFromRequest(r)

	breadcrumbs := shared.Breadcrumbs([]shared.Breadcrumb{
		{
			Name: "Home",
			URL:  shared.AdminHomeURL(r),
		},
		{
			Name: "Blog",
			URL:  shared.URLR(r, shared.CONTROLLER_DASHBOARD, nil),
		},
		{
			Name: "Post Manager",
			URL:  linksHelper.PostManager(nil),
		},
		{
			Name: "Edit Post",
			URL:  linksHelper.PostUpdate(map[string]string{"post_id": post.GetID()}),
		},
	})

	buttonCancel := hb.Hyperlink().
		Class("btn btn-secondary ms-2 float-end").
		Child(hb.I().Class("bi bi-chevron-left").Style("margin-top:-4px;margin-right:8px;font-size:16px;")).
		HTML("Back").
		Href(shared.URLR(r, shared.CONTROLLER_DASHBOARD, nil))

	buttonView := hb.Hyperlink().
		Class("btn btn-info ms-2 float-end").
		Child(hb.I().Class("bi bi-eye").Style("margin-top:-4px;margin-right:8px;font-size:16px;")).
		HTML("View").
		Href("/blog/post/"+post.GetID()+"/"+post.GetSlug()).
		Attr("target", "_blank")

	buttonVersionHistory := hb.Button().
		Class("btn btn-primary ms-2 float-end").
		Child(hb.I().Class("bi bi-clock-history").Style("margin-top:-4px;margin-right:8px;font-size:16px;")).
		HTML("Version History").
		Attr("data-bs-toggle", "modal").
		Attr("data-bs-target", "#versionHistoryModal")

	heading := hb.Heading1().
		HTML("Edit Post").
		Child(buttonCancel).
		Child(buttonView).
		Child(buttonVersionHistory)

	tabs := bs.NavTabs().
		Class("mb-3").
		Child(bs.NavItem().
			Child(bs.NavLink().
				ClassIf(view == "details", "active").
				Href(linksHelper.PostUpdate(map[string]string{
					"post_id": post.GetID(),
					"view":    "details",
				})).
				HTML("Details"))).
		Child(bs.NavItem().
			Child(bs.NavLink().
				ClassIf(view == "content", "active").
				Href(linksHelper.PostUpdate(map[string]string{
					"post_id": post.GetID(),
					"view":    "content",
				})).
				HTML("Content"))).
		Child(bs.NavItem().
			Child(bs.NavLink().
				ClassIf(view == "categories", "active").
				Href(linksHelper.PostUpdate(map[string]string{
					"post_id": post.GetID(),
					"view":    "categories",
				})).
				HTML("Categories"))).
		Child(bs.NavItem().
			Child(bs.NavLink().
				ClassIf(view == "tags", "active").
				Href(linksHelper.PostUpdate(map[string]string{
					"post_id": post.GetID(),
					"view":    "tags",
				})).
				HTML("Tags"))).
		Child(bs.NavItem().
			Child(bs.NavLink().
				ClassIf(view == "seo", "active").
				Href(linksHelper.PostUpdate(map[string]string{
					"post_id": post.GetID(),
					"view":    "seo",
				})).
				HTML("SEO"))).
		Child(bs.NavItem().
			Child(bs.NavLink().
				ClassIf(view == "media", "active").
				Href(linksHelper.PostUpdate(map[string]string{
					"post_id": post.GetID(),
					"view":    "media",
				})).
				HTML("Media")))

	postTitle := hb.Heading2().
		Class("mb-3").
		HTML("Post: ").
		HTML(post.GetTitle())

	var body hb.TagInterface

	switch view {
	case "details":
		body = u.renderDetailsView(r, post)
	case "content":
		body = u.renderContentView(r, post)
	case "categories":
		body = u.renderCategoriesView(r, post)
	case "tags":
		body = u.renderTagsView(r, post)
	case "seo":
		body = u.renderSEOView(r, post)
	case "media":
		body = u.renderMediaView(r, post)
	default:
		body = hb.Div().Text("Not implemented yet")
	}

	card := hb.Div().
		Class("card").
		Child(
			hb.Div().
				Class("card-header").
				Child(hb.Heading4().
					HTMLIf(view == "details", "Post Details").
					HTMLIf(view == "content", "Post Contents").
					HTMLIf(view == "categories", "Post Categories").
					HTMLIf(view == "tags", "Post Tags").
					HTMLIf(view == "seo", "Post SEO").
					HTMLIf(view == "media", "Post Media").
					Style("margin-bottom:0;display:inline-block;")),
		).
		Child(
			hb.Div().
				Class("card-body").
				Child(body),
		)

	versioningModal := u.renderVersioningModal(r, post)

	return hb.Div().
		Class("container").
		Child(heading).
		Child(breadcrumbs).
		Child(postTitle).
		Child(tabs).
		Child(card).
		Child(versioningModal)
}

// Helper function to get content type from editor
func getContentTypeFromEditor(editor string) string {
	switch editor {
	case blogstore.POST_EDITOR_MARKDOWN:
		return blogstore.POST_CONTENT_TYPE_MARKDOWN
	case blogstore.POST_EDITOR_HTMLAREA:
		return blogstore.POST_CONTENT_TYPE_HTML
	case blogstore.POST_EDITOR_TEXTAREA:
		return blogstore.POST_CONTENT_TYPE_PLAIN_TEXT
	case blogstore.POST_EDITOR_BLOCKEDITOR, blogstore.POST_EDITOR_BLOCKAREA:
		return blogstore.POST_CONTENT_TYPE_BLOCKS
	default:
		return blogstore.POST_CONTENT_TYPE_PLAIN_TEXT
	}
}
