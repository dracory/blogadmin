// Package blogadmin provides a standalone blog admin interface following
// the folder-per-controller pattern. Each controller is in its own
// subfolder and handles its own views and AJAX data.
//
// This module is modeled on github.com/dracory/shopadmin.
package blogadmin

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/dracory/blogadmin/ai_post_content_update"
	"github.com/dracory/blogadmin/ai_post_editor"
	"github.com/dracory/blogadmin/ai_post_generator"
	"github.com/dracory/blogadmin/ai_test"
	"github.com/dracory/blogadmin/ai_title_generator"
	"github.com/dracory/blogadmin/ai_tools"
	"github.com/dracory/blogadmin/blog_settings"
	"github.com/dracory/blogadmin/category_manager"
	"github.com/dracory/blogadmin/dashboard"
	"github.com/dracory/blogadmin/post_create"
	"github.com/dracory/blogadmin/post_delete"
	"github.com/dracory/blogadmin/post_manager"
	"github.com/dracory/blogadmin/post_update"
	"github.com/dracory/blogadmin/shared"
	"github.com/dracory/blogadmin/tag_manager"
	"github.com/dracory/blogstore"
	"github.com/dracory/customstore"
	"github.com/dracory/llm"
	"github.com/dracory/req"
	"github.com/dracory/settingstore"
)

// AdminOptions contains all dependencies and configuration for the blog admin.
//
// Store and Logger are required. CustomStore, SettingStore, and LlmFactory
// are required only for the AI controllers; if nil, AI controllers return
// an error to the user instead of panicking.
//
// FuncLayout is an optional function to render the admin interface inside
// your own layout (branding, menus, etc.). If nil, a default bare-bones
// HTML page is used (Bootstrap + Vue CDN). Uses anonymous struct to match
// shopadmin exactly, so consumers can reuse their shopadmin layout
// function for blogadmin.
type AdminOptions struct {
	// Store is the blogstore.StoreInterface (required)
	Store blogstore.StoreInterface

	// Logger is required
	Logger *slog.Logger

	// CustomStore is required for AI controllers (ai_title_generator,
	// ai_post_generator, ai_post_editor). Nil means AI controllers
	// that need it return an error to the user.
	CustomStore customstore.StoreInterface

	// SettingStore is required for ai_title_generator. Nil means the
	// title generator returns an error when reading settings.
	SettingStore settingstore.StoreInterface

	// LlmFactory creates an LLM engine instance. Required for all AI
	// controllers. Nil means AI controllers return an error to the user.
	LlmFactory shared.LlmFactoryFunc

	// FuncLayout is an optional function to render the admin interface
	// inside your own layout (branding, menus, etc.). It receives the
	// request and response writer so the host project can access request
	// context (auth user, locale, etc.) when rendering the layout.
	FuncLayout func(w http.ResponseWriter, r *http.Request, title string, body string, options struct {
		Styles     []string
		StyleURLs  []string
		Scripts    []string
		ScriptURLs []string
	}) string

	// AdminHomeURL is the URL for the admin home page (default: "/admin")
	AdminHomeURL string

	// BlogAdminURL is the base URL for blog admin (default: "/admin/blog")
	BlogAdminURL string

	// FileManagerURL is the URL for the file manager (optional)
	FileManagerURL string

	// AuthUserID returns the authenticated user ID from the request.
	// If it returns "", the user is treated as unauthenticated.
	AuthUserID func(r *http.Request) string
}

// AdminInterface defines the interface for the blog admin
type AdminInterface interface {
	Handle(w http.ResponseWriter, r *http.Request)
}

// admin implements AdminInterface
type admin struct {
	store        blogstore.StoreInterface
	logger       *slog.Logger
	customStore  customstore.StoreInterface
	settingStore settingstore.StoreInterface
	llmFactory   shared.LlmFactoryFunc
	funcLayout   func(w http.ResponseWriter, r *http.Request, title string, body string, options struct {
		Styles     []string
		StyleURLs  []string
		Scripts    []string
		ScriptURLs []string
	}) string
	adminHomeURL   string
	blogAdminURL   string
	fileManagerURL string
	authUserID     func(r *http.Request) string
	routes         map[string]func(w http.ResponseWriter, r *http.Request)
}

// New creates a new blog admin instance.
// Returns ErrStoreRequired if Store is nil, ErrLoggerRequired if Logger is nil.
func New(opts AdminOptions) (AdminInterface, error) {
	if opts.Store == nil {
		return nil, ErrStoreRequired
	}
	if opts.Logger == nil {
		return nil, ErrLoggerRequired
	}

	// Set defaults
	if opts.AdminHomeURL == "" {
		opts.AdminHomeURL = "/admin"
	}
	if opts.BlogAdminURL == "" {
		opts.BlogAdminURL = "/admin/blog"
	}

	a := &admin{
		store:          opts.Store,
		logger:         opts.Logger,
		customStore:    opts.CustomStore,
		settingStore:   opts.SettingStore,
		llmFactory:     opts.LlmFactory,
		funcLayout:     opts.FuncLayout,
		adminHomeURL:   opts.AdminHomeURL,
		blogAdminURL:   opts.BlogAdminURL,
		fileManagerURL: opts.FileManagerURL,
		authUserID:     opts.AuthUserID,
	}

	// Build routes once at construction time
	a.routes = a.buildRoutes()

	return a, nil
}

// Handle processes all blog admin requests.
// Config values are injected into the request context (following the
// shopadmin pattern). Route lookup is map-based.
func (a *admin) Handle(w http.ResponseWriter, r *http.Request) {
	// Check authentication
	if a.authUserID != nil && a.authUserID(r) == "" {
		http.Redirect(w, r, a.adminHomeURL, http.StatusSeeOther)
		return
	}

	// Inject config into request context (like shopadmin)
	ctx := context.WithValue(r.Context(), shared.KeyEndpoint, r.URL.Path)
	ctx = context.WithValue(ctx, shared.KeyAdminHomeURL, a.adminHomeURL)
	ctx = context.WithValue(ctx, shared.KeyBlogAdminURL, a.blogAdminURL)
	ctx = context.WithValue(ctx, shared.KeyFileManagerURL, a.fileManagerURL)
	r = r.WithContext(ctx)

	// Map-based route lookup
	controller := req.GetStringTrimmed(r, "controller")
	if controller == "" {
		controller = shared.CONTROLLER_DASHBOARD
	}

	handler, ok := a.routes[controller]
	if !ok {
		handler = a.routes[shared.CONTROLLER_DASHBOARD]
	}

	handler(w, r)
}

// buildRoutes creates the handler dispatch map once at construction time.
func (a *admin) buildRoutes() map[string]func(w http.ResponseWriter, r *http.Request) {
	uiConfig := shared.UiConfig{
		Store:        a.store,
		Logger:       a.logger,
		CustomStore:  a.customStore,
		SettingStore: a.settingStore,
		LlmFactory:   a.llmFactory,
		Layout:       a.render,
	}

	return map[string]func(w http.ResponseWriter, r *http.Request){
		shared.CONTROLLER_DASHBOARD:    func(w http.ResponseWriter, r *http.Request) { dashboard.UI(uiConfig).Dashboard(w, r) },
		shared.CONTROLLER_POST_MANAGER: func(w http.ResponseWriter, r *http.Request) { post_manager.UI(uiConfig).PostManager(w, r) },
		shared.CONTROLLER_POST_CREATE:  func(w http.ResponseWriter, r *http.Request) { post_create.UI(uiConfig).PostCreate(w, r) },
		shared.CONTROLLER_POST_UPDATE: func(w http.ResponseWriter, r *http.Request) {
			post_update.UI(uiConfig, a.fileManagerURL).PostUpdate(w, r)
		},
		shared.CONTROLLER_POST_DELETE:        func(w http.ResponseWriter, r *http.Request) { post_delete.UI(uiConfig).PostDelete(w, r) },
		shared.CONTROLLER_CATEGORY_MANAGER:   func(w http.ResponseWriter, r *http.Request) { category_manager.UI(uiConfig).CategoryManager(w, r) },
		shared.CONTROLLER_TAG_MANAGER:        func(w http.ResponseWriter, r *http.Request) { tag_manager.UI(uiConfig).TagManager(w, r) },
		shared.CONTROLLER_BLOG_SETTINGS:      func(w http.ResponseWriter, r *http.Request) { blog_settings.UI(uiConfig).BlogSettings(w, r) },
		shared.CONTROLLER_AI_TOOLS:           func(w http.ResponseWriter, r *http.Request) { ai_tools.UI(uiConfig).AiTools(w, r) },
		shared.CONTROLLER_AI_TEST:            func(w http.ResponseWriter, r *http.Request) { ai_test.UI(uiConfig).AiTest(w, r) },
		shared.CONTROLLER_AI_POST_GENERATOR:  func(w http.ResponseWriter, r *http.Request) { ai_post_generator.UI(uiConfig).AiPostGenerator(w, r) },
		shared.CONTROLLER_AI_TITLE_GENERATOR: func(w http.ResponseWriter, r *http.Request) { ai_title_generator.UI(uiConfig).AiTitleGenerator(w, r) },
		shared.CONTROLLER_AI_POST_EDITOR:     func(w http.ResponseWriter, r *http.Request) { ai_post_editor.UI(uiConfig).AiPostEditor(w, r) },
		shared.CONTROLLER_AI_POST_CONTENT_UPDATE: func(w http.ResponseWriter, r *http.Request) {
			ai_post_content_update.UI(uiConfig).AiPostContentUpdate(w, r)
		},
	}
}

// render wraps content in the layout. If FuncLayout is provided and
// returns non-empty, it is used; otherwise the default shared.Layout
// is used (following the shopadmin pattern).
//
// When FuncLayout is set, the default shared.Layout is NOT computed
// (avoids wasted work).
func (a *admin) render(w http.ResponseWriter, r *http.Request, webpageTitle, webpageHtml string, options struct {
	Styles     []string
	StyleURLs  []string
	Scripts    []string
	ScriptURLs []string
}) string {
	// If a custom layout is provided, try it first
	if a.funcLayout != nil {
		custom := a.funcLayout(w, r, webpageTitle, webpageHtml, options)
		if custom != "" {
			if w != nil {
				w.Header().Set("Content-Type", "text/html; charset=utf-8")
				_, _ = w.Write([]byte(custom))
				return ""
			}
			return custom
		}
	}

	webpage := shared.Layout(w, r, webpageTitle, webpageHtml, options)

	if w != nil {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(webpage))
		return ""
	}

	return webpage
}

// Compile-time assertion that llm is used (AI controllers import it via shared)
var _ llm.LlmInterface
