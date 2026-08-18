package shared

import "net/http"

// BlogAdminURL returns the blog admin base URL from request context
func BlogAdminURL(r *http.Request) string {
	value, ok := r.Context().Value(KeyBlogAdminURL).(string)
	if !ok {
		return ""
	}
	return value
}
