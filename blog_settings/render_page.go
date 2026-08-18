package blog_settings

import (
	"net/http"

	"github.com/dracory/blogadmin/shared"
	"github.com/dracory/cdn"
	"github.com/dracory/hb"
)

func (u *ui) renderPage(w http.ResponseWriter, r *http.Request) string {
	htmlContent, err := settingsFiles.ReadFile("settings.html")
	if err != nil {
		u.Logger().Error("Failed to read settings HTML", "error", err)
		return hb.Div().HTML("Error loading settings component").ToHTML()
	}

	jsContent, err := settingsFiles.ReadFile("settings.js")
	if err != nil {
		u.Logger().Error("Failed to read settings JS", "error", err)
		return hb.Div().HTML("Error loading settings component").ToHTML()
	}

	vueCDN := hb.Script("").Src(cdn.VueJs_3_5_32())

	linksHelper := shared.NewLinksFromRequest(r)
	initScript := hb.Script(`
		window.blogSettingsReturnUrl = '` + linksHelper.PostManager(nil) + `';
		const urlBlogSettingsFetchData = '` + linksHelper.BlogSettings(nil) + `?action=` + actionFetchData + `';
		const urlBlogSettingsSubmit = '` + linksHelper.BlogSettings(nil) + `?action=` + actionSubmit + `';
	`)

	htmlTemplate := hb.Wrap().HTML(string(htmlContent))
	componentScript := hb.Script(string(jsContent))

	vueContainer := hb.Div().
		Child(vueCDN).
		Child(htmlTemplate).
		Child(initScript).
		Child(componentScript)

	breadcrumbs := shared.Breadcrumbs([]shared.Breadcrumb{
		{Name: "Dashboard", URL: shared.AdminHomeURL(r)},
		{Name: "Blog", URL: shared.URLR(r, shared.CONTROLLER_DASHBOARD, nil)},
		{Name: "Settings", URL: shared.URLR(r, shared.CONTROLLER_BLOG_SETTINGS, nil)},
	})

	heading := hb.Heading1().HTML("Blog Settings")

	buttonBack := hb.Hyperlink().
		Class("btn btn-secondary ms-3").
		HTML("Back to Blog").
		Href(linksHelper.Home(nil))

	cardBody := hb.Div().
		Class("card-body").
		Child(vueContainer)

	card := hb.Div().
		Class("card shadow-sm").
		Child(hb.Div().
			Class("card-header d-flex justify-content-between align-items-center").
			Child(hb.Heading4().Class("mb-0").HTML("General Settings"))).
		Child(cardBody)

	page := hb.Div().
		Class("container py-4 min-vh-100").
		Child(breadcrumbs).
		Child(hb.Div().
			Class("d-flex align-items-center mb-4").
			Child(heading).
			Child(buttonBack)).
		Child(card)

	return u.Layout(w, r, "Settings | Blog", page.ToHTML(), struct {
		Styles     []string
		StyleURLs  []string
		Scripts    []string
		ScriptURLs []string
	}{
		ScriptURLs: []string{
			cdn.Sweetalert2_11(),
		},
	})
}
