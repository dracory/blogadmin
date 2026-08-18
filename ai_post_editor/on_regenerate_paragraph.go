package ai_post_editor

import (
	"log/slog"

	"github.com/dracory/blogai"
	"github.com/dracory/api"
	"github.com/dracory/req"
	"github.com/spf13/cast"
)

func (u *ui) onRegenerateParagraph(data pageData) string {
	sectionType := req.GetStringTrimmed(data.Request, "section_type")
	sectionIndexStr := req.GetStringTrimmed(data.Request, "section_index")
	paragraphIndexStr := req.GetStringTrimmed(data.Request, "paragraph_index")

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

	sectionIndex := cast.ToInt(sectionIndexStr)
	paragraphIndex := cast.ToInt(paragraphIndexStr)

	newParagraph, err := agent.RegenerateParagraph(llmEngine, data.BlogAiPost, sectionType, sectionIndex, paragraphIndex)
	if err != nil {
		u.Logger().Error("BlogAi. Post Editor. Regenerate Paragraph. Failed", slog.String("error", err.Error()))
		return api.Error("Failed to regenerate paragraph: " + err.Error()).ToString()
	}

	switch sectionType {
	case "introduction":
		if paragraphIndex == len(data.BlogAiPost.Introduction.Paragraphs) {
			data.BlogAiPost.Introduction.Paragraphs = append(data.BlogAiPost.Introduction.Paragraphs, newParagraph)
		} else {
			data.BlogAiPost.Introduction.Paragraphs[paragraphIndex] = newParagraph
		}
	case "conclusion":
		if paragraphIndex == len(data.BlogAiPost.Conclusion.Paragraphs) {
			data.BlogAiPost.Conclusion.Paragraphs = append(data.BlogAiPost.Conclusion.Paragraphs, newParagraph)
		} else {
			data.BlogAiPost.Conclusion.Paragraphs[paragraphIndex] = newParagraph
		}
	case "section":
		if paragraphIndex == len(data.BlogAiPost.Sections[sectionIndex].Paragraphs) {
			data.BlogAiPost.Sections[sectionIndex].Paragraphs = append(data.BlogAiPost.Sections[sectionIndex].Paragraphs, newParagraph)
		} else {
			data.BlogAiPost.Sections[sectionIndex].Paragraphs[paragraphIndex] = newParagraph
		}
	}

	data.Record.SetPayload(data.BlogAiPost.ToJSON())
	if err := u.CustomStore().RecordUpdate(data.Record); err != nil {
		return api.Error("Failed to save updated blog post: " + err.Error()).ToString()
	}

	return api.SuccessWithData("Paragraph regenerated successfully", map[string]any{"paragraph": newParagraph}).ToString()
}
