package main

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/dracory/blogstore"
	"github.com/dromara/carbon/v2"
)

// seedDB populates the store with sample data so the admin panel has
// enough rows to exercise pagination, filtering, and sorting.
//
// Creates:
//   - 2 taxonomies (category, tag)
//   - 10 categories
//   - 20 tags
//   - 60 posts (mixed statuses, varied authors/dates)
func seedDB(store blogstore.StoreInterface, logger *slog.Logger) {
	ctx := context.Background()

	seedTaxonomies(ctx, store, logger)
	seedCategories(ctx, store, logger)
	seedTags(ctx, store, logger)
	seedPosts(ctx, store, logger)

	logger.Info("seedDB complete",
		"posts", 120,
		"categories", 60,
		"tags", 80,
	)
}

func seedTaxonomies(ctx context.Context, store blogstore.StoreInterface, logger *slog.Logger) {
	// Category taxonomy
	catTax := blogstore.NewTaxonomy().
		SetName("Category").
		SetSlug(blogstore.TAXONOMY_CATEGORY).
		SetDescription("Post categories")
	if err := store.TaxonomyCreate(ctx, catTax); err != nil {
		logger.Error("seedTaxonomies: failed to create category taxonomy", "error", err)
	}

	// Tag taxonomy
	tagTax := blogstore.NewTaxonomy().
		SetName("Tag").
		SetSlug(blogstore.TAXONOMY_TAG).
		SetDescription("Post tags")
	if err := store.TaxonomyCreate(ctx, tagTax); err != nil {
		logger.Error("seedTaxonomies: failed to create tag taxonomy", "error", err)
	}
}

func seedCategories(ctx context.Context, store blogstore.StoreInterface, logger *slog.Logger) {
	catTax, err := store.TaxonomyFindBySlug(ctx, blogstore.TAXONOMY_CATEGORY)
	if err != nil || catTax == nil {
		logger.Error("seedCategories: category taxonomy not found", "error", err)
		return
	}

	for i := 1; i <= 60; i++ {
		cat := blogstore.NewTerm().
			SetTaxonomyID(catTax.GetID()).
			SetName(fmt.Sprintf("Category %02d", i)).
			SetSlug(fmt.Sprintf("category-%02d", i)).
			SetDescription(fmt.Sprintf("Description for category %02d", i)).
			SetSequence(i)
		if err := store.TermCreate(ctx, cat); err != nil {
			logger.Error("seedCategories: failed to create category", "index", i, "error", err)
		}
	}
}

func seedTags(ctx context.Context, store blogstore.StoreInterface, logger *slog.Logger) {
	tagTax, err := store.TaxonomyFindBySlug(ctx, blogstore.TAXONOMY_TAG)
	if err != nil || tagTax == nil {
		logger.Error("seedTags: tag taxonomy not found", "error", err)
		return
	}

	for i := 1; i <= 80; i++ {
		tag := blogstore.NewTerm().
			SetTaxonomyID(tagTax.GetID()).
			SetName(fmt.Sprintf("Tag %02d", i)).
			SetSlug(fmt.Sprintf("tag-%02d", i)).
			SetSequence(i)
		if err := store.TermCreate(ctx, tag); err != nil {
			logger.Error("seedTags: failed to create tag", "index", i, "error", err)
		}
	}
}

func seedPosts(ctx context.Context, store blogstore.StoreInterface, logger *slog.Logger) {
	statuses := []string{
		blogstore.POST_STATUS_PUBLISHED,
		blogstore.POST_STATUS_PUBLISHED,
		blogstore.POST_STATUS_PUBLISHED,
		blogstore.POST_STATUS_DRAFT,
		blogstore.POST_STATUS_TRASH,
	}

	for i := 1; i <= 120; i++ {
		status := statuses[(i-1)%len(statuses)]
		post := blogstore.NewPost().
			SetTitle(fmt.Sprintf("Sample Post %02d — Exploring Ideas in Writing", i)).
			SetSlug(fmt.Sprintf("sample-post-%02d", i)).
			SetSummary(fmt.Sprintf("A summary for sample post %02d. This post explores various aspects of the topic and provides insights for readers.", i)).
			SetContent(fmt.Sprintf("# Sample Post %02d\n\nThis is the content of sample post number %d.\n\n## Introduction\n\nLorem ipsum dolor sit amet, consectetur adipiscing elit.\n\n## Main Section\n\nSed do eiusmod tempor incididunt ut labore et dolore magna aliqua.\n\n## Conclusion\n\nUt enim ad minim veniam, quis nostrud exercitation.", i, i)).
			SetStatus(status).
			SetEditor(blogstore.POST_EDITOR_MARKDOWN).
			SetPublishedAt(carbon.Now().SubDays(i).ToDateTimeString())

		if err := store.PostCreate(ctx, post); err != nil {
			logger.Error("seedPosts: failed to create post", "index", i, "error", err)
		}
	}
}
