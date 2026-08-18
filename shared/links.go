package shared

import "net/http"

// Links provides URL helpers for blogadmin controllers.
// The base URL is read from request context (injected by Handle()),
// not hardcoded. This follows the shopadmin pattern.
type Links struct {
	baseURL string
}

// NewLinks creates a Links helper with the given base URL.
// If baseURL is empty, defaults to "/admin/blog".
func NewLinks(baseURL string) *Links {
	if baseURL == "" {
		baseURL = "/admin/blog"
	}
	return &Links{baseURL: baseURL}
}

// NewLinksFromRequest creates a Links helper using the blog admin URL
// from the request context.
func NewLinksFromRequest(r *http.Request) *Links {
	return NewLinks(BlogAdminURL(r))
}

// Home builds the URL for the home/post-manager controller
func (l *Links) Home(params map[string]string) string {
	return l.url(CONTROLLER_POST_MANAGER, params)
}

// PostCreate builds the URL for the post create controller
func (l *Links) PostCreate(params map[string]string) string {
	return l.url(CONTROLLER_POST_CREATE, params)
}

// PostDelete builds the URL for the post delete controller
func (l *Links) PostDelete(params map[string]string) string {
	return l.url(CONTROLLER_POST_DELETE, params)
}

// PostManager builds the URL for the post manager controller
func (l *Links) PostManager(params map[string]string) string {
	return l.url(CONTROLLER_POST_MANAGER, params)
}

// PostUpdate builds the URL for the post update controller
func (l *Links) PostUpdate(params map[string]string) string {
	return l.url(CONTROLLER_POST_UPDATE, params)
}

// PostUpdateV1 builds the URL for the post update v1 controller
func (l *Links) PostUpdateV1(params map[string]string) string {
	return l.url(CONTROLLER_POST_UPDATE_V1, params)
}

// BlogSettings builds the URL for the blog settings controller
func (l *Links) BlogSettings(params map[string]string) string {
	return l.url(CONTROLLER_BLOG_SETTINGS, params)
}

// AiTools builds the URL for the AI tools controller
func (l *Links) AiTools(params map[string]string) string {
	return l.url(CONTROLLER_AI_TOOLS, params)
}

// AiPostContentUpdate builds the URL for the AI post content update controller
func (l *Links) AiPostContentUpdate(params map[string]string) string {
	return l.url(CONTROLLER_AI_POST_CONTENT_UPDATE, params)
}

// AiPostGenerator builds the URL for the AI post generator controller
func (l *Links) AiPostGenerator(params map[string]string) string {
	return l.url(CONTROLLER_AI_POST_GENERATOR, params)
}

// AiTitleGenerator builds the URL for the AI title generator controller
func (l *Links) AiTitleGenerator(params map[string]string) string {
	return l.url(CONTROLLER_AI_TITLE_GENERATOR, params)
}

// AiPostEditor builds the URL for the AI post editor controller
func (l *Links) AiPostEditor(params map[string]string) string {
	return l.url(CONTROLLER_AI_POST_EDITOR, params)
}

// AiTest builds the URL for the AI test controller
func (l *Links) AiTest(params map[string]string) string {
	return l.url(CONTROLLER_AI_TEST, params)
}

// Dashboard builds the URL for the dashboard controller
func (l *Links) Dashboard(params map[string]string) string {
	return l.url(CONTROLLER_DASHBOARD, params)
}

// CategoryManager builds the URL for the category manager controller
func (l *Links) CategoryManager(params map[string]string) string {
	return l.url(CONTROLLER_CATEGORY_MANAGER, params)
}

// TagManager builds the URL for the tag manager controller
func (l *Links) TagManager(params map[string]string) string {
	return l.url(CONTROLLER_TAG_MANAGER, params)
}

// url builds a URL for the given controller. The params map is copied
// before mutation (does not modify caller's map).
func (l *Links) url(controller string, params map[string]string) string {
	return URL(l.baseURL, controller, params)
}
