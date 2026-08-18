package testutils

import (
	"database/sql"
	"log/slog"
	"os"
	"strings"

	"github.com/dracory/blogstore"
	"github.com/dracory/customstore"
	"github.com/dracory/settingstore"
)

// InitStores creates in-memory blogstore, customstore, and settingstore
// instances for testing. The SQLite driver (modernc.org/sqlite) must be
// imported with a side-effect import in _test.go files of the consuming
// package, NOT here. This avoids double driver registration panics when
// a host project that already imports modernc.org/sqlite also imports
// this testutils package.
//
// Example (in your _test.go file):
//
//	import (
//	    "github.com/dracory/blogadmin/testutils"
//	    _ "modernc.org/sqlite"
//	)
//
//	store, customStore, settingStore, err := testutils.InitStores(":memory:")
func InitStores(filepath string) (blogstore.StoreInterface, customstore.StoreInterface, settingstore.StoreInterface, error) {
	db, err := initDB(filepath)
	if err != nil {
		return nil, nil, nil, err
	}

	blogStore, err := blogstore.NewStore(blogstore.NewStoreOptions{
		DB:                     db,
		PostTableName:          "blog_post",
		TaxonomyTableName:      "blog_taxonomy",
		TermTableName:          "blog_term",
		TermRelationTableName:  "blog_term_rel",
		MediaTableName:         "blog_media",
		VersioningEnabled:      true,
		VersioningTableName:    "blog_version",
		TaxonomyEnabled:        true,
		AutomigrateEnabled:     true,
	})
	if err != nil {
		return nil, nil, nil, err
	}

	customStore, err := customstore.NewStore(customstore.NewStoreOptions{
		DB:                 db,
		TableName:          "custom_record",
		AutomigrateEnabled: true,
		Logger:             slog.Default(),
	})
	if err != nil {
		return nil, nil, nil, err
	}

	settingStore, err := settingstore.NewStore(settingstore.NewStoreOptions{
		DB:                 db,
		SettingTableName:   "setting",
		AutomigrateEnabled: true,
	})
	if err != nil {
		return nil, nil, nil, err
	}

	return blogStore, customStore, settingStore, nil
}

// InitBlogStore creates only an in-memory blogstore for testing.
// Useful for tests that don't need the AI stores.
func InitBlogStore(filepath string) (blogstore.StoreInterface, error) {
	db, err := initDB(filepath)
	if err != nil {
		return nil, err
	}

	return blogstore.NewStore(blogstore.NewStoreOptions{
		DB:                     db,
		PostTableName:          "blog_post",
		TaxonomyTableName:      "blog_taxonomy",
		TermTableName:          "blog_term",
		TermRelationTableName:  "blog_term_rel",
		MediaTableName:         "blog_media",
		VersioningEnabled:      true,
		VersioningTableName:    "blog_version",
		TaxonomyEnabled:        true,
		AutomigrateEnabled:     true,
	})
}

func initDB(filepath string) (*sql.DB, error) {
	if filepath != ":memory:" {
		err := os.Remove(filepath)
		if err != nil && !strings.Contains(err.Error(), "no such file or directory") {
			return nil, err
		}
	}

	dsn := filepath + "?parseTime=true"
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}

	return db, nil
}
