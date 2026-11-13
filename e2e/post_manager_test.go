package e2e

import (
	"strings"
	"testing"
	"time"

	"github.com/chromedp/chromedp"
)

func TestPostManagerLoadsSeededPosts(t *testing.T) {
	body := waitTable(t, newTab(t), adminURL("post-manager"))
	if !strings.Contains(body, "E2E Seed Post") {
		t.Fatal("expected seeded 'E2E Seed Post' rows in post manager")
	}
}

// TestPostCreateFlow verifies the "New Post" button opens the create
// modal, and that submitting it creates a post that then appears in the
// manager table.
func TestPostCreateFlow(t *testing.T) {
	ctx := newTab(t)
	title := "E2E Created Post " + time.Now().Format("150405")

	// Waiting for the second rendered table row guarantees Vue has
	// mounted; the modal inputs are only bound after that.
	err := chromedp.Run(ctx,
		chromedp.Navigate(adminURL("post-manager")),
		chromedp.WaitReady("table tbody tr:nth-of-type(2)", chromedp.ByQuery),
	)
	if err != nil {
		t.Fatalf("post manager did not load: %v", err)
	}

	// Open the create modal and fill the form. The "New Post" button is
	// the only .btn-primary that is NOT .btn-sm (the filter button is).
	err = chromedp.Run(ctx,
		chromedp.Click(`button.btn.btn-primary:not(.btn-sm)`, chromedp.ByQuery),
		chromedp.WaitReady(`input[placeholder="Post title"]`, chromedp.ByQuery),
		chromedp.SendKeys(`input[placeholder="Post title"]`, title, chromedp.ByQuery),
		chromedp.Click(`.modal-footer button.btn-primary`, chromedp.ByQuery),
		chromedp.Sleep(2*time.Second),
	)
	if err != nil {
		t.Fatalf("create modal interaction failed: %v", err)
	}

	// The table refreshes after create; the new post sorts newest-first.
	var body string
	err = chromedp.Run(ctx,
		chromedp.Text("body", &body),
	)
	if err != nil {
		t.Fatalf("reading post manager failed: %v", err)
	}
	if !strings.Contains(body, title) {
		t.Fatalf("expected newly created post %q in manager", title)
	}
}
