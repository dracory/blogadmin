package ai_post_editor

import (
	"log/slog"

	"github.com/dracory/blogadmin/blogai"
	"github.com/dracory/api"
)

func (u *ui) onRegenerateImage(data pageData) string {
	agent := blogai.NewBlogWriterAgent(u.Logger())
	if agent == nil {
		return api.Error("failed to initialize LLM engine").ToString()
	}

	llmEngine, err := u.LlmEngine()
	if err != nil {
		return api.Error("failed to initialize LLM engine").ToString()
	}

	imageDataURL, err := agent.GenerateImage(llmEngine, data.BlogAiPost.Title, data.BlogAiPost.Summary)
	if err != nil {
		u.Logger().Error("BlogAi. Post Editor. Generate Image. Failed to generate image", slog.String("error", err.Error()))
		return api.Error("Failed to generate image").ToString()
	}

	data.BlogAiPost.Image = imageDataURL
	data.Record.SetPayload(data.BlogAiPost.ToJSON())
	if err := u.CustomStore().RecordUpdate(data.Record); err != nil {
		return api.Error("Failed to save updated blog post with image: " + err.Error()).ToString()
	}

	return api.SuccessWithData("Image generated successfully", map[string]any{"image": imageDataURL}).ToString()
}
