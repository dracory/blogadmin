package dashboard

import (
	"context"
	"net/http"
	"strconv"

	"github.com/dracory/blogadmin/shared"
	"github.com/dracory/blogstore"
	"github.com/dracory/cdn"
	"github.com/dracory/hb"
	"github.com/dracory/uid"
)

// UiInterface defines the dashboard controller's UI interface
type UiInterface interface {
	shared.UiInterface
	Dashboard(w http.ResponseWriter, r *http.Request)
}

// ui implements UiInterface
type ui struct {
	shared.UiBase
}

// UI creates a new dashboard controller UI from the given config
func UI(config shared.UiConfig) UiInterface {
	return &ui{UiBase: shared.NewUiBase(config)}
}

// Dashboard handles the dashboard controller requests
func (u *ui) Dashboard(w http.ResponseWriter, r *http.Request) {
	html := u.Handler(w, r)
	if html != "" {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(html))
	}
}

// Handler processes the dashboard controller request and returns HTML
func (u *ui) Handler(w http.ResponseWriter, r *http.Request) string {
	data, errorMessage := u.prepareData(r)

	if errorMessage != "" {
		return shared.ErrorAlert(errorMessage)
	}

	content := u.page(r, data)

	return u.Layout(w, r, "Blog | Dashboard", content.ToHTML(), struct {
		Styles     []string
		StyleURLs  []string
		Scripts    []string
		ScriptURLs []string
	}{
		ScriptURLs: []string{
			cdn.Htmx_2_0_0(),
			cdn.Sweetalert2_10(),
		},
	})
}

func (u *ui) page(r *http.Request, data dashboardControllerData) hb.TagInterface {
	breadcrumbs := shared.Breadcrumbs([]shared.Breadcrumb{
		{
			Name: "Home",
			URL:  shared.AdminHomeURL(r),
		},
		{
			Name: "Blog",
			URL:  shared.URLR(r, shared.CONTROLLER_DASHBOARD, nil),
		},
		{
			Name: "Dashboard",
			URL:  "",
		},
	})

	// Add navigation tabs like CMS
	navTabs := u.navTabs(r, data)

	title := hb.Heading1().
		HTML("Blog. Dashboard")

	content := hb.Wrap().
		Child(navTabs).
		Child(hb.BR()).
		Child(u.dashboardCards(r, data))

	return hb.Div().
		Class("container").
		Class("py-4").
		Child(title).
		Child(breadcrumbs).
		Child(content)
}

func (u *ui) navTabs(r *http.Request, data dashboardControllerData) hb.TagInterface {
	linksHelper := shared.NewLinksFromRequest(r)

	children := []hb.TagInterface{
		hb.Hyperlink().
			Class("text-decoration-none").
			HTML("Dashboard").
			Href(linksHelper.Dashboard(nil)),
		hb.Hyperlink().
			Class("text-decoration-none").
			Child(
				hb.Wrap().
					Text("Posts ").
					Child(
						hb.Span().
							Class("badge bg-secondary ms-1").
							Text(strconv.FormatInt(data.postCount, 10)),
					),
			).
			Href(linksHelper.PostManager(nil)),
	}

	// Only show Categories/Tags tabs if taxonomy is enabled
	if data.taxonomyEnabled {
		children = append(children,
			hb.Hyperlink().
				Class("text-decoration-none").
				Child(
					hb.Wrap().
						Text("Categories ").
						Child(
							hb.Span().
								Class("badge bg-secondary ms-1").
								Text(strconv.FormatInt(data.categoryCount, 10)),
						),
				).
				Href(linksHelper.CategoryManager(nil)),
			hb.Hyperlink().
				Class("text-decoration-none").
				Child(
					hb.Wrap().
						Text("Tags ").
						Child(
							hb.Span().
								Class("badge bg-secondary ms-1").
								Text(strconv.FormatInt(data.tagCount, 10)),
						),
				).
				Href(linksHelper.TagManager(nil)),
		)
	}

	return hb.Div().
		Class("card mb-4").
		Child(
			hb.Div().
				Class("card-body d-flex justify-content-center gap-4").
				Children(children),
		)
}

func (u *ui) dashboardCards(r *http.Request, data dashboardControllerData) hb.TagInterface {
	cards := []hb.TagInterface{
		hb.Div().Class("col-md-4").Child(u.cardPosts(r, data)),
	}

	// Only show category/tag cards if taxonomy is enabled
	if data.taxonomyEnabled {
		cards = append(cards,
			hb.Div().Class("col-md-4").Child(u.cardCategories(r, data)),
			hb.Div().Class("col-md-4").Child(u.cardTags(r, data)),
		)
	}

	return hb.Div().
		Class("row").
		Children(cards)
}

func (u *ui) cardPosts(r *http.Request, data dashboardControllerData) hb.TagInterface {
	linksHelper := shared.NewLinksFromRequest(r)

	return hb.Hyperlink().
		Class("text-decoration-none").
		Href(linksHelper.PostManager(nil)).
		Child(
			hb.Div().
				Class("card mb-4").
				Style("background-color: #a8d5ba; border: none;").
				Child(
					hb.Div().
						Class("card-body d-flex justify-content-between align-items-center").
						Children([]hb.TagInterface{
							hb.Div().
								Children([]hb.TagInterface{
									hb.Heading3().Class("mb-0").Text(strconv.FormatInt(data.postCount, 10)).Style("color: #2c5f2d;"),
									hb.Paragraph().Class("mb-0").Text("Total Posts").Style("color: #2c5f2d;"),
								}),
							hb.I().Class("bi bi-file-text fs-1").Style("color: rgba(44, 95, 45, 0.3);"),
						}),
				).
				Child(
					hb.Div().
						Class("card-footer bg-transparent border-0 text-center pb-3").
						Child(
							hb.Span().
								Class("small fw-medium").
								Text("More info ").
								Style("color: #2c5f2d;").
								Child(hb.I().Class("bi bi-arrow-right-circle")),
						),
				),
		)
}

func (u *ui) cardCategories(r *http.Request, data dashboardControllerData) hb.TagInterface {
	linksHelper := shared.NewLinksFromRequest(r)

	return hb.Hyperlink().
		Class("text-decoration-none").
		Href(linksHelper.CategoryManager(nil)).
		Child(
			hb.Div().
				Class("card mb-4").
				Style("background-color: #c5d5f5; border: none;").
				Child(
					hb.Div().
						Class("card-body d-flex justify-content-between align-items-center").
						Children([]hb.TagInterface{
							hb.Div().
								Children([]hb.TagInterface{
									hb.Heading3().Class("mb-0").Text(strconv.FormatInt(data.categoryCount, 10)).Style("color: #1a3a6e;"),
									hb.Paragraph().Class("mb-0").Text("Categories").Style("color: #1a3a6e;"),
								}),
							hb.I().Class("bi bi-folder fs-1").Style("color: rgba(26, 58, 110, 0.3);"),
						}),
				).
				Child(
					hb.Div().
						Class("card-footer bg-transparent border-0 text-center pb-3").
						Child(
							hb.Span().
								Class("small fw-medium").
								Text("More info ").
								Style("color: #1a3a6e;").
								Child(hb.I().Class("bi bi-arrow-right-circle")),
						),
				),
		)
}

func (u *ui) cardTags(r *http.Request, data dashboardControllerData) hb.TagInterface {
	linksHelper := shared.NewLinksFromRequest(r)

	return hb.Hyperlink().
		Class("text-decoration-none").
		Href(linksHelper.TagManager(nil)).
		Child(
			hb.Div().
				Class("card mb-4").
				Style("background-color: #f5e6c8; border: none;").
				Child(
					hb.Div().
						Class("card-body d-flex justify-content-between align-items-center").
						Children([]hb.TagInterface{
							hb.Div().
								Children([]hb.TagInterface{
									hb.Heading3().Class("mb-0").Text(strconv.FormatInt(data.tagCount, 10)).Style("color: #8b6914;"),
									hb.Paragraph().Class("mb-0").Text("Tags").Style("color: #8b6914;"),
								}),
							hb.I().Class("bi bi-tags fs-1").Style("color: rgba(139, 105, 20, 0.3);"),
						}),
				).
				Child(
					hb.Div().
						Class("card-footer bg-transparent border-0 text-center pb-3").
						Child(
							hb.Span().
								Class("small fw-medium").
								Text("More info ").
								Style("color: #8b6914;").
								Child(hb.I().Class("bi bi-arrow-right-circle")),
						),
				),
		)
}

func (u *ui) prepareData(r *http.Request) (data dashboardControllerData, errorMessage string) {
	ctx := r.Context()

	blogStore := u.Store()
	if blogStore == nil {
		return data, "Blog store not available"
	}

	// Check if taxonomy is enabled by attempting to find a taxonomy
	_, err := blogStore.TaxonomyFindBySlug(ctx, blogstore.TAXONOMY_CATEGORY)
	if err != nil && err.Error() == "taxonomy is not enabled" {
		data.taxonomyEnabled = false
		data.taxonomyErrorMsg = "Categories and Tags are not available. To enable them, set TaxonomyEnabled to true in the blog store configuration."
	} else {
		data.taxonomyEnabled = true
		// Ensure taxonomies exist
		if err := u.ensureTaxonomies(ctx, blogStore); err != nil {
			return data, "Failed to ensure taxonomies: " + err.Error()
		}
	}

	// Count posts
	postCount, err := blogStore.PostCount(ctx, blogstore.PostQueryOptions{})
	if err != nil {
		u.Logger().Error("blog dashboard: error counting posts", "error", err)
	}
	data.postCount = postCount

	// Get category taxonomy count only if taxonomy is enabled
	if data.taxonomyEnabled {
		categoryTaxonomy, err := blogStore.TaxonomyFindBySlug(ctx, blogstore.TAXONOMY_CATEGORY)
		if err == nil && categoryTaxonomy != nil {
			categoryCount, err := blogStore.TermCount(ctx, blogstore.TermQueryOptions{
				TaxonomyID: categoryTaxonomy.GetID(),
			})
			if err != nil {
				u.Logger().Error("blog dashboard: error counting categories", "error", err)
			}
			data.categoryCount = categoryCount
		}

		// Get tag taxonomy count
		tagTaxonomy, err := blogStore.TaxonomyFindBySlug(ctx, blogstore.TAXONOMY_TAG)
		if err == nil && tagTaxonomy != nil {
			tagCount, err := blogStore.TermCount(ctx, blogstore.TermQueryOptions{
				TaxonomyID: tagTaxonomy.GetID(),
			})
			if err != nil {
				u.Logger().Error("blog dashboard: error counting tags", "error", err)
			}
			data.tagCount = tagCount
		}
	}

	return data, ""
}

func (u *ui) ensureTaxonomies(ctx context.Context, store blogstore.StoreInterface) error {
	// Check if category taxonomy exists
	_, err := store.TaxonomyFindBySlug(ctx, blogstore.TAXONOMY_CATEGORY)
	if err != nil {
		u.Logger().Info("Creating category taxonomy")
		taxonomy := blogstore.NewTaxonomy()
		taxonomy.SetID(uid.HumanUid()[:8])
		taxonomy.SetName("Category")
		taxonomy.SetSlug(blogstore.TAXONOMY_CATEGORY)
		taxonomy.SetDescription("Blog post categories")
		if err := store.TaxonomyCreate(ctx, taxonomy); err != nil {
			return err
		}
	}

	// Check if tag taxonomy exists
	_, err = store.TaxonomyFindBySlug(ctx, blogstore.TAXONOMY_TAG)
	if err != nil {
		u.Logger().Info("Creating tag taxonomy")
		taxonomy := blogstore.NewTaxonomy()
		taxonomy.SetID(uid.HumanUid()[:8])
		taxonomy.SetName("Tag")
		taxonomy.SetSlug(blogstore.TAXONOMY_TAG)
		taxonomy.SetDescription("Blog post tags")
		if err := store.TaxonomyCreate(ctx, taxonomy); err != nil {
			return err
		}
	}

	return nil
}

type dashboardControllerData struct {
	postCount        int64
	categoryCount    int64
	tagCount         int64
	taxonomyEnabled  bool
	taxonomyErrorMsg string
}
