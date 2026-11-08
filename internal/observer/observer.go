package observer

type EventType string

const (
	PostReactionEvent EventType = "post_reacted"
	CommentReactionEvent EventType = "comment_reacted"
	CommentCreatedEvent EventType = "comment_created"
)

type Event struct {
	Type EventType
	ActorID int
	PostID int
	CommentID int
}

type Observer interface {
	Notify(e Event) error
}

