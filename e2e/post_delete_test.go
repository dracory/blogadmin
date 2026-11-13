package e2e

import (
	"testing"

	"github.com/chromedp/chromedp"
)

// TestPostDeleteShowsModal verifies the post delete controller renders
// its confirmation modal for a seeded post. The delete itself is not
// submitted — mutating state is covered by the create-flow test.
func TestPostDeleteShowsModal(t *testing.T) {
	ctx := newTab(t)
	url := adminURL("post-delete") + "&post_id=" + seedPostID

	err := chromedp.Run(ctx,
		chromedp.Navigate(url),
		chromedp.WaitReady("#ModalPostDelete", chromedp.ByID),
	)
	if err != nil {
		t.Fatalf("post delete modal did not render: %v", err)
	}
}
