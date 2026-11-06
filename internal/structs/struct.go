package structs

import "time"

type Reaction int

const (
	ReactionLike    Reaction = 1
	ReactionNone    Reaction = 0
	ReactionDislike Reaction = -1
)

type Role string

const (
	RoleUser      Role = "user"
	RoleModerator Role = "moderator"
)

type NotificationType string

const (
	NotifComment  NotificationType = "comment"
	NotifReaction NotificationType = "reaction"
)

type User struct {
	ID           int
	Email        string
	Username     string
	Role         Role
	PasswordHash string
	CreatedAt    time.Time
}

type Session struct {
	ID        string
	UserID    int
	ExpiresAt time.Time
	CreatedAt time.Time
}

type Post struct {
	ID             int
	UserID         int
	AuthorUsername string
	Title          string
	Content        string
	Reactions      ReactionSummary
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type ReactionSummary struct {
	LikeCount    int
	DislikeCount int
	UserReaction Reaction
}

type PostReaction struct {
	UserID    int
	PostID    int
	Value     Reaction
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Comment struct {
	ID              int
	PostID          int
	UserID          int
	ParentCommentID *int
	AuthorUsername  string
	Content         string
	Reactions       ReactionSummary
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type CommentReaction struct {
	UserID    int
	CommentID int
	Value     Reaction
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Notification struct {
	ID        int
	UserID    int
	ActorID   int
	PostID    *int
	CommentID *int
	Type      NotificationType
	ReadAt    *time.Time
	CreatedAt time.Time
}
