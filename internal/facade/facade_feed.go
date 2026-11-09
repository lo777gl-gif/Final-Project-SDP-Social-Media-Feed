package facade

import (
	"context"

	"final/internal/decorator"
	"final/internal/factory"
	"final/internal/store"
	"final/internal/structs"
)

const (
	DefaultFeedLimit = 20
	MaxFeedLimit     = 50
)

var (
	ErrInvalidPost  = factory.ErrInvalidPost
	ErrPostNotFound = store.ErrNotFound
)

func (f *Facade) CreatePost(ctx context.Context, kind structs.PostKind, title string, userID int, content, mediaURL string) (decorator.PostView, error) {
	post, err := f.factory.Create(kind, title, userID, content, mediaURL)
	if err != nil {
		return decorator.PostView{}, err
	}
	if err := f.posts.Create(ctx, post); err != nil {
		return decorator.PostView{}, err
	}
	saved, err := f.posts.GetByID(ctx, post.ID, userID)
	if err != nil {
		return decorator.PostView{}, err
	}
	return f.view(*saved), nil
}

func (f *Facade) Feed(ctx context.Context, viewerID, limit, offset int) ([]decorator.PostView, error) {
	if limit <= 0 || limit > MaxFeedLimit {
		limit = DefaultFeedLimit
	}
	if offset < 0 {
		offset = 0
	}

	posts, err := f.posts.List(ctx, viewerID, limit, offset)
	if err != nil {
		return nil, err
	}

	views := make([]decorator.PostView, 0, len(posts))
	for _, p := range posts {
		views = append(views, f.view(p))
	}
	return views, nil
}

func (f *Facade) AllPosts(ctx context.Context) ([]structs.Post, error) {
	return f.posts.List(ctx, 0, store.NoLimit, 0)
}

func (f *Facade) DeletePost(ctx context.Context, id int) error {
	return f.posts.Delete(ctx, id)
}

func (f *Facade) view(p structs.Post) decorator.PostView {
	component := decorator.Flagged(decorator.Edited(decorator.NewPost(p)), func(p structs.Post) bool {
		return !f.moderator.Check(p).Allowed
	})
	return decorator.Build(component)
}
