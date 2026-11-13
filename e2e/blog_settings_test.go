package e2e

import (
	"strings"
	"testing"
)

// TestBlogSettingsRenders verifies the blog settings controller renders
// its form. The page is Vue-mounted; waiting for the topic textarea
// means the app has loaded.
func TestBlogSettingsRenders(t *testing.T) {
	body := waitSelector(t, newTab(t), adminURL("blog-settings"), "#blog_topic")
	if !strings.Contains(body, "Blog Topic") {
		t.Fatal("expected 'Blog Topic' field in blog settings page")
	}
}
