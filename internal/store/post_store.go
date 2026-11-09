package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"final/internal/structs"
)

var ErrNotFound = errors.New("post not found")

const NoLimit = -1

const selectPosts = `SELECT p.id, p.user_id, u.username, p.kind, p.title, p.content, p.media_url,
	(SELECT COUNT(*) FROM post_reactions pr WHERE pr.post_id = p.id AND pr.value = 1),
	(SELECT COUNT(*) FROM post_reactions pr WHERE pr.post_id = p.id AND pr.value = -1),
	COALESCE((SELECT pr.value FROM post_reactions pr WHERE pr.post_id = p.id AND pr.user_id = ?), 0),
	p.created_at, p.updated_at
	FROM posts p JOIN users u ON u.id = p.user_id`

type PostStore struct {
	db *sql.DB
}

func NewPostStore(db *sql.DB) *PostStore {
	return &PostStore{db: db}
}

func (s *PostStore) Create(ctx context.Context, p *structs.Post) error {
	res, err := s.db.ExecContext(ctx,
		`INSERT INTO posts (user_id, kind, title, content, media_url, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		p.UserID, string(p.Kind), p.Title, p.Content, p.MediaURL, p.CreatedAt, p.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("insert post: %w", err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return fmt.Errorf("get post id: %w", err)
	}
	p.ID = int(id)
	return nil
}

func (s *PostStore) GetByID(ctx context.Context, id, viewerID int) (*structs.Post, error) {
	row := s.db.QueryRowContext(ctx, selectPosts+` WHERE p.id = ?`, viewerID, id)

	p, err := scanPost(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get post: %w", err)
	}
	return p, nil
}

func (s *PostStore) List(ctx context.Context, viewerID, limit, offset int) ([]structs.Post, error) {
	rows, err := s.db.QueryContext(ctx,
		selectPosts+` ORDER BY p.created_at DESC, p.id DESC LIMIT ? OFFSET ?`, viewerID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("list posts: %w", err)
	}
	defer rows.Close()

	posts := make([]structs.Post, 0)
	for rows.Next() {
		p, err := scanPost(rows)
		if err != nil {
			return nil, fmt.Errorf("scan post: %w", err)
		}
		posts = append(posts, *p)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("read posts: %w", err)
	}
	return posts, nil
}

func (s *PostStore) Delete(ctx context.Context, id int) error {
	res, err := s.db.ExecContext(ctx, `DELETE FROM posts WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("delete post: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("delete post: %w", err)
	}
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

type scanner interface {
	Scan(dest ...any) error
}

func scanPost(sc scanner) (*structs.Post, error) {
	p := &structs.Post{}
	err := sc.Scan(&p.ID, &p.UserID, &p.AuthorUsername, &p.Kind, &p.Title, &p.Content, &p.MediaURL,
		&p.Reactions.LikeCount, &p.Reactions.DislikeCount, &p.Reactions.UserReaction,
		&p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return p, nil
}
