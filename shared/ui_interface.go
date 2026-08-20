package shared

import (
	"log/slog"
	"net/http"

	"github.com/dracory/blogstore"
	"github.com/dracory/customstore"
	"github.com/dracory/llm"
	"github.com/dracory/settingstore"
)

// UiInterface defines the methods every subcontroller UI must implement.
// This follows the shopadmin pattern.
//
// AI controllers additionally use CustomStore(), SettingStore(), and
// LlmEngine(); core controllers only use Store(), Logger(), and Layout().
// AIEnabled() reports whether AI features are enabled, so controllers
// can conditionally render AI navigation links.
type UiInterface interface {
	Store() blogstore.StoreInterface
	Logger() *slog.Logger
	CustomStore() customstore.StoreInterface
	SettingStore() settingstore.StoreInterface
	LlmEngine() (llm.LlmInterface, error)
	AIEnabled() bool

	Layout(w http.ResponseWriter, r *http.Request, webpageTitle, webpageHtml string, options struct {
		Styles     []string
		StyleURLs  []string
		Scripts    []string
		ScriptURLs []string
	}) string
}
