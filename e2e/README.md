# e2e

End-to-end tests that drive the blogadmin panel in a real headless Chrome
via [chromedp](https://github.com/chromedp/chromedp) (CDP). Unlike go-rod,
chromedp has no `leakless.exe` helper, so it avoids the Windows Defender
false positive entirely.

## How it works

- `TestMain` mounts the admin on `httptest.NewServer` (in-process, random
  port) backed by a seeded in-memory SQLite blogstore via
  `testutils.InitStores` — no binary build needed. Taxonomies, categories,
  tags, and posts are seeded; AI features stay disabled (default).
- One shared browser is launched via `chromedp.NewExecAllocator` for the
  whole run; each test opens a fresh tab with a 30s deadline (`newTab`).
- Chrome binary resolution order: `CHROME_BIN` env → Windows Playwright
  `chromium_headless_shell` → WSL-side Windows shell → Linux Playwright
  shell → chromedp's default lookup (Chrome on PATH / install dirs).

## Run

```bash
go test -v ./e2e/     # or: task e2e
```

Note: on WSL, run tests with the Windows Go toolchain (`go.exe`) —
Windows-side Chrome binds 127.0.0.1, which WSL cannot reach (no
WSL→Windows localhost forwarding). The Linux Playwright shell also
requires system libs (libnspr4 etc.) that may be missing.

## Coverage

Every non-AI registered controller has at least one test:

- `dashboard` — default controller smoke test
- `post-manager` — seeded rows + create-modal flow
- `post-update`, `post-delete`
- `category-manager` — seeded rows + create-modal flow
- `tag-manager` — seeded rows + create-modal flow
- `blog-settings`

AI controllers are not covered — they require an LLM factory and are
disabled by default (`AIEnabled: false`).

## Gotchas learned

- Prefer `chromedp.WaitReady` over `chromedp.WaitVisible`: on some pages
  `WaitVisible` polls forever even though the DOM is fully loaded.
- Routing is query-param based: `?controller=post-manager|category-manager|...`
  (see `shared/consts.go`).
- Manager/list pages render via Vue after an AJAX fetch — wait for the
  second `tbody tr` / `.list-group-item` before asserting (v-if/v-else
  and v-for mean the data rows only exist after Vue mounts).
- `v-model` is compiled away by Vue — it does NOT appear in the rendered
  DOM. Match form inputs by `type` or `placeholder`, never by the
  `v-model` attribute.
- The post manager's "New Post" button shares `.btn.btn-primary` with
  the filter button — disambiguate with `:not(.btn-sm)`.
- Server-rendered pages (`dashboard`, `post-delete`) need only
  `WaitReady("body")` / a known element — no AJAX wait needed.
- chromedp v0.23 rewrote the API (generic `Run[T]`); this suite pins the
  stable v0.14 line with matching `cdproto`/`go-json-experiment` versions.
