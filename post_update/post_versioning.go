package post_update

import (
	"context"
	"errors"

	"github.com/dracory/blogstore"
	"github.com/dracory/neat"
	"github.com/dracory/versionstore"
	"github.com/samber/lo"
)

// createPostVersioning creates a new version snapshot of the post if
// versioning is enabled and the content has changed since the last
// version. This replaces the project/internal/app dependency with the
// injected blogstore.StoreInterface.
func createPostVersioning(ctx context.Context, store blogstore.StoreInterface, post blogstore.PostInterface) error {
	if store == nil {
		return errors.New("blog store not available")
	}

	if post == nil {
		return errors.New("post is nil")
	}

	if !store.VersioningEnabled() {
		return nil
	}

	lastVersioningList, err := store.VersioningList(ctx, blogstore.NewVersioningQuery().
		SetEntityType(blogstore.VERSIONING_TYPE_POST).
		SetEntityID(post.GetID()).
		SetOrderBy(versionstore.COLUMN_CREATED_AT).
		SetSortOrder(neat.SortDesc).
		SetLimit(1))
	if err != nil {
		return err
	}

	content, err := post.MarshalToVersioning()
	if err != nil {
		return err
	}

	lastVersioning := lo.IfF[blogstore.VersioningInterface](len(lastVersioningList) > 0, func() blogstore.VersioningInterface {
		return lastVersioningList[0]
	}).ElseF(func() blogstore.VersioningInterface {
		return nil
	})
	if lastVersioning != nil {
		if lastVersioning.Content() == content {
			return nil
		}
	}

	return store.VersioningCreate(ctx, blogstore.NewVersioning().
		SetEntityID(post.GetID()).
		SetEntityType(blogstore.VERSIONING_TYPE_POST).
		SetContent(content))
}
