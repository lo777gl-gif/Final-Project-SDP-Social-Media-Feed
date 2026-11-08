package observer

import "final/internal/structs"

type Notifier struct {
	store notificationStore
}

func NewNotifier(store notificationStore) *Notifier {
	return &Notifier{store: store}
}

func (n *Notifier) Notify(e Event) error {
	notif := structs.Notification{ActorID: e.ActorID, PostID: &e.PostID}

	var recipientID int
	var err error

	switch e.Type {
	case PostReactionEvent:
		notif.Type = structs.NotifReaction
		recipientID, err = n.store.PostAuthorID(e.PostID)
	case CommentReactionEvent:
		notif.Type = structs.NotifReaction
		notif.CommentID = &e.CommentID
		recipientID, err = n.store.CommentAuthorID(e.CommentID)
	case CommentCreatedEvent:
		notif.Type = structs.NotifComment
		notif.CommentID = &e.CommentID
		recipientID, err = n.store.PostAuthorID(e.PostID)
	default:
		return nil
	}
	if err != nil {
		return err
	}

	if recipientID == e.ActorID {
		return nil
	}

	notif.UserID = recipientID
	return n.store.InsertNotification(notif)
}
