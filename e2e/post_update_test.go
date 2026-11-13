package e2e

import (
	"strings"
	"testing"

	"github.com/chromedp/chromedp"
)

// TestPostUpdateRenders verifies the post update controller renders the
// editor page for a seeded post.
func TestPostUpdateRenders(t *testing.T) {
	ctx := newTab(t)
	url := adminURL("post-update") + "&post_id=" + seedPostID

	var body string
	err := chromedp.Run(ctx,
		chromedp.Navigate(url),
		chromedp.WaitReady("body"),
		chromedp.Text("body", &body),
	)
	if err != nil {
		t.Fatalf("post update page failed: %v", err)
	}
	if strings.Contains(body, "Post not found") {
		t.Fatalf("seeded post %q not found by update controller", seedPostID)
	}
	if strings.TrimSpace(body) == "" {
		t.Fatal("expected non-empty post update page")
	}
}
