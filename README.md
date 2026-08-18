# blogadmin

<img src="https://opengraph.githubassets.com/dracory/blogadmin" />

[![Tests Status](https://github.com/dracory/blogadmin/actions/workflows/tests.yml/badge.svg?branch=main)](https://github.com/dracory/blogadmin/actions/workflows/tests.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/dracory/blogadmin)](https://goreportcard.com/report/github.com/dracory/blogadmin)
[![PkgGoDev](https://pkg.go.dev/badge/github.com/dracory/blogadmin)](https://pkg.go.dev/github.com/dracory/blogadmin)

## License

This project is licensed under the GNU Affero General Public License v3.0 (AGPL-3.0). You can find a copy of the license at [https://www.gnu.org/licenses/agpl-3.0.en.html](https://www.gnu.org/licenses/agpl-3.0.txt)

For commercial use, please use my [contact page](https://lesichkov.co.uk/contact) to obtain a commercial license.

## Introduction

Admin interface for [`github.com/dracory/blogstore`](https://github.com/dracory/blogstore).
Provides a ready-to-use admin panel for managing blog posts, categories,
tags, SEO, media, post versions, and AI-powered content generation.

Modeled after [`github.com/dracory/shopadmin`](https://github.com/dracory/shopadmin)
— same folder-per-controller pattern, same `UiConfig`/`UiBase` conventions.

## Features

- **Post management** — create, update, delete, list with AJAX
- **Category management** — create, update, delete, drag-and-drop reorder
- **Tag management** — create, update, delete
- **Post versioning** — automatic snapshots on every save, with selective
  attribute restoration
- **Media management** — upload, reorder, delete post images
- **SEO management** — slug, canonical URL, meta description, meta keywords,
  meta robots, old slugs
- **Blog settings** — blog-level configuration via AJAX
- **AI tools** — title generator, post generator, post editor with
  section/paragraph regeneration, block-based content editor
- **Multi-filter tags** — stack multiple filter conditions as removable
  badge tags; filter state is shareable via URL
- **Custom layouts** — bring your own layout via `FuncLayout`
- **Bootstrap + Vue CDN** — default UI works out of the box

## Installation

```bash
go get github.com/dracory/blogadmin
```

## Quick Start

```go
package main

import (
    "log/slog"
    "net/http"
    "os"

    "github.com/dracory/blogadmin"
    "github.com/dracory/blogstore"
    "github.com/dracory/customstore"
    "github.com/dracory/settingstore"
)

func main() {
    store, err := blogstore.NewStore(blogstore.NewStoreOptions{
        DB:                 yourDB,
        PostTableName:      "blog_post",
        AutomigrateEnabled: true,
        VersioningEnabled:  true,
        TaxonomyEnabled:    true,
    })
    if err != nil {
        log.Fatal(err)
    }

    customStore, _ := customstore.NewStore(customstore.NewStoreOptions{
        DB:                 yourDB,
        TableName:          "custom_record",
        AutomigrateEnabled: true,
    })

    settingStore, _ := settingstore.NewStore(settingstore.NewStoreOptions{
        DB:                 yourDB,
        SettingTableName:   "setting",
        AutomigrateEnabled: true,
    })

    admin, err := blogadmin.New(blogadmin.AdminOptions{
        Store:        store,
        Logger:       slog.New(slog.NewTextHandler(os.Stderr, nil)),
        CustomStore:  customStore,
        SettingStore: settingStore,
        AdminHomeURL: "/admin",
        BlogAdminURL: "/admin/blog",
    })
    if err != nil {
        log.Fatal(err)
    }

    http.Handle("/admin/blog", http.HandlerFunc(admin.Handle))
    http.ListenAndServe(":8080", nil)
}
```

See [`example/`](example/) for a complete runnable server with
in-memory SQLite and seed data.

## Integration with a Router

`blogadmin.AdminInterface` exposes `Handle(w, r)`, which is an
`http.HandlerFunc`-compatible method. Wire it into any router that
accepts standard `http.Handler`:

```go
// stdlib
mux.Handle("/admin/blog", http.HandlerFunc(admin.Handle))

// github.com/dracory/rtr
route := rtr.NewRoute().
    SetName("Admin > Blog").
    SetPath("/admin/blog").
    SetHTMLHandler(admin.Handle)
```

## AI Features

AI controllers (title generator, post generator, post editor) require
an LLM factory. Provide one via `LlmFactory`:

```go
admin, _ := blogadmin.New(blogadmin.AdminOptions{
    Store:        store,
    Logger:       logger,
    CustomStore:  customStore,
    SettingStore: settingStore,
    LlmFactory: func() (llm.LlmInterface, error) {
        return llm.NewLlm(llm.LlmOptions{
            Provider: llm.PROVIDER_OPENAI,
            ApiKey:   os.Getenv("OPENAI_API_KEY"),
            Model:    "gpt-4o",
        })
    },
})
```

If `LlmFactory` is nil, AI controllers return an error to the user
instead of panicking. `CustomStore` and `SettingStore` are also
required for AI controllers — nil means they return an error.

## Custom Layout

By default, blogadmin renders a bare-bones HTML page with Bootstrap and
Vue from CDN. To embed the admin inside your own layout (branding, menus,
etc.), provide `FuncLayout`:

```go
admin, _ := blogadmin.New(blogadmin.AdminOptions{
    Store:  store,
    Logger: logger,
    FuncLayout: func(title, body string, opts struct {
        Styles     []string
        StyleURLs  []string
        Scripts    []string
        ScriptURLs []string
    }) string {
        return myLayout(title, body, opts)
    },
})
```

The anonymous struct matches `shopadmin` exactly, so you can reuse
your existing shopadmin layout function.

## Authentication

Provide an `AuthUserID` function to gate access. If it returns `""`, the
request is redirected to `AdminHomeURL`:

```go
admin, _ := blogadmin.New(blogadmin.AdminOptions{
    Store:      store,
    Logger:     logger,
    AuthUserID: func(r *http.Request) string {
        // return authenticated user ID, or ""
    },
})
```

## Testing

```bash
go test ./...
```

Tests use an in-memory SQLite database via `modernc.org/sqlite` — no
external services required.
