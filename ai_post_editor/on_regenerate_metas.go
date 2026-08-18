package ai_post_editor

import (
	"strings"

	"github.com/dracory/blogadmin/blogai"
	"github.com/dracory/api"
)

func (u *ui) onRegenerateMetas(data pageData) string {
	agent := blogai.NewBlogWriterAgent(u.Logger())
	if agent == nil {
		return api.Error("failed to initialize LLM engine").ToString()
	}

	llmEngine, err := u.LlmEngine()
	if err != nil {
		return api.Error("failed to initialize LLM engine: " + err.Error()).ToString()
	}
	if llmEngine == nil {
		return api.Error("failed to initialize LLM engine").ToString()
	}

	metaTitle, metaDescription, metaKeywords, err := agent.GenerateMetas(llmEngine, data.BlogAiPost)
	if err != nil {
		return api.Error("Failed to regenerate meta information: " + err.Error()).ToString()
	}

	data.BlogAiPost.MetaTitle = metaTitle
	data.BlogAiPost.MetaDescription = metaDescription
	data.BlogAiPost.Keywords = strings.Split(metaKeywords, ",")
	data.Record.SetPayload(data.BlogAiPost.ToJSON())
	if err := u.CustomStore().RecordUpdate(data.Record); err != nil {
		return api.Error("Failed to save updated blog post: " + err.Error()).ToString()
	}

	return api.SuccessWithData("Meta information regenerated successfully", map[string]any{
		"metaTitle":       metaTitle,
		"metaDescription": metaDescription,
		"metaKeywords":    metaKeywords,
	}).ToString()
}
