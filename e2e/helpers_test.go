package e2e

import (
	"context"
	"testing"

	"github.com/chromedp/chromedp"
)

// adminURL builds a controller URL, e.g. /admin/blog?controller=post-manager.
func adminURL(controller string) string {
	return baseURL + adminPath + "?controller=" + controller
}

// waitTable navigates to the post manager and waits until Vue has
// loaded data and rendered the results table. Returns the page's text
// content.
//
// Waits for the SECOND table row: the raw v-for template row may exist
// in the DOM before Vue mounts, so `table tr` alone can resolve while
// the table is still a template. A second row only exists after v-for
// has expanded with real data.
func waitTable(t *testing.T, ctx context.Context, url string) string {
	t.Helper()
	return waitSelector(t, ctx, url, "table tbody tr:nth-of-type(2)")
}

// waitList navigates to a list page (category/tag managers render
// .list-group-item divs, not table rows) and waits until Vue has
// rendered at least two real items. Returns the page's text content.
func waitList(t *testing.T, ctx context.Context, url string) string {
	t.Helper()
	return waitSelector(t, ctx, url, ".list-group .list-group-item:nth-of-type(2)")
}

// waitSelector navigates to url, waits for sel to be ready, and returns
// the body's text content.
func waitSelector(t *testing.T, ctx context.Context, url, sel string) string {
	t.Helper()
	var body string
	err := chromedp.Run(ctx,
		chromedp.Navigate(url),
		chromedp.WaitReady(sel, chromedp.ByQuery),
		chromedp.Text("body", &body),
	)
	if err != nil {
		t.Fatalf("page %s failed: %v", url, err)
	}
	return body
}
