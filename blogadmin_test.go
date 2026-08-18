package blogadmin

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/dracory/blogadmin/testutils"
	_ "modernc.org/sqlite"
)

func TestNew_ValidOptions(t *testing.T) {
	store, customStore, settingStore, err := testutils.InitStores(":memory:")
	if err != nil {
		t.Fatalf("Failed to init stores: %v", err)
	}

	options := AdminOptions{
		Store:          store,
		Logger:         slog.New(slog.NewTextHandler(os.Stderr, nil)),
		CustomStore:    customStore,
		SettingStore:   settingStore,
		AdminHomeURL:   "/admin",
		BlogAdminURL:   "/admin/blog",
		FileManagerURL: "/admin/files",
	}
	a, err := New(options)
	if err != nil {
		t.Fatalf("Failed to create admin: %v", err)
	}
	if a == nil {
		t.Errorf("Expected admin to be created, got nil")
	}
}

func TestNew_MissingStore(t *testing.T) {
	options := AdminOptions{
		Logger: slog.New(slog.NewTextHandler(os.Stderr, nil)),
	}
	a, err := New(options)
	if err == nil {
		t.Errorf("Expected error when store is missing")
	}
	if a != nil {
		t.Errorf("Expected nil admin when store is missing")
	}
	if !strings.Contains(err.Error(), ErrStoreRequired.Error()) {
		t.Errorf("Expected error to contain '%s', got '%s'", ErrStoreRequired.Error(), err.Error())
	}
}

func TestNew_MissingLogger(t *testing.T) {
	store, _, _, err := testutils.InitStores(":memory:")
	if err != nil {
		t.Fatalf("Failed to init stores: %v", err)
	}

	options := AdminOptions{
		Store: store,
	}
	a, err := New(options)
	if err == nil {
		t.Errorf("Expected error when logger is missing")
	}
	if a != nil {
		t.Errorf("Expected nil admin when logger is missing")
	}
	if !strings.Contains(err.Error(), ErrLoggerRequired.Error()) {
		t.Errorf("Expected error to contain '%s', got '%s'", ErrLoggerRequired.Error(), err.Error())
	}
}

func TestNew_Defaults(t *testing.T) {
	store, _, _, err := testutils.InitStores(":memory:")
	if err != nil {
		t.Fatalf("Failed to init stores: %v", err)
	}

	options := AdminOptions{
		Store:  store,
		Logger: slog.New(slog.NewTextHandler(os.Stderr, nil)),
	}
	a, err := New(options)
	if err != nil {
		t.Fatalf("Failed to create admin: %v", err)
	}

	// Verify defaults were set
	admin := a.(*admin)
	if admin.adminHomeURL != "/admin" {
		t.Errorf("Expected default adminHomeURL '/admin', got '%s'", admin.adminHomeURL)
	}
	if admin.blogAdminURL != "/admin/blog" {
		t.Errorf("Expected default blogAdminURL '/admin/blog', got '%s'", admin.blogAdminURL)
	}
}

func TestHandle_DashboardController(t *testing.T) {
	store, _, _, err := testutils.InitStores(":memory:")
	if err != nil {
		t.Fatalf("Failed to init stores: %v", err)
	}

	a, err := New(AdminOptions{
		Store:  store,
		Logger: slog.New(slog.NewTextHandler(os.Stderr, nil)),
	})
	if err != nil {
		t.Fatalf("Failed to create admin: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/admin/blog", nil)
	rr := httptest.NewRecorder()

	a.Handle(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rr.Code)
	}

	body := rr.Body.String()
	if !strings.Contains(body, "Dashboard") {
		t.Errorf("Expected body to contain 'Dashboard', got: %s", body[:min(200, len(body))])
	}
}

func TestHandle_PostManagerController(t *testing.T) {
	store, _, _, err := testutils.InitStores(":memory:")
	if err != nil {
		t.Fatalf("Failed to init stores: %v", err)
	}

	a, err := New(AdminOptions{
		Store:  store,
		Logger: slog.New(slog.NewTextHandler(os.Stderr, nil)),
	})
	if err != nil {
		t.Fatalf("Failed to create admin: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/admin/blog?controller=post-manager", nil)
	rr := httptest.NewRecorder()

	a.Handle(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rr.Code)
	}

	body := rr.Body.String()
	if !strings.Contains(body, "Post Manager") {
		t.Errorf("Expected body to contain 'Post Manager'")
	}
}

func TestHandle_CategoryManagerController(t *testing.T) {
	store, _, _, err := testutils.InitStores(":memory:")
	if err != nil {
		t.Fatalf("Failed to init stores: %v", err)
	}

	a, err := New(AdminOptions{
		Store:  store,
		Logger: slog.New(slog.NewTextHandler(os.Stderr, nil)),
	})
	if err != nil {
		t.Fatalf("Failed to create admin: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/admin/blog?controller=category-manager", nil)
	rr := httptest.NewRecorder()

	a.Handle(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rr.Code)
	}

	body := rr.Body.String()
	if !strings.Contains(body, "Category") {
		t.Errorf("Expected body to contain 'Category'")
	}
}

func TestHandle_TagManagerController(t *testing.T) {
	store, _, _, err := testutils.InitStores(":memory:")
	if err != nil {
		t.Fatalf("Failed to init stores: %v", err)
	}

	a, err := New(AdminOptions{
		Store:  store,
		Logger: slog.New(slog.NewTextHandler(os.Stderr, nil)),
	})
	if err != nil {
		t.Fatalf("Failed to create admin: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/admin/blog?controller=tag-manager", nil)
	rr := httptest.NewRecorder()

	a.Handle(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rr.Code)
	}

	body := rr.Body.String()
	if !strings.Contains(body, "Tag") {
		t.Errorf("Expected body to contain 'Tag'")
	}
}

func TestHandle_AiToolsController(t *testing.T) {
	store, customStore, settingStore, err := testutils.InitStores(":memory:")
	if err != nil {
		t.Fatalf("Failed to init stores: %v", err)
	}

	a, err := New(AdminOptions{
		Store:        store,
		Logger:       slog.New(slog.NewTextHandler(os.Stderr, nil)),
		CustomStore:  customStore,
		SettingStore: settingStore,
	})
	if err != nil {
		t.Fatalf("Failed to create admin: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/admin/blog?controller=ai-tools", nil)
	rr := httptest.NewRecorder()

	a.Handle(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rr.Code)
	}

	body := rr.Body.String()
	if !strings.Contains(body, "BlogAI") {
		t.Errorf("Expected body to contain 'BlogAI'")
	}
}

func TestHandle_UnknownControllerFallsBackToDashboard(t *testing.T) {
	store, _, _, err := testutils.InitStores(":memory:")
	if err != nil {
		t.Fatalf("Failed to init stores: %v", err)
	}

	a, err := New(AdminOptions{
		Store:  store,
		Logger: slog.New(slog.NewTextHandler(os.Stderr, nil)),
	})
	if err != nil {
		t.Fatalf("Failed to create admin: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/admin/blog?controller=nonexistent", nil)
	rr := httptest.NewRecorder()

	a.Handle(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rr.Code)
	}

	body := rr.Body.String()
	if !strings.Contains(body, "Dashboard") {
		t.Errorf("Expected fallback to dashboard with 'Dashboard'")
	}
}

func TestHandle_AuthUserIDRedirect(t *testing.T) {
	store, _, _, err := testutils.InitStores(":memory:")
	if err != nil {
		t.Fatalf("Failed to init stores: %v", err)
	}

	a, err := New(AdminOptions{
		Store:      store,
		Logger:     slog.New(slog.NewTextHandler(os.Stderr, nil)),
		AuthUserID: func(r *http.Request) string { return "" },
	})
	if err != nil {
		t.Fatalf("Failed to create admin: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/admin/blog", nil)
	rr := httptest.NewRecorder()

	a.Handle(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Errorf("Expected redirect status 303, got %d", rr.Code)
	}
}

func TestHandle_AuthUserIDPasses(t *testing.T) {
	store, _, _, err := testutils.InitStores(":memory:")
	if err != nil {
		t.Fatalf("Failed to init stores: %v", err)
	}

	a, err := New(AdminOptions{
		Store:      store,
		Logger:     slog.New(slog.NewTextHandler(os.Stderr, nil)),
		AuthUserID: func(r *http.Request) string { return "user-123" },
	})
	if err != nil {
		t.Fatalf("Failed to create admin: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/admin/blog", nil)
	rr := httptest.NewRecorder()

	a.Handle(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200 when authenticated, got %d", rr.Code)
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
