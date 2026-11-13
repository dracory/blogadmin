// Package e2e contains end-to-end tests that drive the blogadmin panel
// with a real headless Chrome via chromedp (CDP). The admin is served
// in-process by httptest on a random port — no binary build needed.
package e2e

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/dracory/blogadmin"
	"github.com/dracory/blogadmin/testutils"
	"github.com/dracory/blogstore"
	_ "modernc.org/sqlite"
)

var (
	baseURL          string
	allocCtx         context.Context
	testStore        blogstore.StoreInterface
	seedPostID       string
	seedPostTitle    string
	seedCategoryName = "Category 01"
	seedTagName      = "Tag 01"
)

const adminPath = "/admin/blog"

func TestMain(m *testing.M) {
	// In-process admin server backed by seeded in-memory stores.
	store, _, _, err := testutils.InitStores(":memory:")
	if err != nil {
		panic("e2e: failed to init stores: " + err.Error())
	}
	testStore = store

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	seedDB(store, logger)

	// Grab a seeded post so view/update/delete tests have a stable
	// post_id to navigate to.
	posts, err := store.PostList(context.Background(), blogstore.PostQueryOptions{})
	if err != nil || len(posts) == 0 {
		panic("e2e: no seeded posts: " + fmt.Sprint(err))
	}
	seedPostID = posts[0].GetID()
	seedPostTitle = posts[0].GetTitle()

	panel, err := blogadmin.New(blogadmin.AdminOptions{
		Store:        store,
		Logger:       logger,
		AdminHomeURL: "/admin",
		BlogAdminURL: adminPath,
	})
	if err != nil {
		panic("e2e: failed to create blogadmin: " + err.Error())
	}

	mux := http.NewServeMux()
	mux.Handle(adminPath, http.HandlerFunc(panel.Handle))
	mux.Handle(adminPath+"/", http.HandlerFunc(panel.Handle))
	srv := httptest.NewServer(mux)
	baseURL = srv.URL

	// One shared browser process for the entire test run.
	opts := append(chromedp.DefaultExecAllocatorOptions[:], chromedp.Headless)
	if bin := chromeBin(); bin != "" {
		opts = append(opts, chromedp.ExecPath(bin))
	}
	var cancel context.CancelFunc
	allocCtx, cancel = chromedp.NewExecAllocator(context.Background(), opts...)

	code := m.Run()

	cancel()
	srv.Close()
	os.Exit(code)
}

// seedDB populates the store with taxonomies, categories, tags, and
// posts so the manager pages have rows to render. Mirrors
// example/seed.go but with fewer rows — e2e only needs a couple of
// table pages.
func seedDB(store blogstore.StoreInterface, logger *slog.Logger) {
	ctx := context.Background()

	// Taxonomies (required parents for category/tag terms)
	catTax := blogstore.NewTaxonomy().
		SetName("Category").
		SetSlug(blogstore.TAXONOMY_CATEGORY).
		SetDescription("Post categories")
	if err := store.TaxonomyCreate(ctx, catTax); err != nil {
		logger.Error("seedDB: failed to create category taxonomy", "error", err)
	}
	tagTax := blogstore.NewTaxonomy().
		SetName("Tag").
		SetSlug(blogstore.TAXONOMY_TAG).
		SetDescription("Post tags")
	if err := store.TaxonomyCreate(ctx, tagTax); err != nil {
		logger.Error("seedDB: failed to create tag taxonomy", "error", err)
	}

	catTaxFound, err := store.TaxonomyFindBySlug(ctx, blogstore.TAXONOMY_CATEGORY)
	if err != nil || catTaxFound == nil {
		logger.Error("seedDB: category taxonomy not found", "error", err)
		return
	}
	tagTaxFound, err := store.TaxonomyFindBySlug(ctx, blogstore.TAXONOMY_TAG)
	if err != nil || tagTaxFound == nil {
		logger.Error("seedDB: tag taxonomy not found", "error", err)
		return
	}

	for i := 1; i <= 10; i++ {
		cat := blogstore.NewTerm().
			SetTaxonomyID(catTaxFound.GetID()).
			SetName(fmt.Sprintf("Category %02d", i)).
			SetSlug(fmt.Sprintf("category-%02d", i)).
			SetDescription(fmt.Sprintf("Description for category %02d", i)).
			SetSequence(i)
		if err := store.TermCreate(ctx, cat); err != nil {
			logger.Error("seedDB: failed to create category", "index", i, "error", err)
		}
	}

	for i := 1; i <= 10; i++ {
		tag := blogstore.NewTerm().
			SetTaxonomyID(tagTaxFound.GetID()).
			SetName(fmt.Sprintf("Tag %02d", i)).
			SetSlug(fmt.Sprintf("tag-%02d", i)).
			SetSequence(i)
		if err := store.TermCreate(ctx, tag); err != nil {
			logger.Error("seedDB: failed to create tag", "index", i, "error", err)
		}
	}

	statuses := []string{
		blogstore.POST_STATUS_PUBLISHED,
		blogstore.POST_STATUS_PUBLISHED,
		blogstore.POST_STATUS_DRAFT,
	}
	for i := 1; i <= 20; i++ {
		post := blogstore.NewPost().
			SetTitle(fmt.Sprintf("E2E Seed Post %02d", i)).
			SetSlug(fmt.Sprintf("e2e-seed-post-%02d", i)).
			SetSummary(fmt.Sprintf("Summary for seed post %02d.", i)).
			SetContent(fmt.Sprintf("# Seed Post %02d\n\nBody of seed post %d.", i, i)).
			SetStatus(statuses[(i-1)%len(statuses)]).
			SetEditor(blogstore.POST_EDITOR_MARKDOWN)
		if err := store.PostCreate(ctx, post); err != nil {
			logger.Error("seedDB: failed to create post", "index", i, "error", err)
		}
	}
}

// newTab opens a fresh tab in the shared browser and returns its
// context. Each tab gets a 30s deadline; cleanup closes the tab.
func newTab(t *testing.T) context.Context {
	t.Helper()
	ctx, cancel := chromedp.NewContext(allocCtx)
	ctx, cancelTimeout := context.WithTimeout(ctx, 30*time.Second)
	t.Cleanup(func() { cancelTimeout(); cancel() })
	return ctx
}

// chromeBin resolves a headless Chromium binary. Order: CHROME_BIN env,
// Playwright's downloaded headless shell, empty (chromedp finds Chrome
// on PATH / standard install locations).
func chromeBin() string {
	if b := os.Getenv("CHROME_BIN"); b != "" {
		return b
	}
	patterns := []string{
		// Windows
		filepath.Join(os.Getenv("LOCALAPPDATA"),
			"ms-playwright/chromium_headless_shell-*/chrome-headless-shell-win64/chrome-headless-shell.exe"),
		// WSL: Windows-side Playwright shell via interop
		"/mnt/*/Users/*/AppData/Local/ms-playwright/chromium_headless_shell-*/chrome-headless-shell-win64/chrome-headless-shell.exe",
		// Linux
		filepath.Join(os.Getenv("HOME"),
			".cache/ms-playwright/chromium_headless_shell-*/chrome-headless-shell-linux64/chrome-headless-shell"),
	}
	for _, pattern := range patterns {
		if matches, _ := filepath.Glob(pattern); len(matches) > 0 {
			return matches[len(matches)-1]
		}
	}
	return ""
}
