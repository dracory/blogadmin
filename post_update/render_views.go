package post_update

import (
	"net/http"

	"github.com/dracory/blogadmin/shared"
	"github.com/dracory/blogstore"
	"github.com/dracory/cdn"
	"github.com/dracory/hb"
)

func (u *ui) renderCategoriesView(r *http.Request, post blogstore.PostInterface) hb.TagInterface {
	htmlContent, err := postCategoriesFiles.ReadFile("post_categories.html")
	if err != nil {
		u.Logger().Error("Failed to read post categories HTML template", "error", err)
		return hb.Div().HTML("Error loading categories component")
	}

	jsContent, err := postCategoriesFiles.ReadFile("post_categories.js")
	if err != nil {
		u.Logger().Error("Failed to read post categories JavaScript file", "error", err)
		return hb.Div().HTML("Error loading categories component")
	}

	vueCDN := hb.Script("").Src(cdn.VueJs_3_5_32())
	linksHelper := shared.NewLinksFromRequest(r)

	initScript := hb.Script(`
		const postID = '` + post.GetID() + `';
		const urlCategoriesLoad = '` + linksHelper.PostUpdate(map[string]string{"post_id": post.GetID(), "action": "load-categories"}) + `';
		const urlCategoryAdd = '` + linksHelper.PostUpdate(map[string]string{"post_id": post.GetID(), "action": "add-category"}) + `';
		const urlCategoryRemove = '` + linksHelper.PostUpdate(map[string]string{"post_id": post.GetID(), "action": "remove-category"}) + `';
	`)

	htmlTemplate := hb.Wrap().HTML(string(htmlContent))
	componentScript := hb.Script(string(jsContent))

	vueContainer := hb.Div().
		Child(vueCDN).
		Child(htmlTemplate).
		Child(initScript).
		Child(componentScript)

	return vueContainer
}

func (u *ui) renderTagsView(r *http.Request, post blogstore.PostInterface) hb.TagInterface {
	htmlContent, err := postCategoriesFiles.ReadFile("post_tags.html")
	if err != nil {
		u.Logger().Error("Failed to read post tags HTML template", "error", err)
		return hb.Div().HTML("Error loading tags component")
	}

	jsContent, err := postCategoriesFiles.ReadFile("post_tags.js")
	if err != nil {
		u.Logger().Error("Failed to read post tags JavaScript file", "error", err)
		return hb.Div().HTML("Error loading tags component")
	}

	vueCDN := hb.Script("").Src(cdn.VueJs_3_5_32())
	linksHelper := shared.NewLinksFromRequest(r)

	initScript := hb.Script(`
		const postID = '` + post.GetID() + `';
		const urlTagsLoad = '` + linksHelper.PostUpdate(map[string]string{"post_id": post.GetID(), "action": "load-tags"}) + `';
		const urlTagAdd = '` + linksHelper.PostUpdate(map[string]string{"post_id": post.GetID(), "action": "add-tag"}) + `';
		const urlTagRemove = '` + linksHelper.PostUpdate(map[string]string{"post_id": post.GetID(), "action": "remove-tag"}) + `';
	`)

	htmlTemplate := hb.Wrap().HTML(string(htmlContent))
	componentScript := hb.Script(string(jsContent))

	vueContainer := hb.Div().
		Child(vueCDN).
		Child(htmlTemplate).
		Child(initScript).
		Child(componentScript)

	return vueContainer
}

func (u *ui) renderDetailsView(r *http.Request, post blogstore.PostInterface) hb.TagInterface {
	htmlContent, err := postCategoriesFiles.ReadFile("post_details.html")
	if err != nil {
		u.Logger().Error("Failed to read post details HTML template", "error", err)
		return hb.Div().HTML("Error loading details component")
	}

	jsContent, err := postCategoriesFiles.ReadFile("post_details.js")
	if err != nil {
		u.Logger().Error("Failed to read post details JavaScript file", "error", err)
		return hb.Div().HTML("Error loading details component")
	}

	vueCDN := hb.Script("").Src(cdn.VueJs_3_5_32())
	linksHelper := shared.NewLinksFromRequest(r)

	initScript := hb.Script(`
		const postId = '` + post.GetID() + `';
		const urlDetailsLoad = '` + linksHelper.PostUpdate(map[string]string{"post_id": post.GetID(), "action": "load-details"}) + `';
		const urlDetailsSave = '` + linksHelper.PostUpdate(map[string]string{"post_id": post.GetID(), "action": "save-details"}) + `';
		const urlRegenerateImage = '` + linksHelper.PostUpdate(map[string]string{"post_id": post.GetID(), "action": "regenerate-image"}) + `';
		const urlMediaLoad = '` + linksHelper.PostUpdate(map[string]string{"post_id": post.GetID(), "action": "load-media"}) + `';
	`)

	htmlTemplate := hb.Wrap().HTML(string(htmlContent))
	componentScript := hb.Script(string(jsContent))

	vueContainer := hb.Div().
		Child(vueCDN).
		Child(htmlTemplate).
		Child(initScript).
		Child(componentScript)

	return vueContainer
}

func (u *ui) renderContentView(r *http.Request, post blogstore.PostInterface) hb.TagInterface {
	htmlContent, err := postCategoriesFiles.ReadFile("post_content.html")
	if err != nil {
		u.Logger().Error("Failed to read post content HTML template", "error", err)
		return hb.Div().HTML("Error loading content component")
	}

	jsContent, err := postCategoriesFiles.ReadFile("post_content.js")
	if err != nil {
		u.Logger().Error("Failed to read post content JavaScript file", "error", err)
		return hb.Div().HTML("Error loading content component")
	}

	vueCDN := hb.Script("").Src(cdn.VueJs_3_5_32())
	linksHelper := shared.NewLinksFromRequest(r)

	initScript := hb.Script(`
		const postId = '` + post.GetID() + `';
		const urlContentLoad = '` + linksHelper.PostUpdate(map[string]string{"post_id": post.GetID(), "action": "load-content"}) + `';
		const urlContentSave = '` + linksHelper.PostUpdate(map[string]string{"post_id": post.GetID(), "action": "save-content"}) + `';
		const urlBlockEditorHandle = '` + linksHelper.PostUpdate(map[string]string{"post_id": post.GetID(), "action": "blockeditor-handle"}) + `';
	`)

	htmlTemplate := hb.Wrap().HTML(string(htmlContent))
	componentScript := hb.Script(string(jsContent))

	vueContainer := hb.Div().
		Child(vueCDN).
		Child(htmlTemplate).
		Child(initScript).
		Child(componentScript)

	return vueContainer
}

func (u *ui) renderSEOView(r *http.Request, post blogstore.PostInterface) hb.TagInterface {
	htmlContent, err := postCategoriesFiles.ReadFile("post_seo.html")
	if err != nil {
		u.Logger().Error("Failed to read post SEO HTML template", "error", err)
		return hb.Div().HTML("Error loading SEO component")
	}

	jsContent, err := postCategoriesFiles.ReadFile("post_seo.js")
	if err != nil {
		u.Logger().Error("Failed to read post SEO JavaScript file", "error", err)
		return hb.Div().HTML("Error loading SEO component")
	}

	vueCDN := hb.Script("").Src(cdn.VueJs_3_5_32())
	linksHelper := shared.NewLinksFromRequest(r)

	initScript := hb.Script(`
		const postId = '` + post.GetID() + `';
		const urlSEOLoad = '` + linksHelper.PostUpdate(map[string]string{"post_id": post.GetID(), "action": "load-seo"}) + `';
		const urlSEOSave = '` + linksHelper.PostUpdate(map[string]string{"post_id": post.GetID(), "action": "save-seo"}) + `';
	`)

	htmlTemplate := hb.Wrap().HTML(string(htmlContent))
	componentScript := hb.Script(string(jsContent))

	vueContainer := hb.Div().
		Child(vueCDN).
		Child(htmlTemplate).
		Child(initScript).
		Child(componentScript)

	return vueContainer
}

func (u *ui) renderVersioningModal(r *http.Request, post blogstore.PostInterface) hb.TagInterface {
	htmlContent, err := postCategoriesFiles.ReadFile("post_versioning.html")
	if err != nil {
		u.Logger().Error("Failed to read post versioning HTML template", "error", err)
		return hb.Div().HTML("Error loading versioning component")
	}

	jsContent, err := postCategoriesFiles.ReadFile("post_versioning.js")
	if err != nil {
		u.Logger().Error("Failed to read post versioning JavaScript file", "error", err)
		return hb.Div().HTML("Error loading versioning component")
	}

	htmlStr := string(htmlContent)
	htmlTemplate := hb.Wrap().HTML(htmlStr)
	componentScript := hb.Script(string(jsContent))

	linksHelper := shared.NewLinksFromRequest(r)

	// Config script for versioning component (uses global object to avoid const redeclaration)
	configScript := hb.Script(`
		window.postVersioningConfig = window.postVersioningConfig || {};
		window.postVersioningConfig.postId = '` + post.GetID() + `';
		window.postVersioningConfig.urlVersionsLoad = '` + linksHelper.PostUpdate(map[string]string{"post_id": post.GetID(), "action": "load-versions"}) + `';
		window.postVersioningConfig.urlVersionDetail = '` + linksHelper.PostUpdate(map[string]string{"post_id": post.GetID(), "action": "load-version-detail"}) + `';
		window.postVersioningConfig.urlVersionRestore = '` + linksHelper.PostUpdate(map[string]string{"post_id": post.GetID(), "action": "restore-version-attributes"}) + `';
	`)

	// CSS for v-cloak
	vCloakStyle := hb.Style(`
		[v-cloak] { display: none; }
	`)

	vueCDN := hb.Script("").Src(cdn.VueJs_3_5_32())

	vueContainer := hb.Div().
		Child(vCloakStyle).
		Child(vueCDN).
		Child(configScript).
		Child(htmlTemplate).
		Child(componentScript)

	return vueContainer
}

func (u *ui) renderMediaView(r *http.Request, post blogstore.PostInterface) hb.TagInterface {
	return u.newPostMediaComponent().Render(r, post)
}
