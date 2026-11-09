package api

import (
	"errors"
	"log"
	"net/http"
	"strconv"

	"final/internal/facade"
	"final/internal/structs"
)

func (s *Server) handleCreatePost(w http.ResponseWriter, r *http.Request) {
	user := s.CurrentUser(r)

	var req struct {
		Kind     string `json:"kind"`
		Title    string `json:"title"`
		Content  string `json:"content"`
		MediaURL string `json:"media_url"`
	}
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "bad request body")
		return
	}

	kind := structs.PostKind(req.Kind)
	if req.Kind == "" {
		kind = structs.PostKindText
	}

	post, err := s.facade.CreatePost(r.Context(), kind, req.Title, user.ID, req.Content, req.MediaURL)
	if errors.Is(err, facade.ErrInvalidPost) {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err != nil {
		log.Println("create post:", err)
		writeError(w, http.StatusInternalServerError, "server error")
		return
	}
	writeJSON(w, http.StatusCreated, post)
}

func (s *Server) handleFeed(w http.ResponseWriter, r *http.Request) {
	user := s.CurrentUser(r)
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	posts, err := s.facade.Feed(r.Context(), user.ID, limit, offset)
	if err != nil {
		log.Println("feed:", err)
		writeError(w, http.StatusInternalServerError, "server error")
		return
	}
	writeJSON(w, http.StatusOK, posts)
}
