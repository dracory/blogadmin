package e2e

import (
	"strings"
	"testing"
	"time"

	"github.com/chromedp/chromedp"
)

func TestTagManagerLoadsSeededTags(t *testing.T) {
	body := waitList(t, newTab(t), adminURL("tag-manager"))
	if !strings.Contains(body, seedTagName) {
		t.Fatalf("expected seeded %q in tag manager", seedTagName)
	}
}

// TestTagCreateFlow verifies the "Add Tag" button opens the create
// modal, and that submitting it adds a tag that appears in the list.
func TestTagCreateFlow(t *testing.T) {
	ctx := newTab(t)
	name := "E2E Tag " + time.Now().Format("150405")

	err := chromedp.Run(ctx,
		chromedp.Navigate(adminURL("tag-manager")),
		chromedp.WaitReady(".list-group .list-group-item:nth-of-type(2)", chromedp.ByQuery),
	)
	if err != nil {
		t.Fatalf("tag manager did not load: %v", err)
	}

	err = chromedp.Run(ctx,
		chromedp.Click(`button.btn.btn-primary.btn-sm`, chromedp.ByQuery),
		chromedp.WaitReady(`input[placeholder="Tag name"]`, chromedp.ByQuery),
		chromedp.SendKeys(`input[placeholder="Tag name"]`, name, chromedp.ByQuery),
		chromedp.Click(`.modal-footer button.btn-primary`, chromedp.ByQuery),
		chromedp.Sleep(2*time.Second),
	)
	if err != nil {
		t.Fatalf("tag create modal interaction failed: %v", err)
	}

	var body string
	err = chromedp.Run(ctx,
		chromedp.Text("body", &body),
	)
	if err != nil {
		t.Fatalf("reading tag manager failed: %v", err)
	}
	if !strings.Contains(body, name) {
		t.Fatalf("expected newly created tag %q in manager", name)
	}
}
