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

func (r *Repo) PostAuthorID(postID int) (int, error) {
	var id int
	err := r.db.QueryRow(`SELECT user_id FROM posts WHERE id = ?`, postID).Scan(&id)
	return id, err
}

func (r *Repo) CommentAuthorID(commentID int) (int, error) {
	var id int
	err := r.db.QueryRow(`SELECT user_id FROM comments WHERE id = ?`, commentID).Scan(&id)
	return id, err
}

func (r *Repo) CommentPostID(commentID int) (int, error) {
	var id int
	err := r.db.QueryRow(`SELECT post_id FROM comments WHERE id = ?`, commentID).Scan(&id)
	return id, err
}

func (r *Repo) MarkNotificationRead(id int, userID int) (bool, error) {
	res, err := r.db.Exec(`UPDATE notifications SET read_at = CURRENT_TIMESTAMP WHERE id = ? AND user_id = ?`, id, userID)
	if err != nil {
		return false, err
	}
	n, err := res.RowsAffected()
	return n > 0, err
}

func (r *Repo) CountUnreadNotifications(userID int) (int, error) {
	var n int
	err := r.db.QueryRow(`SELECT COUNT(*) FROM notifications WHERE user_id = ? AND read_at IS NULL`, userID).Scan(&n)
	return n, err
}

func (r *Repo) CreateComment(postID int, userID int, parentCommentID *int, content string) (int, error) {
	res, err := r.db.Exec(`INSERT INTO comments (post_id, user_id, parent_comment_id, content) VALUES (?, ?, ?, ?)`,
		postID, userID, parentCommentID, content)
	if err != nil {
		return 0, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	return int(id), nil
}

func (r *Repo) ListComments(postID int, viewerID int) ([]structs.Comment, error) {
	rows, err := r.db.Query(`SELECT c.id, c.post_id, c.user_id, c.parent_comment_id, u.username, c.content,
		(SELECT COUNT(*) FROM comment_reactions cr WHERE cr.comment_id = c.id AND cr.value = 1),
		(SELECT COUNT(*) FROM comment_reactions cr WHERE cr.comment_id = c.id AND cr.value = -1),
		COALESCE((SELECT cr.value FROM comment_reactions cr WHERE cr.comment_id = c.id AND cr.user_id = ?), 0),
		c.created_at, c.updated_at
		FROM comments c JOIN users u ON u.id = c.user_id
		WHERE c.post_id = ? ORDER BY c.created_at, c.id`, viewerID, postID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := []structs.Comment{}
	for rows.Next() {
		c := structs.Comment{}
		if err := rows.Scan(&c.ID, &c.PostID, &c.UserID, &c.ParentCommentID, &c.AuthorUsername, &c.Content,
			&c.Reactions.LikeCount, &c.Reactions.DislikeCount, &c.Reactions.UserReaction,
			&c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, err
		}
		result = append(result, c)
	}
	return result, rows.Err()
}
