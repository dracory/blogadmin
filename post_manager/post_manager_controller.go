package post_manager

import (
	"embed"
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/dracory/api"
	"github.com/dracory/blogadmin/shared"
	"github.com/dracory/blogstore"
	"github.com/dracory/cdn"
	"github.com/dracory/hb"
	"github.com/dracory/neat"
	"github.com/dracory/req"
	"github.com/dromara/carbon/v2"
	"github.com/spf13/cast"
)

//go:embed *.html
//go:embed *.js
var postsFiles embed.FS

const (
	actionLoadPosts  = "load-posts"
	actionDeletePost = "delete-post"
	actionCreatePost = "create-post"
)

// UiInterface defines the post manager controller's UI interface
type UiInterface interface {
	shared.UiInterface
	PostManager(w http.ResponseWriter, r *http.Request)
}

// ui implements UiInterface
type ui struct {
	shared.UiBase
}

// UI creates a new post manager controller UI from the given config
func UI(config shared.UiConfig) UiInterface {
	return &ui{UiBase: shared.NewUiBase(config)}
}

// PostManager handles the post manager controller requests
func (u *ui) PostManager(w http.ResponseWriter, r *http.Request) {
	html := u.Handler(w, r)
	if html != "" {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(html))
	}
}

// Handler processes the post manager controller request and returns HTML
func (u *ui) Handler(w http.ResponseWriter, r *http.Request) string {
	action := req.GetStringTrimmed(r, "action")

	switch action {
	case actionLoadPosts:
		return u.handleLoadPosts(w, r)
	case actionCreatePost:
		return u.handleCreatePost(w, r)
	case actionDeletePost:
		return u.handleDeletePost(w, r)
	default:
		return u.renderPage(w, r)
	}
}

func (u *ui) renderPage(w http.ResponseWriter, r *http.Request) string {
	breadcrumbs := shared.Breadcrumbs([]shared.Breadcrumb{
		{Name: "Home", URL: shared.AdminHomeURL(r)},
		{Name: "Blog", URL: shared.URLR(r, shared.CONTROLLER_DASHBOARD, nil)},
		{Name: "Post Manager", URL: ""},
	})

	linksHelper := shared.NewLinksFromRequest(r)

	actionButtons := hb.Div().
		Class("d-flex gap-2 float-end")

	buttonAiHome := hb.Hyperlink().
		Class("btn btn-light text-dark d-inline-flex align-items-center").
		Child(hb.I().Class("bi bi-stars me-2")).
		HTML("AI Tools").
		Href(linksHelper.AiTools(nil))

	buttonSettings := hb.Hyperlink().
		Class("btn btn-outline-secondary d-inline-flex align-items-center").
		Child(hb.I().Class("bi bi-gear me-2")).
		HTML("Settings").
		Href(linksHelper.BlogSettings(nil))

	actionButtons = actionButtons.
		Child(buttonAiHome).
		Child(buttonSettings)

	heading := hb.Heading1().HTML("Blog. Post Manager").Child(actionButtons)

	htmlContent, err := postsFiles.ReadFile("posts.html")
	if err != nil {
		u.Logger().Error("Failed to read posts HTML template", "error", err)
		return hb.Div().HTML("Error loading posts component").ToHTML()
	}

	jsContent, err := postsFiles.ReadFile("posts.js")
	if err != nil {
		u.Logger().Error("Failed to read posts JavaScript file", "error", err)
		return hb.Div().HTML("Error loading posts component").ToHTML()
	}

	vueCDN := hb.Script("").Src(cdn.VueJs_3_5_32())

	initScript := hb.Script(`
		const urlPostsLoad = '` + linksHelper.PostManager(map[string]string{"action": actionLoadPosts}) + `';
		const urlPostDelete = '` + linksHelper.PostManager(map[string]string{"action": actionDeletePost}) + `';
		const urlPostCreate = '` + linksHelper.PostManager(map[string]string{"action": actionCreatePost}) + `';
		const urlAiPostContentUpdate = '` + linksHelper.AiPostContentUpdate(map[string]string{"post_id": "POST_ID_PLACEHOLDER"}) + `';
		const urlPostUpdate = '` + linksHelper.PostUpdate(map[string]string{"post_id": "POST_ID_PLACEHOLDER"}) + `';
	`)

	htmlTemplate := hb.Wrap().HTML(string(htmlContent))
	componentScript := hb.Script(string(jsContent))

	vueContainer := hb.Div().
		Child(vueCDN).
		Child(htmlTemplate).
		Child(initScript).
		Child(componentScript)

	content := hb.Div().
		Class("container").
		Child(heading).
		Child(breadcrumbs).
		Child(hb.HR()).
		Child(vueContainer)

	return u.Layout(w, r, "Blog | Post Manager", content.ToHTML(), struct {
		Styles     []string
		StyleURLs  []string
		Scripts    []string
		ScriptURLs []string
	}{
		ScriptURLs: []string{
			cdn.Sweetalert2_10(),
		},
	})
}

func (u *ui) handleLoadPosts(w http.ResponseWriter, r *http.Request) string {
	ctx := r.Context()

	blogStore := u.Store()
	if blogStore == nil {
		return api.Error("Blog store not available").ToString()
	}

	// Read parameters from query string for GET requests
	page := cast.ToInt(req.GetStringTrimmedOr(r, "page", "0"))
	perPage := cast.ToInt(req.GetStringTrimmedOr(r, "per_page", "10"))
	sortOrder := req.GetStringTrimmedOr(r, "sort_order", neat.SortDesc)
	sortBy := req.GetStringTrimmedOr(r, "sort_by", blogstore.COLUMN_CREATED_AT)
	status := req.GetStringTrimmed(r, "status")
	search := req.GetStringTrimmed(r, "search")
	dateFrom := req.GetStringTrimmedOr(r, "date_from", carbon.Now().AddYears(-1).ToDateString())
	dateTo := req.GetStringTrimmedOr(r, "date_to", carbon.Now().ToDateString())

	query := blogstore.PostQueryOptions{
		Search:               search,
		Offset:               page * perPage,
		Limit:                perPage,
		Status:               status,
		CreatedAtGreaterThan: dateFrom + " 00:00:00",
		CreatedAtLessThan:    dateTo + " 23:59:59",
		SortOrder:            sortOrder,
		OrderBy:              sortBy,
	}

	posts, err := blogStore.PostList(ctx, query)
	if err != nil {
		u.Logger().Error("Failed to load posts", "error", err)
		return api.Error("Failed to load posts").ToString()
	}

	postList := []map[string]any{}
	for _, post := range posts {
		postList = append(postList, map[string]any{
			"id":           post.GetID(),
			"title":        post.GetTitle(),
			"status":       post.GetStatus(),
			"featured":     post.GetFeatured(),
			"published_at": post.GetPublishedAt(),
			"created_at":   post.GetCreatedAt(),
			"updated_at":   post.GetUpdatedAt(),
			"slug":         post.GetSlug(),
			"image_url":    shared.PostImageURL(ctx, blogStore, post),
		})
	}

	count, err := blogStore.PostCount(ctx, query)
	if err != nil {
		u.Logger().Error("Failed to get posts count", "error", err)
		return api.Error("Failed to get posts count").ToString()
	}

	return api.SuccessWithData("Posts loaded successfully", map[string]any{
		"posts": postList,
		"total": count,
	}).ToString()
}

func (u *ui) handleDeletePost(w http.ResponseWriter, r *http.Request) string {
	ctx := r.Context()

	if r.Method != http.MethodPost {
		return api.Error("Method not allowed").ToString()
	}

	var reqData struct {
		PostID string `json:"post_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&reqData); err != nil {
		return api.Error("Invalid request body").ToString()
	}

	if reqData.PostID == "" {
		return api.Error("Post ID is required").ToString()
	}

	blogStore := u.Store()
	if blogStore == nil {
		return api.Error("Blog store not available").ToString()
	}

	post, err := blogStore.PostFindByID(ctx, reqData.PostID)
	if err != nil {
		u.Logger().Error("Failed to find post for delete", "error", err)
		return api.Error("Post not found").ToString()
	}

	if err := blogStore.PostDelete(ctx, post); err != nil {
		u.Logger().Error("Failed to delete post", "error", err)
		return api.Error("Failed to delete post").ToString()
	}

	return api.SuccessWithData("Post deleted successfully", map[string]any{}).ToString()
}

func (u *ui) handleCreatePost(w http.ResponseWriter, r *http.Request) string {
	ctx := r.Context()

	if r.Method != http.MethodPost {
		return api.Error("Method not allowed").ToString()
	}

	var reqData struct {
		Title string `json:"title"`
	}

	if err := json.NewDecoder(r.Body).Decode(&reqData); err != nil {
		return api.Error("Invalid request body").ToString()
	}

	if reqData.Title == "" {
		return api.Error("Title is required").ToString()
	}

	blogStore := u.Store()
	if blogStore == nil {
		return api.Error("Blog store not available").ToString()
	}

	post := blogstore.NewPost()
	post.SetTitle(reqData.Title)

	if err := blogStore.PostCreate(ctx, post); err != nil {
		u.Logger().Error("Failed to create post", "error", err)
		return api.Error("Failed to create post").ToString()
	}

	return api.SuccessWithData("Post created successfully", map[string]any{
		"id": post.GetID(),
	}).ToString()
}

// Ensure slog is used (referenced in error paths above)
var _ = slog.Error
