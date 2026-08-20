# blogadmin example

A self-contained example server that boots `blogadmin` on an in-memory
SQLite database. No external services required.

## Run

```bash
go run ./example
```

Then open http://localhost:8080/ in your browser.

## Hot reload (recommended for development)

The repo includes a [`taskfile.yml`](../taskfile.yml) and [`.air.toml`](../.air.toml)
for hot-reload development with [Air](https://github.com/air-verse/air)
and [Task](https://taskfile.dev).

Install the tools once:

```bash
# Install task:  https://taskfile.dev/installation/
# Install air:
task air:install
```

Then start the example with hot reload:

```bash
task dev          # AI disabled (default)
task dev:ai       # AI enabled (requires OPENAI_API_KEY)
```

Or use Air directly:

```bash
air               # AI disabled
BLOGADMIN_AI_ENABLED=true air   # AI enabled
```

## What it does

- Creates an in-memory SQLite database (reset on every restart)
- Initializes `blogstore`, `customstore`, and `settingstore` with
  auto-migration
- Seeds 60 sample posts, 10 categories, and 20 tags
- Mounts the blog admin panel at `/admin/blog`
- Serves a landing page at `/` that links into the admin

## AI features

AI features are **opt-in** and disabled by default. To enable them,
set `BLOGADMIN_AI_ENABLED=true` (or `1`) and provide an OpenAI API key
via `OPENAI_API_KEY`:

```bash
BLOGADMIN_AI_ENABLED=true OPENAI_API_KEY=sk-... go run ./example
```

Optional: `OPENAI_MODEL` selects the model (default `gpt-4o`).

When enabled, the example wires up a real `LlmFactoryFunc` using
`llm.NewLLM` with `llm.ProviderOpenAI`, and passes `AIEnabled: true`
to `blogadmin.New`. When disabled, AI routes are not registered and
AI navigation links are hidden — no API key is required.

## Persistence

To use a file-based database instead of in-memory, change `dbFile`:

```go
const dbFile = "blogadmin_example.db"
```

Data will persist across restarts.
