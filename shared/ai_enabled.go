package shared

import "net/http"

// AIEnabled reports whether AI features are enabled for the current
// request. The value is injected into the request context by Handle()
// from AdminOptions.AIEnabled. When false, AI controllers are not
// registered and AI navigation links are hidden.
func AIEnabled(r *http.Request) bool {
	value, ok := r.Context().Value(KeyAIEnabled).(bool)
	return ok && value
}
