package e2e

import (
	"strings"
	"testing"
	"time"

	"github.com/chromedp/chromedp"
)

func TestCategoryManagerLoadsSeededCategories(t *testing.T) {
	body := waitList(t, newTab(t), adminURL("category-manager"))
	if !strings.Contains(body, seedCategoryName) {
		t.Fatalf("expected seeded %q in category manager", seedCategoryName)
	}
}

// TestCategoryCreateFlow verifies the "Add Category" button opens the
// create modal, and that submitting it adds a category that appears in
// the list.
func TestCategoryCreateFlow(t *testing.T) {
	ctx := newTab(t)
	name := "E2E Category " + time.Now().Format("150405")

	err := chromedp.Run(ctx,
		chromedp.Navigate(adminURL("category-manager")),
		chromedp.WaitReady(".list-group .list-group-item:nth-of-type(2)", chromedp.ByQuery),
	)
	if err != nil {
		t.Fatalf("category manager did not load: %v", err)
	}

	err = chromedp.Run(ctx,
		chromedp.Click(`button.btn.btn-primary.btn-sm`, chromedp.ByQuery),
		chromedp.WaitReady(`input[placeholder="Category name"]`, chromedp.ByQuery),
		chromedp.SendKeys(`input[placeholder="Category name"]`, name, chromedp.ByQuery),
		chromedp.Click(`.modal-footer button.btn-primary`, chromedp.ByQuery),
		chromedp.Sleep(2*time.Second),
	)
	if err != nil {
		t.Fatalf("category create modal interaction failed: %v", err)
	}

	var body string
	err = chromedp.Run(ctx,
		chromedp.Text("body", &body),
	)
	if err != nil {
		t.Fatalf("reading category manager failed: %v", err)
	}
	if !strings.Contains(body, name) {
		t.Fatalf("expected newly created category %q in manager", name)
	}
}
