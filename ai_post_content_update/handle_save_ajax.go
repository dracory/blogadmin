package ai_post_content_update

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/dracory/blogadmin/shared"
	"github.com/dracory/api"
	"github.com/dracory/blogstore"
)

func (u *ui) handleSave(w http.ResponseWriter, r *http.Request) string {
	var reqBody struct {
		Action  string              `json:"action"`
		Title   string              `json:"title"`
		Summary string              `json:"summary"`
		Blocks  []map[string]string `json:"blocks"`
	}
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		return api.Error("Invalid request body").ToString()
	}

	postID := strings.TrimSpace(r.URL.Query().Get("post_id"))
	if postID == "" {
		postID = r.PostFormValue("post_id")
	}
	if postID == "" {
		return api.Error("Post ID is required").ToString()
	}

	if u.Store() == nil {
		return api.Error("Blog store is not configured").ToString()
	}

	post, err := u.Store().PostFindByID(r.Context(), postID)
	if err != nil {
		u.Logger().Error("AI content editor: failed to load post for save", "error", err.Error())
		return api.Error("Failed to load post").ToString()
	}
	if post == nil {
		return api.Error("Post not found").ToString()
	}

	// Reconstruct blocks from request
	blocks := make([]Block, 0, len(reqBody.Blocks))
	for _, b := range reqBody.Blocks {
		blocks = append(blocks, Block{
			ID:   b["id"],
			Type: BlockType(b["type"]),
			Text: b["text"],
		})
	}

	title := strings.TrimSpace(reqBody.Title)
	summary := strings.TrimSpace(reqBody.Summary)
	if title != "" {
		post.SetTitle(title)
	}
	if summary != "" {
		post.SetSummary(summary)
	}

	markdown := BlocksToMarkdown(blocks)
	post.SetContent(markdown)
	post.SetEditor(blogstore.POST_EDITOR_MARKDOWN)

	if err := u.Store().PostUpdate(r.Context(), post); err != nil {
		u.Logger().Error("AI content editor: failed to save post", "error", err.Error())
		return api.Error("Failed to save post. Please try again later.").ToString()
	}

	if reqBody.Action == "save_close" {
		return api.SuccessWithData("Post saved successfully", map[string]any{
			"redirect_url": shared.NewLinksFromRequest(r).PostManager(nil),
		}).ToString()
	}

	return api.Success("Changes applied successfully").ToString()
}
