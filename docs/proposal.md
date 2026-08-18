# blogadmin — Standalone Module Proposal

## Goal

Extract the `blogadmin` package from `coursethread.com/pkg/blogadmin` into a
standalone, portable Go module (`github.com/dracory/blogadmin`) that has no
dependencies on the host project's `internal/` packages.

This follows the same pattern established by `github.com/dracory/shopadmin`.

## Motivation

The original `blogadmin` package was tightly coupled to the host project via:

- `project/internal/app` — the host's application container
- `project/internal/helpers` — flash messages, URL helpers
- `project/internal/layouts` — admin layout rendering
- `project/internal/links` — URL builders
- `project/internal/rules` — business rules
- `project/internal/config` — configuration access
- `project/internal/testutils` — test infrastructure

This made it impossible to reuse `blogadmin` in other projects without
copying the entire host project. Extracting it as a standalone module
enables:

1. **Reusability** — any Go project can import `blogadmin` and get a
   full blog admin panel
2. **Independent versioning** — `blogadmin` can be released and
   versioned independently of the host project
3. **Cleaner architecture** — the module defines its own interfaces
   (`StoreInterface`, `LlmFactoryFunc`, `LayoutFunction`) rather than
   depending on concrete host implementations
4. **Testability** — the module ships its own `testutils` package with
   in-memory SQLite helpers

## Architecture

### Module structure

```
blogadmin/
├── blogadmin.go              # New() entry point, route dispatch
├── errors.go                 # Sentinel errors
├── blogai/                   # AI agents (writer, title generator, proofreader)
├── go.mod
├── shared/                   # Shared types, URL builders, layout
│   ├── consts.go
│   ├── ui_config.go
│   ├── ui_base.go
│   ├── ui_interface.go
│   ├── layout.go
│   ├── links.go
│   ├── URL.go
│   ├── header.go
│   ├── breadcrumbs.go
│   ├── error_alert.go
│   └── functions.go
├── dashboard/                # Dashboard controller
├── post_manager/             # Post list controller
├── post_create/              # Post create controller
├── post_update/              # Post update controller (largest)
├── post_delete/              # Post delete controller
├── category_manager/         # Category manager controller
├── tag_manager/              # Tag manager controller
├── blog_settings/            # Blog settings controller
├── ai_tools/                 # AI tools index
├── ai_test/                  # AI connectivity test
├── ai_post_generator/        # AI post generator
├── ai_title_generator/       # AI title generator
├── ai_post_editor/           # AI post editor
├── ai_post_content_update/   # AI block-based content editor
├── testutils/                # In-memory test helpers
├── example/                  # Runnable example server
└── docs/
    └── proposal.md           # This file
```

### Dependency injection

The module uses dependency injection via `AdminOptions`:

| Field          | Type                          | Required | Used by                          |
|----------------|-------------------------------|----------|----------------------------------|
| Store          | `blogstore.StoreInterface`    | Yes      | All controllers                  |
| Logger         | `*slog.Logger`                | Yes      | All controllers                  |
| CustomStore    | `customstore.StoreInterface`  | No       | AI controllers                   |
| SettingStore   | `settingstore.StoreInterface` | No       | AI title generator               |
| LlmFactory     | `shared.LlmFactoryFunc`       | No       | All AI controllers               |
| FuncLayout     | `func(...) string`            | No       | Layout rendering                 |
| AdminHomeURL   | `string`                      | No       | Breadcrumbs (default: `/admin`)  |
| BlogAdminURL   | `string`                      | No       | URL building (default: `/admin/blog`) |
| FileManagerURL | `string`                      | No       | Post update media tab            |
| AuthUserID     | `func(*http.Request) string`  | No       | Authentication gate              |

### Controller pattern

Each controller follows the same pattern:

1. **`UiInterface`** — exported interface with the controller's HTTP
   handler method (e.g., `PostUpdate(w, r)`)
2. **`ui` struct** — embeds `shared.UiBase` for access to stores,
   logger, LLM engine, and layout
3. **`UI(config)` constructor** — takes a `shared.UiConfig`, returns
   the `UiInterface`
4. **`Handler(w, r) string`** — internal method that processes the
   request and returns HTML (or writes directly to `w` for AJAX)
5. **View renderers** — one function per tab/view
6. **AJAX handlers** — one function per `action` parameter

### URL building

URLs are built via `shared.NewLinksFromRequest(r)` which reads the
`BlogAdminURL` from the request context (injected by `admin.Handle`).
This allows the host project to mount `blogadmin` at any URL prefix.

### Layout rendering

If `FuncLayout` is provided, it is used to render the admin inside the
host project's layout (branding, menus, etc.). If not, a default
bare-bones HTML page is used (Bootstrap + Vue CDN).

### Server-side rendering vs Vue

The original `blogadmin` had server-side rendering methods (e.g.,
`table_post_list.go` in `post_manager`) that relied on
`links.Website()` from the host project. These were dropped in favor
of Vue.js-based rendering, which is self-contained and does not need
host-project URLs.

## Migration path

1. Host project replaces `project/pkg/blogadmin` imports with
   `github.com/dracory/blogadmin`
2. Host project replaces `NewBlogAdminController(app)` calls with
   `blogadmin.New(blogadmin.AdminOptions{...})`
3. Host project provides a `FuncLayout` that wraps content in its own
   admin layout
4. Host project provides `LlmFactory` if AI features are needed

## Dependencies

| Module                          | Purpose                          |
|---------------------------------|----------------------------------|
| `github.com/dracory/blogstore`  | Blog data persistence            |
| `github.com/dracory/customstore`| Custom records (AI posts)        |
| `github.com/dracory/settingstore`| Settings (blog topic)           |
| `github.com/dracory/llm`        | LLM engine interface             |
| `github.com/dracory/hb`         | HTML builder                     |
| `github.com/dracory/bs`         | Bootstrap components             |
| `github.com/dracory/cdn`        | CDN URLs                         |
| `github.com/dracory/req`        | Request helpers                  |
| `github.com/dracory/wf`         | Workflow pipeline                |
| `github.com/dracory/neat`       | Database ORM                     |
| `github.com/dracory/versionstore`| Post versioning                 |
| `github.com/dracory/api`        | JSON API responses               |
| `github.com/dracory/uid`        | ID generation                    |
| `github.com/dracory/str`        | String helpers                   |
| `github.com/dracory/cast`       | Type casting                     |
| `github.com/dromara/carbon/v2`  | Date/time helpers                |
| `github.com/samber/lo`          | Slice/map utilities              |
| `github.com/flosch/pongo2/v6`   | Template engine (post editor)    |
| `modernc.org/sqlite`            | Pure-Go SQLite (tests/example)   |
