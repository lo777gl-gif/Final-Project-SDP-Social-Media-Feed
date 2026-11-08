package observer

import (
	"database/sql"
	"final/internal/structs"
)

type Repo struct {
	db *sql.DB
}

func NewRepo(db *sql.DB) *Repo {
	return &Repo{
		db: db,
	}
}

func (r *Repo) SavePostReactions(userID int, postID int, reaction structs.Reaction) error {
	if reaction == structs.ReactionNone {
		_, err := r.db.Exec(
			`DELETE FROM post_reactions WHERE user_id = ? AND post_id = ?`,
			userID, postID)
		return err
	}
	_, err := r.db.Exec(`INSERT INTO post_reactions (user_id, post_id, value)
	VALUES (?, ?, ?) ON CONFLICT (user_id, post_id) DO UPDATE SET value = excluded.value, updated_at = CURRENT_TIMESTAMP`,
		userID, postID, reaction)
	return err
}

func (r *Repo) SaveCommentReactions(userID int, commentID int, reaction structs.Reaction) error {
	if reaction == structs.ReactionNone {
		_, err := r.db.Exec(
			`DELETE FROM comment_reactions WHERE user_id = ? AND comment_id = ?`,
			userID, commentID)
		return err
	}
	_, err := r.db.Exec(`INSERT INTO comment_reactions (user_id, comment_id, value)
	VALUES (?, ?, ?) ON CONFLICT (user_id, comment_id) DO UPDATE SET value = excluded.value, updated_at = CURRENT_TIMESTAMP`,
		userID, commentID, reaction)
	return err
}

func (r *Repo) GetPostReactionSummary(postID int, userID int) (structs.ReactionSummary, error) {
	summary := &structs.ReactionSummary{}
	err := r.db.QueryRow(`SELECT COUNT(*) FILTER (WHERE value = 1),
	COUNT(*) FILTER (WHERE value = -1), COALESCE(SUM(value) FILTER (WHERE user_id = ?), 0)
	FROM post_reactions WHERE post_id = ?`, userID, postID).Scan(&summary.LikeCount, &summary.DislikeCount,
		&summary.UserReaction)
	return *summary, err
}

func (r *Repo) GetCommentReactionSummary(commentID int, userID int) (structs.ReactionSummary, error) {
	summary := &structs.ReactionSummary{}
	err := r.db.QueryRow(`SELECT COUNT(*) FILTER (WHERE value = 1),
	COUNT(*) FILTER (WHERE value = -1), COALESCE(SUM(value) FILTER (WHERE user_id = ?), 0)
	FROM comment_reactions WHERE comment_id = ?`, userID, commentID).Scan(&summary.LikeCount, &summary.DislikeCount,
		&summary.UserReaction)
	return *summary, err
}

func (r *Repo) InsertNotification(notif structs.Notification) error {
	_, err := r.db.Exec(`INSERT INTO notifications (user_id, post_id, actor_id, comment_id, type) VALUES (?, ?, ?, ?, ?)`,
		notif.UserID, notif.PostID, notif.ActorID, notif.CommentID, notif.Type)
	return err
}

func (r *Repo) ListNotificationForUser(userID int) ([]structs.Notification, error) {
	rows, err := r.db.Query(`SELECT id, user_id, actor_id, post_id, comment_id, type, read_at, created_at FROM notifications WHERE user_id = ? ORDER BY created_at DESC, id DESC`,
		userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := []structs.Notification{}
	for rows.Next() {
		notif := structs.Notification{}
		if err := rows.Scan(&notif.ID, &notif.UserID, &notif.ActorID, &notif.PostID, &notif.CommentID,
			&notif.Type, &notif.ReadAt, &notif.CreatedAt); err != nil {
			return nil, err
		}
		result = append(result, notif)
	}
	return result, rows.Err()
}

func (r *Repo) MarkNotificationRead(id int) error {
	_, err := r.db.Exec(`UPDATE notifications SET read_at = CURRENT_TIMESTAMP WHERE id = ?`, id)
	return err
}
