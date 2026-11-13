package e2e

import (
	"strings"
	"testing"

	"github.com/chromedp/chromedp"
)

// TestDashboardLoads verifies the default controller (dashboard) renders
// a non-empty page. The dashboard is server-rendered, so waiting for
// body is enough — no AJAX wait needed.
func TestDashboardLoads(t *testing.T) {
	ctx := newTab(t)
	var body string
	err := chromedp.Run(ctx,
		chromedp.Navigate(baseURL+adminPath),
		chromedp.WaitReady("body"),
		chromedp.Text("body", &body),
	)
	if err != nil {
		t.Fatalf("dashboard navigation failed: %v", err)
	}
	if strings.TrimSpace(body) == "" {
		t.Fatal("expected non-empty dashboard page")
	}
}
