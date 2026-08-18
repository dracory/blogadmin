package ai_post_content_update

import (
	"net/http"
	"strings"

	"github.com/dracory/blogadmin/shared"
	"github.com/dracory/cdn"
	"github.com/dracory/hb"
)

func (u *ui) renderPage(w http.ResponseWriter, r *http.Request) string {
	postID := strings.TrimSpace(r.URL.Query().Get("post_id"))
	if postID == "" {
		postID = r.PostFormValue("post_id")
	}
	if postID == "" {
		return shared.ErrorPopup("Post ID is required").ToHTML()
	}

	htmlContent, err := editorFiles.ReadFile("editor.html")
	if err != nil {
		u.Logger().Error("Failed to read editor HTML", "error", err)
		return hb.Div().HTML("Error loading editor").ToHTML()
	}

	jsContent, err := editorFiles.ReadFile("editor.js")
	if err != nil {
		u.Logger().Error("Failed to read editor JS", "error", err)
		return hb.Div().HTML("Error loading editor").ToHTML()
	}

	linksHelper := shared.NewLinksFromRequest(r)

	vueCDN := hb.Script("").Src(cdn.VueJs_3_5_32())

	initScript := hb.Script(`
		window.postEditorPostId = '` + postID + `';
		window.postEditorBackUrl = '` + linksHelper.PostUpdate(map[string]string{"post_id": postID}) + `';
		const urlPostEditorFetchData = '` + linksHelper.AiPostContentUpdate(map[string]string{"post_id": postID, "action": actionFetchData}) + `';
		const urlPostEditorRegenerate = '` + linksHelper.AiPostContentUpdate(map[string]string{"post_id": postID, "action": actionRegenerateBlock}) + `';
		const urlPostEditorSave = '` + linksHelper.AiPostContentUpdate(map[string]string{"post_id": postID, "action": actionSave}) + `';
	`)

	htmlTemplate := hb.Wrap().HTML(string(htmlContent))
	componentScript := hb.Script(string(jsContent))

	vueContainer := hb.Div().
		Child(vueCDN).
		Child(htmlTemplate).
		Child(initScript).
		Child(componentScript)

	return u.Layout(w, r, "Edit Post Content", vueContainer.ToHTML(), struct {
		Styles     []string
		StyleURLs  []string
		Scripts    []string
		ScriptURLs []string
	}{
		ScriptURLs: []string{
			cdn.Sweetalert2_11(),
			"https://cdn.jsdelivr.net/npm/sortablejs@1.15.0/Sortable.min.js",
		},
	})
}
