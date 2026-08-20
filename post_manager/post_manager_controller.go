package post_manager

import (
	"embed"
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"

	"github.com/dracory/api"
	"github.com/dracory/blogadmin/shared"
	"github.com/dracory/blogstore"
	"github.com/dracory/cdn"
	"github.com/dracory/hb"
	"github.com/dracory/neat"
	"github.com/dracory/req"
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

// Condition field constants for the multi-filter system.
const (
	CondFieldSearch   = "search"
	CondFieldStatus   = "status"
	CondFieldSlug     = "slug"
	CondFieldDateFrom = "date_from"
	CondFieldDateTo   = "date_to"
	CondFieldFeatured = "featured"
)

// condition is a single filter condition sent from the frontend.
type condition struct {
	Field    string `json:"field"`
	Operator string `json:"operator"`
	Value    string `json:"value"`
}

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

	buttonSettings := hb.Hyperlink().
		Class("btn btn-outline-secondary d-inline-flex align-items-center").
		Child(hb.I().Class("bi bi-gear me-2")).
		HTML("Settings").
		Href(linksHelper.BlogSettings(nil))

	actionButtons = actionButtons.Child(buttonSettings)

	// Only show the AI Tools button when AI features are enabled.
	if u.AIEnabled() {
		buttonAiHome := hb.Hyperlink().
			Class("btn btn-light text-dark d-inline-flex align-items-center").
			Child(hb.I().Class("bi bi-stars me-2")).
			HTML("AI Tools").
			Href(linksHelper.AiTools(nil))
		// Prepend so AI Tools sits to the left of Settings.
		actionButtons = hb.Div().
			Class("d-flex gap-2 float-end").
			Child(buttonAiHome).
			Child(buttonSettings)
	}

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

	// aiEnabledJS exposes the AI-enabled flag to the Vue app so it can
	// hide the per-row AI Content Editor button when AI is disabled.
	aiEnabledJS := "false"
	if u.AIEnabled() {
		aiEnabledJS = "true"
	}

	initScript := hb.Script(`
		const aiEnabled = ` + aiEnabledJS + `;
		const urlPostsLoad = '` + linksHelper.PostManager(map[string]string{"action": actionLoadPosts}) + `';
		const urlPostDelete = '` + linksHelper.PostManager(map[string]string{"action": actionDeletePost}) + `';
		const urlPostCreate = '` + linksHelper.PostManager(map[string]string{"action": actionCreatePost}) + `';
		const urlAiPostContentUpdate = '` + linksHelper.AiPostContentUpdate(map[string]string{"post_id": "POST_ID_PLACEHOLDER"}) + `';
		const urlPostUpdate = '` + linksHelper.PostUpdate(map[string]string{"post_id": "POST_ID_PLACEHOLDER"}) + `';
	`)

	htmlTemplate := hb.Wrap().HTML(string(htmlContent))
	componentScript := hb.Script(string(jsContent))

	vueContainer := hb.Div().
		Child(shared.VueLoaderScript()).
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

	// Accept both POST (JSON body with conditions) and GET (query params
	// for backward compatibility). POST is the primary path used by the
	// Vue frontend; GET keeps shareable URLs working.
	var reqBody struct {
		Page       int         `json:"page"`
		PerPage    int         `json:"per_page"`
		SortBy     string      `json:"sort_by"`
		SortOrder  string      `json:"sort_order"`
		Conditions []condition `json:"conditions"`
	}

	if r.Method == http.MethodPost {
		if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
			return api.Error("Invalid request body").ToString()
		}
	} else {
		// Fallback: read from query params (legacy / shareable URLs)
		reqBody.Page = cast.ToInt(req.GetStringTrimmedOr(r, "page", "0"))
		reqBody.PerPage = cast.ToInt(req.GetStringTrimmedOr(r, "per_page", "10"))
		reqBody.SortBy = req.GetStringTrimmedOr(r, "sort_by", blogstore.COLUMN_CREATED_AT)
		reqBody.SortOrder = req.GetStringTrimmedOr(r, "sort_order", neat.SortDesc)

		// Build conditions from individual query params
		if v := req.GetStringTrimmed(r, "search"); v != "" {
			reqBody.Conditions = append(reqBody.Conditions, condition{CondFieldSearch, "contains", v})
		}
		if v := req.GetStringTrimmed(r, "status"); v != "" {
			reqBody.Conditions = append(reqBody.Conditions, condition{CondFieldStatus, "equals", v})
		}
		if v := req.GetStringTrimmed(r, "slug"); v != "" {
			reqBody.Conditions = append(reqBody.Conditions, condition{CondFieldSlug, "equals", v})
		}
		if v := req.GetStringTrimmed(r, "date_from"); v != "" {
			reqBody.Conditions = append(reqBody.Conditions, condition{CondFieldDateFrom, "equals", v})
		}
		if v := req.GetStringTrimmed(r, "date_to"); v != "" {
			reqBody.Conditions = append(reqBody.Conditions, condition{CondFieldDateTo, "equals", v})
		}
		if v := req.GetStringTrimmed(r, "featured"); v != "" {
			reqBody.Conditions = append(reqBody.Conditions, condition{CondFieldFeatured, "equals", v})
		}
	}

	if reqBody.Page < 0 {
		reqBody.Page = 0
	}
	if reqBody.PerPage <= 0 || reqBody.PerPage > 500 {
		reqBody.PerPage = 10
	}
	if reqBody.SortBy == "" {
		reqBody.SortBy = blogstore.COLUMN_CREATED_AT
	}
	if reqBody.SortOrder == "" {
		reqBody.SortOrder = neat.SortDesc
	}

	// Build query from conditions
	query := blogstore.PostQueryOptions{
		Offset:    reqBody.Page * reqBody.PerPage,
		Limit:     reqBody.PerPage,
		OrderBy:   reqBody.SortBy,
		SortOrder: reqBody.SortOrder,
	}

	applyConditions(&query, reqBody.Conditions)

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

	totalPages := int(count) / reqBody.PerPage
	if int(count)%reqBody.PerPage != 0 {
		totalPages++
	}

	return api.SuccessWithData("Posts loaded successfully", map[string]any{
		"posts":       postList,
		"total":       count,
		"page":        reqBody.Page,
		"per_page":    reqBody.PerPage,
		"total_pages": totalPages,
	}).ToString()
}

// applyConditions maps filter conditions to PostQueryOptions fields.
// Multiple conditions on the same field: last one wins for single-value
// fields (status, slug, search). Date conditions accumulate.
func applyConditions(query *blogstore.PostQueryOptions, conds []condition) {
	for _, c := range conds {
		val := strings.TrimSpace(c.Value)
		if val == "" {
			continue
		}
		switch c.Field {
		case CondFieldSearch:
			query.Search = val
		case CondFieldStatus:
			query.Status = val
		case CondFieldSlug:
			query.Slug = val
		case CondFieldDateFrom:
			query.CreatedAtGreaterThan = val + " 00:00:00"
		case CondFieldDateTo:
			query.CreatedAtLessThan = val + " 23:59:59"
		case CondFieldFeatured:
			// Featured is stored in meta; we can't filter at the store
			// level easily, so this is a no-op for now. The frontend
			// could filter in-memory if needed.
		}
	}
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
