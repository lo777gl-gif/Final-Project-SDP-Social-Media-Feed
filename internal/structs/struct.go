package internal

import "time"

const (
	ReactionLike    = 1
	ReactionDislike = -1
)

type User struct {
	ID        int
	Email     string
	Username  string
	Password  string
	CreatedAt time.Time
}

type Session struct {
	ID        int
	UserID    int
	ExpiresAt time.Time
	CreatedAt time.Time
}

type Post struct {
	ID             int
	UserID         int
	CommentID      int
	AuthorUsername string
	Title          string
	Content        string
	LikeCount      int
	DislikeCount   int
	UserReaction   int
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type PostReaction struct {
	UserID    int
	PostID    int
	Value     int
	CreatedAt time.Time
}

type Comment struct {
	ID             int
	PostID         int
	UserID         int
	AuthorUsername string
	Content        string
	LikeCount      int
	DislikeCount   int
	UserReaction   int
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type CommentReaction struct {
	UserID    int
	CommentID int
	Value     int
	CreatedAt time.Time
}
