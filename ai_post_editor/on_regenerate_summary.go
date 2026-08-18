package ai_post_editor

import (
	"github.com/dracory/blogadmin/blogai"
	"github.com/dracory/api"
)

func (u *ui) onRegenerateSummary(data pageData) string {
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

	summary, err := agent.GenerateSummary(llmEngine, data.BlogAiPost)
	if err != nil {
		return api.Error("Failed to regenerate summary: " + err.Error()).ToString()
	}

	data.BlogAiPost.Summary = summary
	data.Record.SetPayload(data.BlogAiPost.ToJSON())
	if err := u.CustomStore().RecordUpdate(data.Record); err != nil {
		return api.Error("Failed to save updated blog post: " + err.Error()).ToString()
	}

	return api.SuccessWithData("Summary regenerated successfully", map[string]any{"summary": summary}).ToString()
}
