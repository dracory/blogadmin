package shared

// Context keys for config values injected by Handle()
const KeyEndpoint = "endpoint"
const KeyAdminHomeURL = "admin_home_url"
const KeyBlogAdminURL = "blog_admin_url"
const KeyFileManagerURL = "file_manager_url"

// Controller names used in the ?controller= query parameter
const (
	CONTROLLER_HOME               = "home"
	CONTROLLER_POST_CREATE        = "post-create"
	CONTROLLER_POST_DELETE        = "post-delete"
	CONTROLLER_POST_MANAGER       = "post-manager"
	CONTROLLER_POST_UPDATE        = "post-update"
	CONTROLLER_POST_UPDATE_V1     = "post-update-v1"
	CONTROLLER_BLOG_SETTINGS      = "blog-settings"
	CONTROLLER_AI_TOOLS           = "ai-tools"
	CONTROLLER_AI_POST_CONTENT_UPDATE = "ai-post-content-update"
	CONTROLLER_AI_POST_GENERATOR  = "ai-post-generator"
	CONTROLLER_AI_TITLE_GENERATOR = "ai-title-generator"
	CONTROLLER_AI_POST_EDITOR     = "ai-post-editor"
	CONTROLLER_AI_TEST            = "ai-test"
	CONTROLLER_DASHBOARD          = "dashboard"
	CONTROLLER_CATEGORY_MANAGER   = "category-manager"
	CONTROLLER_TAG_MANAGER        = "tag-manager"
)

// CatchAll is the catch-all route suffix
const CatchAll = "/*"

// Error messages
const ERROR_STORE_IS_NIL = "store cannot be nil"
const ERROR_LOGGER_IS_NIL = "logger cannot be nil"
