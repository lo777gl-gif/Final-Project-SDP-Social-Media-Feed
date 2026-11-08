package api

import (
	"database/sql"
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"
	"unicode/utf8"

	"final/internal/observer"
	"final/internal/structs"
)

const maxCommentLength = 1000

func pathID(w http.ResponseWriter, r *http.Request, what string) (int, bool) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id <= 0 {
		writeError(w, http.StatusBadRequest, "bad "+what+" id")
		return 0, false
	}
	return id, true
}

func readReactionValue(w http.ResponseWriter, r *http.Request) (structs.Reaction, bool) {
	var req struct {
		Value int `json:"value"`
	}
	if err := readJSON(r, &req); err != nil || req.Value < -1 || req.Value > 1 {
		writeError(w, http.StatusBadRequest, "value must be -1, 0 or 1")
		return 0, false
	}
	return structs.Reaction(req.Value), true
}

func serverError(w http.ResponseWriter, what string, err error) {
	log.Println(what+":", err)
	writeError(w, http.StatusInternalServerError, "server error")
}

func (s *Server) handleReactPost(w http.ResponseWriter, r *http.Request) {
	user := s.CurrentUser(r)
	postID, ok := pathID(w, r, "post")
	if !ok {
		return
	}
	value, ok := readReactionValue(w, r)
	if !ok {
		return
	}

	if _, err := s.social.PostAuthorID(postID); errors.Is(err, sql.ErrNoRows) {
		writeError(w, http.StatusNotFound, "post not found")
		return
	} else if err != nil {
		serverError(w, "react post", err)
		return
	}

	if err := s.social.SavePostReactions(user.ID, postID, value); err != nil {
		serverError(w, "react post", err)
		return
	}
	if value != structs.ReactionNone {
		s.events.Publish(observer.Event{Type: observer.PostReactionEvent, ActorID: user.ID, PostID: postID})
	}

	summary, err := s.social.GetPostReactionSummary(postID, user.ID)
	if err != nil {
		serverError(w, "react post", err)
		return
	}
	writeJSON(w, http.StatusOK, summary)
}

func (s *Server) handleReactComment(w http.ResponseWriter, r *http.Request) {
	user := s.CurrentUser(r)
	commentID, ok := pathID(w, r, "comment")
	if !ok {
		return
	}
	value, ok := readReactionValue(w, r)
	if !ok {
		return
	}

	postID, err := s.social.CommentPostID(commentID)
	if errors.Is(err, sql.ErrNoRows) {
		writeError(w, http.StatusNotFound, "comment not found")
		return
	}
	if err != nil {
		serverError(w, "react comment", err)
		return
	}

	if err := s.social.SaveCommentReactions(user.ID, commentID, value); err != nil {
		serverError(w, "react comment", err)
		return
	}
	if value != structs.ReactionNone {
		s.events.Publish(observer.Event{Type: observer.CommentReactionEvent, ActorID: user.ID, PostID: postID, CommentID: commentID})
	}

	summary, err := s.social.GetCommentReactionSummary(commentID, user.ID)
	if err != nil {
		serverError(w, "react comment", err)
		return
	}
	writeJSON(w, http.StatusOK, summary)
}

func (s *Server) handleListComments(w http.ResponseWriter, r *http.Request) {
	user := s.CurrentUser(r)
	postID, ok := pathID(w, r, "post")
	if !ok {
		return
	}

	comments, err := s.social.ListComments(postID, user.ID)
	if err != nil {
		serverError(w, "list comments", err)
		return
	}
	writeJSON(w, http.StatusOK, comments)
}

func (s *Server) handleCreateComment(w http.ResponseWriter, r *http.Request) {
	user := s.CurrentUser(r)
	postID, ok := pathID(w, r, "post")
	if !ok {
		return
	}

	var req struct {
		Content         string `json:"content"`
		ParentCommentID *int   `json:"parent_comment_id"`
	}
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "bad request body")
		return
	}
	req.Content = strings.TrimSpace(req.Content)
	if req.Content == "" || utf8.RuneCountInString(req.Content) > maxCommentLength {
		writeError(w, http.StatusBadRequest, "comment must be 1-1000 characters")
		return
	}

	if _, err := s.social.PostAuthorID(postID); errors.Is(err, sql.ErrNoRows) {
		writeError(w, http.StatusNotFound, "post not found")
		return
	} else if err != nil {
		serverError(w, "create comment", err)
		return
	}

	if req.ParentCommentID != nil {
		parentPost, err := s.social.CommentPostID(*req.ParentCommentID)
		if errors.Is(err, sql.ErrNoRows) || (err == nil && parentPost != postID) {
			writeError(w, http.StatusBadRequest, "parent comment is not on this post")
			return
		}
		if err != nil {
			serverError(w, "create comment", err)
			return
		}
	}

	id, err := s.social.CreateComment(postID, user.ID, req.ParentCommentID, req.Content)
	if err != nil {
		serverError(w, "create comment", err)
		return
	}
	s.events.Publish(observer.Event{Type: observer.CommentCreatedEvent, ActorID: user.ID, PostID: postID, CommentID: id})

	writeJSON(w, http.StatusCreated, map[string]int{"id": id})
}

func (s *Server) handleListNotifications(w http.ResponseWriter, r *http.Request) {
	user := s.CurrentUser(r)
	list, err := s.social.ListNotificationForUser(user.ID)
	if err != nil {
		serverError(w, "list notifications", err)
		return
	}
	writeJSON(w, http.StatusOK, list)
}

func (s *Server) handleUnreadCount(w http.ResponseWriter, r *http.Request) {
	user := s.CurrentUser(r)
	n, err := s.social.CountUnreadNotifications(user.ID)
	if err != nil {
		serverError(w, "unread count", err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]int{"count": n})
}

func (s *Server) handleMarkRead(w http.ResponseWriter, r *http.Request) {
	user := s.CurrentUser(r)
	id, ok := pathID(w, r, "notification")
	if !ok {
		return
	}
	found, err := s.social.MarkNotificationRead(id, user.ID)
	if err != nil {
		serverError(w, "mark read", err)
		return
	}
	if !found {
		writeError(w, http.StatusNotFound, "notification not found")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "read"})
}
