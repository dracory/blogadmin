# blogadmin example

A self-contained example server that boots `blogadmin` on an in-memory
SQLite database. No external services required.

## Run

```bash
go run ./example
```

Then open http://localhost:8080/ in your browser.

## What it does

- Creates an in-memory SQLite database (reset on every restart)
- Initializes `blogstore`, `customstore`, and `settingstore` with
  auto-migration
- Seeds 60 sample posts, 10 categories, and 20 tags
- Mounts the blog admin panel at `/admin/blog`
- Serves a landing page at `/` that links into the admin

## AI features

AI controllers (title generator, post generator, post editor) require
an LLM factory. This example leaves `LlmFactory` as `nil`, so AI
controllers return an error to the user instead of making API calls.
To enable AI features, provide a real `LlmFactoryFunc`:

```go
admin, err := blogadmin.New(blogadmin.AdminOptions{
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
    // ...
})
```

## Persistence

To use a file-based database instead of in-memory, change `dbFile`:

```go
const dbFile = "blogadmin_example.db"
```

Data will persist across restarts.
