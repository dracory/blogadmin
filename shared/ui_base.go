package shared

import (
	"log/slog"
	"net/http"

	"github.com/dracory/blogstore"
	"github.com/dracory/customstore"
	"github.com/dracory/llm"
	"github.com/dracory/settingstore"
)

// UiBase is a base struct that implements shared.UiInterface.
// Subcontroller ui structs can embed this to get the Store(),
// Logger(), CustomStore(), SettingStore(), LlmFactory(), and Layout()
// methods for free, following the shopadmin pattern.
type UiBase struct {
	StoreField        blogstore.StoreInterface
	LoggerField       *slog.Logger
	CustomStoreField  customstore.StoreInterface
	SettingStoreField settingstore.StoreInterface
	LlmFactoryField   LlmFactoryFunc
	LayoutField       func(w http.ResponseWriter, r *http.Request, webpageTitle, webpageHtml string, options struct {
		Styles     []string
		StyleURLs  []string
		Scripts    []string
		ScriptURLs []string
	}) string
}

func (u UiBase) Store() blogstore.StoreInterface { return u.StoreField }
func (u UiBase) Logger() *slog.Logger            { return u.LoggerField }
func (u UiBase) CustomStore() customstore.StoreInterface { return u.CustomStoreField }
func (u UiBase) SettingStore() settingstore.StoreInterface { return u.SettingStoreField }
func (u UiBase) LlmFactory() LlmFactoryFunc { return u.LlmFactoryField }

func (u UiBase) LlmEngine() (llm.LlmInterface, error) {
	if u.LlmFactoryField == nil {
		return nil, errLlmFactoryNil
	}
	return u.LlmFactoryField()
}

func (u UiBase) Layout(w http.ResponseWriter, r *http.Request, webpageTitle, webpageHtml string, options struct {
	Styles     []string
	StyleURLs  []string
	Scripts    []string
	ScriptURLs []string
}) string {
	return u.LayoutField(w, r, webpageTitle, webpageHtml, options)
}

// NewUiBase creates a UiBase from a UiConfig
func NewUiBase(config UiConfig) UiBase {
	return UiBase{
		StoreField:        config.Store,
		LoggerField:       config.Logger,
		CustomStoreField:  config.CustomStore,
		SettingStoreField: config.SettingStore,
		LlmFactoryField:   config.LlmFactory,
		LayoutField:       config.Layout,
	}
}
