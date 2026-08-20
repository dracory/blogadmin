package tag_manager

import (
	"context"
	"embed"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/dracory/api"
	"github.com/dracory/blogadmin/shared"
	"github.com/dracory/blogstore"
	"github.com/dracory/cdn"
	"github.com/dracory/hb"
	"github.com/dracory/str"
	"github.com/dracory/uid"
)

//go:embed *.html
//go:embed *.js
var tagsFiles embed.FS

// UiInterface defines the tag manager controller's UI interface
type UiInterface interface {
	shared.UiInterface
	TagManager(w http.ResponseWriter, r *http.Request)
}

// ui implements UiInterface
type ui struct {
	shared.UiBase
}

// UI creates a new tag manager controller UI from the given config
func UI(config shared.UiConfig) UiInterface {
	return &ui{UiBase: shared.NewUiBase(config)}
}

// TagManager handles the tag manager controller requests
func (u *ui) TagManager(w http.ResponseWriter, r *http.Request) {
	html := u.Handler(w, r)
	if html != "" {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(html))
	}
}

// Handler processes the tag manager controller request and returns HTML
func (u *ui) Handler(w http.ResponseWriter, r *http.Request) string {
	action := r.URL.Query().Get("action")

	switch action {
	case "load-tags":
		return u.handleLoadTags(r)
	case "load-tag-posts":
		return u.handleLoadTagPosts(r)
	case "create-tag":
		return u.handleCreateTag(w, r)
	case "update-tag":
		return u.handleUpdateTag(w, r)
	case "delete-tag":
		return u.handleDeleteTag(w, r)
	default:
		return u.renderPage(r)
	}
}

func (u *ui) renderPage(r *http.Request) string {
	breadcrumbs := shared.Breadcrumbs([]shared.Breadcrumb{
		{Name: "Home", URL: shared.AdminHomeURL(r)},
		{Name: "Blog", URL: shared.URLR(r, shared.CONTROLLER_DASHBOARD, nil)},
		{Name: "Tags", URL: ""},
	})

	heading := hb.Heading1().HTML("Blog. Tag Manager")

	htmlContent, err := tagsFiles.ReadFile("tags.html")
	if err != nil {
		u.Logger().Error("Failed to read tags HTML template", "error", err)
		return hb.Div().HTML("Error loading tags component").ToHTML()
	}

	jsContent, err := tagsFiles.ReadFile("tags.js")
	if err != nil {
		u.Logger().Error("Failed to read tags JavaScript file", "error", err)
		return hb.Div().HTML("Error loading tags component").ToHTML()
	}

	linksHelper := shared.NewLinksFromRequest(r)
	initScript := hb.Script(`
		const urlTagsLoad = '` + linksHelper.TagManager(map[string]string{"action": "load-tags"}) + `';
		const urlTagPostsLoad = '` + linksHelper.TagManager(map[string]string{"action": "load-tag-posts", "tag_id": "TAG_ID_PLACEHOLDER"}) + `';
		const urlTagCreate = '` + linksHelper.TagManager(map[string]string{"action": "create-tag"}) + `';
		const urlTagUpdate = '` + linksHelper.TagManager(map[string]string{"action": "update-tag", "tag_id": "TAG_ID_PLACEHOLDER"}) + `';
		const urlTagDelete = '` + linksHelper.TagManager(map[string]string{"action": "delete-tag"}) + `';
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

	return u.Layout(nil, r, "Blog | Tag Manager", content.ToHTML(), struct {
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

func (u *ui) handleLoadTags(r *http.Request) string {
	ctx := r.Context()

	blogStore := u.Store()
	if blogStore == nil {
		return api.Error("Blog store not available").ToString()
	}

	tagTaxonomy, err := u.ensureTaxonomy(ctx, blogStore)
	if err != nil {
		return api.Error("Failed to ensure taxonomy: " + err.Error()).ToString()
	}

	terms, err := blogStore.TermList(ctx, blogstore.TermQueryOptions{
		TaxonomyID: tagTaxonomy.GetID(),
		OrderBy:    "name",
		SortOrder:  "asc",
	})
	if err != nil {
		u.Logger().Error("Failed to load tags", "error", err)
		return api.Error("Failed to load tags").ToString()
	}

	tagList := []map[string]any{}
	for _, term := range terms {
		tagList = append(tagList, map[string]any{
			"id":    term.GetID(),
			"name":  term.GetName(),
			"slug":  term.GetSlug(),
			"count": term.GetCount(),
		})
	}

	return api.SuccessWithData("Tags loaded successfully", map[string]any{
		"tags": tagList,
	}).ToString()
}

func (u *ui) handleLoadTagPosts(r *http.Request) string {
	ctx := r.Context()

	tagID := r.URL.Query().Get("tag_id")
	if tagID == "" {
		return api.Error("Tag ID is required").ToString()
	}

	blogStore := u.Store()
	if blogStore == nil {
		return api.Error("Blog store not available").ToString()
	}

	// Get the tag to verify it exists
	tag, err := blogStore.TermFindByID(ctx, tagID)
	if err != nil {
		return api.Error("Tag not found").ToString()
	}

	// Get posts associated with this tag (all statuses)
	posts, err := blogStore.PostListByTermID(ctx, tagID, blogstore.PostQueryOptions{
		OrderBy:   "published_at",
		SortOrder: "desc",
		Limit:     100,
	})
	if err != nil {
		u.Logger().Error("Failed to load posts for tag", "error", err, "tag_id", tagID)
		return api.Error("Failed to load posts for tag").ToString()
	}

	postList := []map[string]any{}
	for _, post := range posts {
		postList = append(postList, map[string]any{
			"id":           post.GetID(),
			"title":        post.GetTitle(),
			"slug":         post.GetSlug(),
			"status":       post.GetStatus(),
			"published_at": post.GetPublishedAt(),
		})
	}

	return api.SuccessWithData("Tag information loaded", map[string]any{
		"tag": map[string]any{
			"id":    tag.GetID(),
			"name":  tag.GetName(),
			"slug":  tag.GetSlug(),
			"count": tag.GetCount(),
		},
		"posts": postList,
	}).ToString()
}

func (u *ui) handleCreateTag(w http.ResponseWriter, r *http.Request) string {
	ctx := r.Context()

	if r.Method != http.MethodPost {
		return api.Error("Method not allowed").ToString()
	}

	var reqData struct {
		Name string `json:"name"`
		Slug string `json:"slug"`
	}

	if err := json.NewDecoder(r.Body).Decode(&reqData); err != nil {
		return api.Error("Invalid request body").ToString()
	}

	if reqData.Name == "" {
		return api.Error("Tag name is required").ToString()
	}

	blogStore := u.Store()
	if blogStore == nil {
		return api.Error("Blog store not available").ToString()
	}

	tagTaxonomy, err := u.ensureTaxonomy(ctx, blogStore)
	if err != nil {
		return api.Error("Failed to ensure taxonomy: " + err.Error()).ToString()
	}

	slug := reqData.Slug
	if slug == "" {
		slug = str.Slugify(reqData.Name, '-')
	}

	term := blogstore.NewTerm()
	term.SetID(uid.HumanUid()[:8])
	term.SetName(reqData.Name)
	term.SetSlug(slug)
	term.SetTaxonomyID(tagTaxonomy.GetID())

	if err := blogStore.TermCreate(ctx, term); err != nil {
		u.Logger().Error("Failed to create tag", "error", err)
		return api.Error("Failed to create tag").ToString()
	}

	return api.SuccessWithData("Tag created successfully", map[string]any{
		"id":   term.GetID(),
		"name": term.GetName(),
		"slug": term.GetSlug(),
	}).ToString()
}

func (u *ui) handleUpdateTag(w http.ResponseWriter, r *http.Request) string {
	ctx := r.Context()

	if r.Method != http.MethodPut && r.Method != http.MethodPost {
		return api.Error("Method not allowed").ToString()
	}

	tagID := r.URL.Query().Get("tag_id")
	if tagID == "" {
		return api.Error("Tag ID is required").ToString()
	}

	var reqData struct {
		Name string `json:"name"`
		Slug string `json:"slug"`
	}

	if err := json.NewDecoder(r.Body).Decode(&reqData); err != nil {
		return api.Error("Invalid request body").ToString()
	}

	if reqData.Name == "" {
		return api.Error("Tag name is required").ToString()
	}

	blogStore := u.Store()
	if blogStore == nil {
		return api.Error("Blog store not available").ToString()
	}

	term, err := blogStore.TermFindByID(ctx, tagID)
	if err != nil {
		return api.Error("Tag not found").ToString()
	}

	slug := reqData.Slug
	if slug == "" {
		slug = str.Slugify(reqData.Name, '-')
	}

	term.SetName(reqData.Name)
	term.SetSlug(slug)

	if err := blogStore.TermUpdate(ctx, term); err != nil {
		u.Logger().Error("Failed to update tag", "error", err)
		return api.Error("Failed to update tag").ToString()
	}

	return api.SuccessWithData("Tag updated successfully", map[string]any{
		"id":   term.GetID(),
		"name": term.GetName(),
		"slug": term.GetSlug(),
	}).ToString()
}

func (u *ui) handleDeleteTag(w http.ResponseWriter, r *http.Request) string {
	ctx := r.Context()

	if r.Method != http.MethodPost {
		return api.Error("Method not allowed").ToString()
	}

	var reqData struct {
		TagID string `json:"tag_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&reqData); err != nil {
		reqData.TagID = r.FormValue("tag_id")
	}

	if reqData.TagID == "" {
		return api.Error("Tag ID is required").ToString()
	}

	blogStore := u.Store()
	if blogStore == nil {
		return api.Error("Blog store not available").ToString()
	}

	term, err := blogStore.TermFindByID(ctx, reqData.TagID)
	if err != nil {
		u.Logger().Error("Failed to find tag for delete", "error", err)
		return api.Error("Tag not found").ToString()
	}

	if err := blogStore.TermDelete(ctx, term); err != nil {
		u.Logger().Error("Failed to delete tag", "error", err)
		return api.Error("Failed to delete tag").ToString()
	}

	return api.SuccessWithData("Tag deleted successfully", map[string]any{}).ToString()
}

func (u *ui) ensureTaxonomy(ctx context.Context, store blogstore.StoreInterface) (blogstore.TaxonomyInterface, error) {
	tagTaxonomy, err := store.TaxonomyFindBySlug(ctx, blogstore.TAXONOMY_TAG)
	if err != nil || tagTaxonomy == nil {
		u.Logger().Info("Creating tag taxonomy")
		tagTaxonomy = blogstore.NewTaxonomy()
		tagTaxonomy.SetName("Tag")
		tagTaxonomy.SetSlug(blogstore.TAXONOMY_TAG)
		tagTaxonomy.SetDescription("Blog post tags")
		if err := store.TaxonomyCreate(ctx, tagTaxonomy); err != nil {
			return nil, err
		}
	}

	if tagTaxonomy == nil {
		return nil, errors.New("tag taxonomy is nil after ensure")
	}

	return tagTaxonomy, nil
}
