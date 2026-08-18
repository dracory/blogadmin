package shared

import (
	"log/slog"
	"net/http"

	"github.com/dracory/blogstore"
	"github.com/dracory/customstore"
	"github.com/dracory/llm"
	"github.com/dracory/settingstore"
)

// LlmFactoryFunc creates an LLM engine instance. The host project provides
// this so blogadmin does not depend on any specific config/auth package.
// Called by the AI controllers (ai_tools, ai_test, ai_post_generator,
// ai_title_generator, ai_post_editor, ai_post_content_update, and the
// regenerate-image flow in post_update).
//
// If nil, AI controllers return an error to the user instead of panicking.
type LlmFactoryFunc func() (llm.LlmInterface, error)

// UiConfig holds the dependencies passed to subcontroller UI factories.
// The Layout function uses an anonymous struct for options to match
// cmsstore/admin and shopadmin exactly, allowing consumers to reuse
// their existing layout function for blogadmin.
//
// AI controllers require CustomStore, SettingStore, and LlmFactory.
// Core controllers only need Store and Logger. Nil AI dependencies are
// safe — AI controllers return an error to the user instead of panicking.
type UiConfig struct {
	Store        blogstore.StoreInterface
	Logger       *slog.Logger
	CustomStore  customstore.StoreInterface
	SettingStore settingstore.StoreInterface
	LlmFactory   LlmFactoryFunc
	Layout       func(w http.ResponseWriter, r *http.Request, webpageTitle, webpageHtml string, options struct {
		Styles     []string
		StyleURLs  []string
		Scripts    []string
		ScriptURLs []string
	}) string
}
