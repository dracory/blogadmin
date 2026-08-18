package ai_title_generator

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/dracory/blogadmin/shared"
	"github.com/dracory/api"
)

func (u *ui) handleSettingsSubmit(r *http.Request) string {
	var reqBody struct {
		Action    string `json:"action"`
		BlogTopic string `json:"blog_topic"`
	}
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		return api.Error("Invalid request body").ToString()
	}

	blogTopic := strings.TrimSpace(reqBody.BlogTopic)
	if blogTopic == "" {
		return api.Error("Blog topic is required").ToString()
	}

	store := u.SettingStore()
	if store == nil {
		return api.Error("Setting store is not configured").ToString()
	}

	if err := store.Set(r.Context(), SETTING_KEY_BLOG_TOPIC, blogTopic); err != nil {
		u.Logger().Error("AI title generator settings: failed to save blog topic", "error", err.Error())
		return api.Error("Failed to save blog topic. Please try again later.").ToString()
	}

	if reqBody.Action == "save_close" {
		return api.SuccessWithData("Settings saved successfully", map[string]any{
			"redirect_url": shared.NewLinksFromRequest(r).AiTitleGenerator(nil),
		}).ToString()
	}

	return api.Success("Settings saved successfully").ToString()
}
