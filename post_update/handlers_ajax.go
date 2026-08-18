package post_update

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/dracory/api"
	"github.com/dracory/blogai"
	"github.com/dracory/blogstore"
	"github.com/dracory/neat"
	"github.com/dracory/req"
	"github.com/dracory/versionstore"
)

// == Categories =============================================================

func (u *ui) handleLoadCategories(w http.ResponseWriter, r *http.Request) string {
	ctx := r.Context()

	postID := req.GetStringTrimmed(r, "post_id")
	if postID == "" {
		return api.Error("Post ID is required").ToString()
	}

	blogStore := u.Store()
	if blogStore == nil {
		return api.Error("Blog store not available").ToString()
	}

	// Get category taxonomy
	categoryTaxonomy, err := blogStore.TaxonomyFindBySlug(ctx, blogstore.TAXONOMY_CATEGORY)
	if err != nil || categoryTaxonomy == nil {
		return api.Error("Category taxonomy not found").ToString()
	}

	// Get all available categories
	allCategories, err := blogStore.TermList(ctx, blogstore.TermQueryOptions{
		TaxonomyID: categoryTaxonomy.GetID(),
		OrderBy:    "sequence",
		SortOrder:  "asc",
	})
	if err != nil {
		u.Logger().Error("Failed to load categories", "error", err)
		return api.Error("Failed to load categories").ToString()
	}

	// Get categories assigned to this post
	assignedCategoryIDs := make(map[string]bool)
	postCategories, err := blogStore.TermListByPostID(ctx, postID, blogstore.TAXONOMY_CATEGORY)
	if err == nil {
		for _, category := range postCategories {
			assignedCategoryIDs[category.GetID()] = true
		}
	}

	categoryList := []map[string]any{}
	for _, category := range allCategories {
		categoryList = append(categoryList, map[string]any{
			"id":          category.GetID(),
			"name":        category.GetName(),
			"slug":        category.GetSlug(),
			"description": category.GetDescription(),
			"assigned":    assignedCategoryIDs[category.GetID()],
		})
	}

	return api.SuccessWithData("Categories loaded successfully", map[string]any{
		"categories": categoryList,
	}).ToString()
}

func (u *ui) handleAddCategory(w http.ResponseWriter, r *http.Request) string {
	ctx := r.Context()

	if r.Method != http.MethodPost {
		return api.Error("Method not allowed").ToString()
	}

	postID := req.GetStringTrimmed(r, "post_id")

	var reqData struct {
		CategoryID string `json:"category_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&reqData); err != nil {
		return api.Error("Invalid request body").ToString()
	}

	if reqData.CategoryID == "" {
		return api.Error("Category ID is required").ToString()
	}

	blogStore := u.Store()
	if blogStore == nil {
		return api.Error("Blog store not available").ToString()
	}

	// Add category to post using PostAddTerm
	if err := blogStore.PostAddTerm(ctx, postID, reqData.CategoryID); err != nil {
		u.Logger().Error("Failed to add category to post", "error", err)
		return api.Error("Failed to add category to post").ToString()
	}

	return api.Success("Category added to post successfully").ToString()
}

func (u *ui) handleRemoveCategory(w http.ResponseWriter, r *http.Request) string {
	ctx := r.Context()

	if r.Method != http.MethodPost {
		return api.Error("Method not allowed").ToString()
	}

	postID := req.GetStringTrimmed(r, "post_id")

	var reqData struct {
		CategoryID string `json:"category_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&reqData); err != nil {
		return api.Error("Invalid request body").ToString()
	}

	if reqData.CategoryID == "" {
		return api.Error("Category ID is required").ToString()
	}

	blogStore := u.Store()
	if blogStore == nil {
		return api.Error("Blog store not available").ToString()
	}

	// Remove category from post using PostRemoveTerm
	if err := blogStore.PostRemoveTerm(ctx, postID, reqData.CategoryID); err != nil {
		u.Logger().Error("Failed to remove category from post", "error", err)
		return api.Error("Failed to remove category from post").ToString()
	}

	return api.Success("Category removed from post successfully").ToString()
}

// == Tags ===================================================================

func (u *ui) handleLoadTags(w http.ResponseWriter, r *http.Request) string {
	ctx := r.Context()

	postID := req.GetStringTrimmed(r, "post_id")
	if postID == "" {
		return api.Error("Post ID is required").ToString()
	}

	blogStore := u.Store()
	if blogStore == nil {
		return api.Error("Blog store not available").ToString()
	}

	// Get tag taxonomy
	tagTaxonomy, err := blogStore.TaxonomyFindBySlug(ctx, blogstore.TAXONOMY_TAG)
	if err != nil || tagTaxonomy == nil {
		return api.Error("Tag taxonomy not found").ToString()
	}

	// Get all available tags
	allTags, err := blogStore.TermList(ctx, blogstore.TermQueryOptions{
		TaxonomyID: tagTaxonomy.GetID(),
		OrderBy:    "name",
		SortOrder:  "asc",
	})
	if err != nil {
		u.Logger().Error("Failed to load tags", "error", err)
		return api.Error("Failed to load tags").ToString()
	}

	// Get tags assigned to this post
	assignedTagIDs := make(map[string]bool)
	postTags, err := blogStore.TermListByPostID(ctx, postID, blogstore.TAXONOMY_TAG)
	if err == nil {
		for _, tag := range postTags {
			assignedTagIDs[tag.GetID()] = true
		}
	}

	tagList := []map[string]any{}
	for _, tag := range allTags {
		tagList = append(tagList, map[string]any{
			"id":       tag.GetID(),
			"name":     tag.GetName(),
			"slug":     tag.GetSlug(),
			"assigned": assignedTagIDs[tag.GetID()],
		})
	}

	return api.SuccessWithData("Tags loaded successfully", map[string]any{
		"tags": tagList,
	}).ToString()
}

func (u *ui) handleAddTag(w http.ResponseWriter, r *http.Request) string {
	ctx := r.Context()

	if r.Method != http.MethodPost {
		return api.Error("Method not allowed").ToString()
	}

	postID := req.GetStringTrimmed(r, "post_id")

	var reqData struct {
		TagID string `json:"tag_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&reqData); err != nil {
		return api.Error("Invalid request body").ToString()
	}

	if reqData.TagID == "" {
		return api.Error("Tag ID is required").ToString()
	}

	blogStore := u.Store()
	if blogStore == nil {
		return api.Error("Blog store not available").ToString()
	}

	// Add tag to post using PostAddTerm
	if err := blogStore.PostAddTerm(ctx, postID, reqData.TagID); err != nil {
		u.Logger().Error("Failed to add tag to post", "error", err)
		return api.Error("Failed to add tag to post").ToString()
	}

	return api.Success("Tag added to post successfully").ToString()
}

func (u *ui) handleRemoveTag(w http.ResponseWriter, r *http.Request) string {
	ctx := r.Context()

	if r.Method != http.MethodPost {
		return api.Error("Method not allowed").ToString()
	}

	postID := req.GetStringTrimmed(r, "post_id")

	var reqData struct {
		TagID string `json:"tag_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&reqData); err != nil {
		return api.Error("Invalid request body").ToString()
	}

	if reqData.TagID == "" {
		return api.Error("Tag ID is required").ToString()
	}

	blogStore := u.Store()
	if blogStore == nil {
		return api.Error("Blog store not available").ToString()
	}

	// Remove tag from post using PostRemoveTerm
	if err := blogStore.PostRemoveTerm(ctx, postID, reqData.TagID); err != nil {
		u.Logger().Error("Failed to remove tag from post", "error", err)
		return api.Error("Failed to remove tag from post").ToString()
	}

	return api.Success("Tag removed from post successfully").ToString()
}

// == Details ================================================================

func (u *ui) handleLoadDetails(w http.ResponseWriter, r *http.Request) string {
	ctx := r.Context()

	postID := req.GetStringTrimmed(r, "post_id")
	if postID == "" {
		return api.Error("Post ID is required").ToString()
	}

	blogStore := u.Store()
	if blogStore == nil {
		return api.Error("Blog store not available").ToString()
	}

	post, err := blogStore.PostFindByID(ctx, postID)
	if err != nil || post == nil {
		return api.Error("Post not found").ToString()
	}

	return api.SuccessWithData("Details loaded successfully", map[string]any{
		"status":       post.GetStatus(),
		"image_url":    post.GetImageUrl(),
		"featured":     post.GetFeatured(),
		"published_at": post.GetPublishedAtCarbon().ToDateTimeString(),
		"editor":       post.GetEditor(),
		"memo":         post.GetMemo(),
	}).ToString()
}

func (u *ui) handleSaveDetails(w http.ResponseWriter, r *http.Request) string {
	ctx := r.Context()

	if r.Method != http.MethodPost {
		return api.Error("Method not allowed").ToString()
	}

	postID := req.GetStringTrimmed(r, "post_id")

	var reqData struct {
		Status      string `json:"post_status"`
		ImageURL    string `json:"post_image_url"`
		Featured    string `json:"post_featured"`
		PublishedAt string `json:"post_published_at"`
		Editor      string `json:"post_editor"`
		Memo        string `json:"post_memo"`
	}

	if err := json.NewDecoder(r.Body).Decode(&reqData); err != nil {
		return api.Error("Invalid request body").ToString()
	}

	if reqData.Status == "" {
		return api.Error("Status is required").ToString()
	}

	blogStore := u.Store()
	if blogStore == nil {
		return api.Error("Blog store not available").ToString()
	}

	post, err := blogStore.PostFindByID(ctx, postID)
	if err != nil || post == nil {
		return api.Error("Post not found").ToString()
	}

	// Normalize published_at - keep existing value if empty
	var publishedAt string
	if strings.TrimSpace(reqData.PublishedAt) == "" {
		// Keep the existing published_at value
		publishedAt = post.GetPublishedAtCarbon().ToDateTimeString()
	} else {
		parsedTime, err := time.Parse("2006-01-02T15:04", reqData.PublishedAt)
		if err != nil {
			return api.Error("Invalid published_at format").ToString()
		}
		publishedAt = parsedTime.Format("2006-01-02 15:04:05")
	}

	post.SetEditor(reqData.Editor)
	post.SetContentType(getContentTypeFromEditor(reqData.Editor))
	post.SetFeatured(reqData.Featured)
	post.SetImageUrl(reqData.ImageURL)
	post.SetMemo(reqData.Memo)
	post.SetPublishedAt(publishedAt)
	post.SetStatus(reqData.Status)

	if err := blogStore.PostUpdate(ctx, post); err != nil {
		u.Logger().Error("Error saving post details", "error", err.Error())
		return api.Error("System error. Saving post failed").ToString()
	}

	if err := createPostVersioning(context.Background(), blogStore, post); err != nil {
		u.Logger().Error("Error creating post versioning", "error", err.Error())
	}

	return api.Success("Post saved successfully").ToString()
}

func (u *ui) handleRegenerateImage(w http.ResponseWriter, r *http.Request) string {
	ctx := r.Context()

	if r.Method != http.MethodPost {
		return api.Error("Method not allowed").ToString()
	}

	postID := req.GetStringTrimmed(r, "post_id")

	var reqData struct {
		PostID string `json:"post_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&reqData); err != nil {
		return api.Error("Invalid request body").ToString()
	}

	blogStore := u.Store()
	if blogStore == nil {
		return api.Error("Blog store not available").ToString()
	}

	post, err := blogStore.PostFindByID(ctx, postID)
	if err != nil || post == nil {
		return api.Error("Post not found").ToString()
	}

	// AI image regeneration requires an LLM engine. If no LlmFactory
	// was provided, return an error to the user instead of panicking.
	llmEngine, err := u.LlmEngine()
	if err != nil || llmEngine == nil {
		return api.Error("AI image generation is not configured").ToString()
	}

	agent := blogai.NewBlogWriterAgent(u.Logger())
	if agent == nil {
		return api.Error("Failed to initialize AI engine").ToString()
	}

	imageURL, err := agent.GenerateImage(llmEngine, post.GetTitle(), post.GetSummary())
	if err != nil {
		u.Logger().Error("BlogAi.PostUpdate.RegenerateImage", "error", err.Error())
		return api.Error("Failed to generate image").ToString()
	}

	post.SetImageUrl(imageURL)
	if err := blogStore.PostUpdate(ctx, post); err != nil {
		u.Logger().Error("BlogAi.PostUpdate.RegenerateImage.Save", "error", err.Error())
		return api.Error("Failed to save generated image").ToString()
	}

	return api.SuccessWithData("Image regenerated successfully", map[string]any{
		"image_url": imageURL,
	}).ToString()
}

// == Content ================================================================

func (u *ui) handleLoadContent(w http.ResponseWriter, r *http.Request) string {
	ctx := r.Context()

	postID := req.GetStringTrimmed(r, "post_id")
	if postID == "" {
		return api.Error("Post ID is required").ToString()
	}

	blogStore := u.Store()
	if blogStore == nil {
		return api.Error("Blog store not available").ToString()
	}

	post, err := blogStore.PostFindByID(ctx, postID)
	if err != nil || post == nil {
		return api.Error("Post not found").ToString()
	}

	return api.SuccessWithData("Content loaded successfully", map[string]any{
		"title":   post.GetTitle(),
		"summary": post.GetSummary(),
		"content": post.GetContent(),
		"editor":  post.GetEditor(),
	}).ToString()
}

func (u *ui) handleSaveContent(w http.ResponseWriter, r *http.Request) string {
	ctx := r.Context()

	if r.Method != http.MethodPost {
		return api.Error("Method not allowed").ToString()
	}

	postID := req.GetStringTrimmed(r, "post_id")

	var reqData struct {
		Title   string `json:"post_title"`
		Summary string `json:"post_summary"`
		Content string `json:"post_content"`
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

	post, err := blogStore.PostFindByID(ctx, postID)
	if err != nil || post == nil {
		return api.Error("Post not found").ToString()
	}

	post.SetTitle(reqData.Title)
	post.SetSummary(reqData.Summary)
	post.SetContent(reqData.Content)

	if err := blogStore.PostUpdate(ctx, post); err != nil {
		u.Logger().Error("Error saving post content", "error", err.Error())
		return api.Error("System error. Saving post failed").ToString()
	}

	if err := createPostVersioning(context.Background(), blogStore, post); err != nil {
		u.Logger().Error("Error creating post versioning", "error", err.Error())
	}

	return api.Success("Post saved successfully").ToString()
}

func (u *ui) handleBlockEditorHandle(w http.ResponseWriter, r *http.Request) string {
	// This is a placeholder for BlockEditor handling
	// The actual implementation would depend on the BlockEditor library
	return api.Error("BlockEditor handle not implemented").ToString()
}

// == SEO ====================================================================

func (u *ui) handleLoadSEO(w http.ResponseWriter, r *http.Request) string {
	ctx := r.Context()

	postID := req.GetStringTrimmed(r, "post_id")
	if postID == "" {
		return api.Error("Post ID is required").ToString()
	}

	blogStore := u.Store()
	if blogStore == nil {
		return api.Error("Blog store not available").ToString()
	}

	post, err := blogStore.PostFindByID(ctx, postID)
	if err != nil || post == nil {
		return api.Error("Post not found").ToString()
	}

	return api.SuccessWithData("SEO data loaded successfully", map[string]any{
		"slug":             post.GetSlug(),
		"canonical_url":    post.GetCanonicalURL(),
		"meta_description": post.GetMetaDescription(),
		"meta_keywords":    post.GetMetaKeywords(),
		"meta_robots":      post.GetMetaRobots(),
		"old_slugs":        post.GetOldSlugs(),
	}).ToString()
}

func (u *ui) handleSaveSEO(w http.ResponseWriter, r *http.Request) string {
	ctx := r.Context()

	if r.Method != http.MethodPost {
		return api.Error("Method not allowed").ToString()
	}

	postID := req.GetStringTrimmed(r, "post_id")

	var reqData struct {
		Slug            string   `json:"post_slug"`
		CanonicalURL    string   `json:"post_canonical_url"`
		MetaDescription string   `json:"post_meta_description"`
		MetaKeywords    string   `json:"post_meta_keywords"`
		MetaRobots      string   `json:"post_meta_robots"`
		OldSlugs        []string `json:"post_old_slugs"`
	}

	if err := json.NewDecoder(r.Body).Decode(&reqData); err != nil {
		return api.Error("Invalid request body").ToString()
	}

	blogStore := u.Store()
	if blogStore == nil {
		return api.Error("Blog store not available").ToString()
	}

	post, err := blogStore.PostFindByID(ctx, postID)
	if err != nil || post == nil {
		return api.Error("Post not found").ToString()
	}

	post.SetSlug(reqData.Slug)
	post.SetCanonicalURL(reqData.CanonicalURL)
	post.SetMetaDescription(reqData.MetaDescription)
	post.SetMetaKeywords(reqData.MetaKeywords)
	post.SetMetaRobots(reqData.MetaRobots)
	if err := post.SetOldSlugs(reqData.OldSlugs); err != nil {
		u.Logger().Error("Error setting old slugs", "error", err.Error())
		return api.Error("System error. Setting old slugs failed").ToString()
	}

	if err := blogStore.PostUpdate(ctx, post); err != nil {
		u.Logger().Error("Error saving post SEO", "error", err.Error())
		return api.Error("System error. Saving post failed").ToString()
	}

	if err := createPostVersioning(context.Background(), blogStore, post); err != nil {
		u.Logger().Error("Error creating post versioning", "error", err.Error())
	}

	return api.Success("Post saved successfully").ToString()
}

// == Versioning =============================================================

func (u *ui) handleLoadVersions(w http.ResponseWriter, r *http.Request) string {
	ctx := r.Context()

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

	if !blogStore.VersioningEnabled() {
		return api.SuccessWithData("Versions loaded successfully", map[string]any{
			"versioning_enabled": false,
			"versions":           []any{},
		}).ToString()
	}

	query := blogstore.NewVersioningQuery()
	query.SetEntityType(blogstore.VERSIONING_TYPE_POST)
	query.SetEntityID(reqData.PostID)
	query.SetOrderBy(versionstore.COLUMN_CREATED_AT)
	query.SetSortOrder(neat.SortDesc)
	query.SetLimit(50)

	u.Logger().Info("Loading versions for post", "post_id", reqData.PostID)

	versions, err := blogStore.VersioningList(ctx, query)
	if err != nil {
		u.Logger().Error("Failed to load versions", "error", err.Error(), "post_id", reqData.PostID)
		return api.Error("Failed to load versions").ToString()
	}

	u.Logger().Info("Versions loaded from query", "count", len(versions), "post_id", reqData.PostID)

	// Filter versions by entity_id as a safety measure (in case versionstore doesn't filter correctly)
	filteredVersions := []blogstore.VersioningInterface{}
	for _, version := range versions {
		if version.EntityID() == reqData.PostID {
			filteredVersions = append(filteredVersions, version)
		}
	}

	u.Logger().Info("Versions after filtering", "count", len(filteredVersions), "post_id", reqData.PostID)

	// Convert versions to serializable format
	versionList := []map[string]any{}
	for _, version := range filteredVersions {
		versionList = append(versionList, map[string]any{
			"id":         version.ID(),
			"content":    version.Content(),
			"created_at": version.GetCreatedAt(),
		})
	}

	return api.SuccessWithData("Versions loaded successfully", map[string]any{
		"versioning_enabled": true,
		"versions":           versionList,
	}).ToString()
}

func (u *ui) handleLoadVersionDetail(w http.ResponseWriter, r *http.Request) string {
	ctx := r.Context()

	var reqData struct {
		VersionID string `json:"version_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&reqData); err != nil {
		return api.Error("Invalid request body").ToString()
	}

	if reqData.VersionID == "" {
		return api.Error("Version ID is required").ToString()
	}

	blogStore := u.Store()
	if blogStore == nil {
		return api.Error("Blog store not available").ToString()
	}

	version, err := blogStore.VersioningFindByID(ctx, reqData.VersionID)
	if err != nil || version == nil {
		return api.Error("Version not found").ToString()
	}

	var versionData map[string]string
	if err := json.Unmarshal([]byte(version.Content()), &versionData); err != nil {
		return api.Error("Invalid version data").ToString()
	}

	// Attribute label mapping for UI display
	attributeLabels := map[string]string{
		"title":            "Title",
		"slug":             "Slug",
		"content":          "Content",
		"summary":          "Summary",
		"status":           "Status",
		"featured":         "Featured",
		"image_url":        "Image URL",
		"memo":             "Memo",
		"canonical_url":    "Canonical URL",
		"meta_description": "Meta Description",
		"meta_keywords":    "Meta Keywords",
		"meta_robots":      "Meta Robots",
		"published_at":     "Published At",
		"author_id":        "Author ID",
		"metas":            "Metadata",
		"id":               "ID",
	}

	attributeList := []map[string]any{}
	for key, value := range versionData {
		label, ok := attributeLabels[key]
		if !ok {
			label = key
		}
		attributeList = append(attributeList, map[string]any{
			"key":   key,
			"label": label,
			"value": value,
		})
	}

	return api.SuccessWithData("Version detail loaded", map[string]any{
		"attributes": attributeList,
		"created_at": version.GetCreatedAt(),
	}).ToString()
}

func (u *ui) handleRestoreVersionAttributes(w http.ResponseWriter, r *http.Request) string {
	ctx := r.Context()

	if r.Method != http.MethodPost {
		return api.Error("Method not allowed").ToString()
	}

	var reqData struct {
		PostID     string   `json:"post_id"`
		VersionID  string   `json:"version_id"`
		Attributes []string `json:"attributes"`
	}

	if err := json.NewDecoder(r.Body).Decode(&reqData); err != nil {
		return api.Error("Invalid request body").ToString()
	}

	if reqData.PostID == "" {
		return api.Error("Post ID is required").ToString()
	}

	if reqData.VersionID == "" {
		return api.Error("Version ID is required").ToString()
	}

	if len(reqData.Attributes) == 0 {
		return api.Error("At least one attribute must be selected").ToString()
	}

	blogStore := u.Store()
	if blogStore == nil {
		return api.Error("Blog store not available").ToString()
	}

	// Get the version to restore from
	version, err := blogStore.VersioningFindByID(ctx, reqData.VersionID)
	if err != nil || version == nil {
		return api.Error("Version not found").ToString()
	}

	// Get the current post
	post, err := blogStore.PostFindByID(ctx, reqData.PostID)
	if err != nil || post == nil {
		return api.Error("Post not found").ToString()
	}

	// Parse version content
	versionData := map[string]string{}
	if err := json.Unmarshal([]byte(version.Content()), &versionData); err != nil {
		return api.Error("Invalid version data").ToString()
	}

	// Apply selected attributes
	for _, attr := range reqData.Attributes {
		if val, ok := versionData[attr]; ok {
			post.Set(attr, val)
		}
	}

	// Update the post
	if err := blogStore.PostUpdate(ctx, post); err != nil {
		u.Logger().Error("Error updating post from version attributes", "error", err.Error())
		return api.Error("Error restoring attributes").ToString()
	}

	// Create a new version for the restoration
	if err := createPostVersioning(context.Background(), blogStore, post); err != nil {
		u.Logger().Error("Error creating post versioning after restore", "error", err.Error())
	}

	return api.Success("Selected attributes restored successfully").ToString()
}
