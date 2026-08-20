package category_manager

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
)

//go:embed *.html
//go:embed *.js
var categoriesFiles embed.FS

// UiInterface defines the category manager controller's UI interface
type UiInterface interface {
	shared.UiInterface
	CategoryManager(w http.ResponseWriter, r *http.Request)
}

// ui implements UiInterface
type ui struct {
	shared.UiBase
}

// UI creates a new category manager controller UI from the given config
func UI(config shared.UiConfig) UiInterface {
	return &ui{UiBase: shared.NewUiBase(config)}
}

// CategoryManager handles the category manager controller requests
func (u *ui) CategoryManager(w http.ResponseWriter, r *http.Request) {
	html := u.Handler(w, r)
	if html != "" {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(html))
	}
}

// Handler processes the category manager controller request and returns HTML
func (u *ui) Handler(w http.ResponseWriter, r *http.Request) string {
	action := r.URL.Query().Get("action")

	switch action {
	case "load-categories":
		return u.handleLoadCategories(r)
	case "create-category":
		return u.handleCreateCategory(w, r)
	case "update-category":
		return u.handleUpdateCategory(w, r)
	case "delete-category":
		return u.handleDeleteCategory(w, r)
	case "reorder-categories":
		return u.handleReorderCategories(w, r)
	default:
		return u.renderPage(r)
	}
}

func (u *ui) renderPage(r *http.Request) string {
	breadcrumbs := shared.Breadcrumbs([]shared.Breadcrumb{
		{Name: "Home", URL: shared.AdminHomeURL(r)},
		{Name: "Blog", URL: shared.URLR(r, shared.CONTROLLER_DASHBOARD, nil)},
		{Name: "Categories", URL: ""},
	})

	heading := hb.Heading1().HTML("Blog. Category Manager")

	htmlContent, err := categoriesFiles.ReadFile("categories.html")
	if err != nil {
		u.Logger().Error("Failed to read categories HTML template", "error", err)
		return hb.Div().HTML("Error loading categories component").ToHTML()
	}

	jsContent, err := categoriesFiles.ReadFile("categories.js")
	if err != nil {
		u.Logger().Error("Failed to read categories JavaScript file", "error", err)
		return hb.Div().HTML("Error loading categories component").ToHTML()
	}

	sortableCDN := hb.Script("").Src("https://cdn.jsdelivr.net/npm/sortablejs@1.15.0/Sortable.min.js")

	linksHelper := shared.NewLinksFromRequest(r)
	initScript := hb.Script(`
		const urlCategoriesLoad = '` + linksHelper.CategoryManager(map[string]string{"action": "load-categories"}) + `';
		const urlCategoryCreate = '` + linksHelper.CategoryManager(map[string]string{"action": "create-category"}) + `';
		const urlCategoryUpdate = '` + linksHelper.CategoryManager(map[string]string{"action": "update-category", "category_id": "CATEGORY_ID_PLACEHOLDER"}) + `';
		const urlCategoryDelete = '` + linksHelper.CategoryManager(map[string]string{"action": "delete-category"}) + `';
		const urlCategoriesReorder = '` + linksHelper.CategoryManager(map[string]string{"action": "reorder-categories"}) + `';
	`)

	htmlTemplate := hb.Wrap().HTML(string(htmlContent))
	componentScript := hb.Script(string(jsContent))

	vueContainer := hb.Div().
		Child(shared.VueLoaderScript()).
		Child(sortableCDN).
		Child(htmlTemplate).
		Child(initScript).
		Child(componentScript)

	content := hb.Div().
		Class("container").
		Child(heading).
		Child(breadcrumbs).
		Child(hb.HR()).
		Child(vueContainer)

	return u.Layout(nil, r, "Blog | Category Manager", content.ToHTML(), struct {
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

func (u *ui) handleLoadCategories(r *http.Request) string {
	ctx := r.Context()

	blogStore := u.Store()
	if blogStore == nil {
		return api.Error("Blog store not available").ToString()
	}

	categoryTaxonomy, err := u.ensureTaxonomy(ctx, blogStore)
	if err != nil {
		return api.Error("Failed to ensure taxonomy: " + err.Error()).ToString()
	}

	terms, err := blogStore.TermList(ctx, blogstore.TermQueryOptions{
		TaxonomyID: categoryTaxonomy.GetID(),
		OrderBy:    "sequence",
		SortOrder:  "asc",
	})
	if err != nil {
		u.Logger().Error("Failed to load categories", "error", err)
		return api.Error("Failed to load categories").ToString()
	}

	categoryList := []map[string]any{}
	for _, term := range terms {
		categoryList = append(categoryList, map[string]any{
			"id":          term.GetID(),
			"name":        term.GetName(),
			"slug":        term.GetSlug(),
			"description": term.GetDescription(),
			"count":       term.GetCount(),
		})
	}

	return api.SuccessWithData("Categories loaded successfully", map[string]any{
		"categories": categoryList,
	}).ToString()
}

func (u *ui) handleCreateCategory(w http.ResponseWriter, r *http.Request) string {
	ctx := r.Context()

	if r.Method != http.MethodPost {
		return api.Error("Method not allowed").ToString()
	}

	var reqData struct {
		Name        string `json:"name"`
		Slug        string `json:"slug"`
		Description string `json:"description"`
	}

	if err := json.NewDecoder(r.Body).Decode(&reqData); err != nil {
		return api.Error("Invalid request body").ToString()
	}

	if reqData.Name == "" {
		return api.Error("Category name is required").ToString()
	}

	blogStore := u.Store()
	if blogStore == nil {
		return api.Error("Blog store not available").ToString()
	}

	categoryTaxonomy, err := u.ensureTaxonomy(ctx, blogStore)
	if err != nil {
		return api.Error("Failed to ensure taxonomy: " + err.Error()).ToString()
	}

	slug := reqData.Slug
	if slug == "" {
		slug = str.Slugify(reqData.Name, '-')
	}

	term := blogstore.NewTerm()
	term.SetName(reqData.Name)
	term.SetSlug(slug)
	term.SetDescription(reqData.Description)
	term.SetTaxonomyID(categoryTaxonomy.GetID())

	if err := blogStore.TermCreate(ctx, term); err != nil {
		u.Logger().Error("Failed to create category", "error", err)
		return api.Error("Failed to create category").ToString()
	}

	return api.SuccessWithData("Category created successfully", map[string]any{
		"id":   term.GetID(),
		"name": term.GetName(),
		"slug": term.GetSlug(),
	}).ToString()
}

func (u *ui) handleUpdateCategory(w http.ResponseWriter, r *http.Request) string {
	ctx := r.Context()

	if r.Method != http.MethodPut && r.Method != http.MethodPost {
		return api.Error("Method not allowed").ToString()
	}

	categoryID := r.URL.Query().Get("category_id")
	if categoryID == "" {
		return api.Error("Category ID is required").ToString()
	}

	var reqData struct {
		Name        string `json:"name"`
		Slug        string `json:"slug"`
		Description string `json:"description"`
	}

	if err := json.NewDecoder(r.Body).Decode(&reqData); err != nil {
		return api.Error("Invalid request body").ToString()
	}

	if reqData.Name == "" {
		return api.Error("Category name is required").ToString()
	}

	blogStore := u.Store()
	if blogStore == nil {
		return api.Error("Blog store not available").ToString()
	}

	term, err := blogStore.TermFindByID(ctx, categoryID)
	if err != nil {
		return api.Error("Category not found").ToString()
	}

	slug := reqData.Slug
	if slug == "" {
		slug = str.Slugify(reqData.Name, '-')
	}

	term.SetName(reqData.Name)
	term.SetSlug(slug)
	term.SetDescription(reqData.Description)

	if err := blogStore.TermUpdate(ctx, term); err != nil {
		u.Logger().Error("Failed to update category", "error", err)
		return api.Error("Failed to update category").ToString()
	}

	return api.SuccessWithData("Category updated successfully", map[string]any{
		"id":   term.GetID(),
		"name": term.GetName(),
		"slug": term.GetSlug(),
	}).ToString()
}

func (u *ui) handleReorderCategories(w http.ResponseWriter, r *http.Request) string {
	ctx := r.Context()

	if r.Method != http.MethodPost {
		return api.Error("Method not allowed").ToString()
	}

	var reqData struct {
		CategoryIDs []string `json:"category_ids"`
	}

	if err := json.NewDecoder(r.Body).Decode(&reqData); err != nil {
		return api.Error("Invalid request body").ToString()
	}

	blogStore := u.Store()
	if blogStore == nil {
		return api.Error("Blog store not available").ToString()
	}

	// Update each category's sequence based on the new order
	for seq, categoryID := range reqData.CategoryIDs {
		term, err := blogStore.TermFindByID(ctx, categoryID)
		if err != nil {
			u.Logger().Error("Failed to find category for reorder", "category_id", categoryID, "error", err)
			continue
		}
		term.SetSequence(seq)
		if err := blogStore.TermUpdate(ctx, term); err != nil {
			u.Logger().Error("Failed to update category sequence", "category_id", categoryID, "error", err)
			return api.Error("Failed to save category order").ToString()
		}
	}

	u.Logger().Info("Categories reordered successfully", "count", len(reqData.CategoryIDs))
	return api.Success("Categories reordered successfully").ToString()
}

func (u *ui) handleDeleteCategory(w http.ResponseWriter, r *http.Request) string {
	ctx := r.Context()

	if r.Method != http.MethodPost {
		return api.Error("Method not allowed").ToString()
	}

	var reqData struct {
		CategoryID string `json:"category_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&reqData); err != nil {
		reqData.CategoryID = r.FormValue("category_id")
	}

	if reqData.CategoryID == "" {
		return api.Error("Category ID is required").ToString()
	}

	blogStore := u.Store()
	if blogStore == nil {
		return api.Error("Blog store not available").ToString()
	}

	term, err := blogStore.TermFindByID(ctx, reqData.CategoryID)
	if err != nil {
		u.Logger().Error("Failed to find category for delete", "error", err)
		return api.Error("Category not found").ToString()
	}

	if err := blogStore.TermDelete(ctx, term); err != nil {
		u.Logger().Error("Failed to delete category", "error", err)
		return api.Error("Failed to delete category").ToString()
	}

	return api.SuccessWithData("Category deleted successfully", map[string]any{}).ToString()
}

func (u *ui) ensureTaxonomy(ctx context.Context, store blogstore.StoreInterface) (blogstore.TaxonomyInterface, error) {
	categoryTaxonomy, err := store.TaxonomyFindBySlug(ctx, blogstore.TAXONOMY_CATEGORY)
	if err != nil || categoryTaxonomy == nil {
		u.Logger().Info("Creating category taxonomy")
		categoryTaxonomy = blogstore.NewTaxonomy()
		categoryTaxonomy.SetName("Category")
		categoryTaxonomy.SetSlug(blogstore.TAXONOMY_CATEGORY)
		categoryTaxonomy.SetDescription("Blog post categories")
		if err := store.TaxonomyCreate(ctx, categoryTaxonomy); err != nil {
			return nil, err
		}
	}

	if categoryTaxonomy == nil {
		return nil, errors.New("category taxonomy is nil after ensure")
	}

	return categoryTaxonomy, nil
}
