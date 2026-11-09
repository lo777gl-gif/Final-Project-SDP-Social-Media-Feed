package api

import (
	"errors"
	"log"
	"net/http"
	"strconv"

	"final/internal/facade"
	"final/internal/strategy"
	"final/internal/structs"
)

type moderatedPost struct {
	Post   structs.Post    `json:"post"`
	Result strategy.Result `json:"result"`
}

func (s *Server) handleStrategies(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, strategy.Names())
}

func (s *Server) handleModerationPosts(w http.ResponseWriter, r *http.Request) {
	st, ok := strategy.ByName(r.URL.Query().Get("strategy"))
	if !ok {
		writeError(w, http.StatusBadRequest, "unknown strategy")
		return
	}
	mod := strategy.NewModerator(strategy.AllowAll{})
	mod.SetStrategy(st)

	posts, err := s.facade.AllPosts(r.Context())
	if err != nil {
		log.Println("moderation list:", err)
		writeError(w, http.StatusInternalServerError, "server error")
		return
	}

	result := []moderatedPost{}
	for _, p := range posts {
		result = append(result, moderatedPost{Post: p, Result: mod.Check(p)})
	}
	writeJSON(w, http.StatusOK, result)
}

func (s *Server) handleModerationCheck(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Strategy string `json:"strategy"`
		Title    string `json:"title"`
		Content  string `json:"content"`
		MediaURL string `json:"media_url"`
	}
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "bad request body")
		return
	}
	st, ok := strategy.ByName(req.Strategy)
	if !ok {
		writeError(w, http.StatusBadRequest, "unknown strategy")
		return
	}
	p := structs.Post{Title: req.Title, Content: req.Content, MediaURL: req.MediaURL}
	writeJSON(w, http.StatusOK, strategy.NewModerator(st).Check(p))
}

func (s *Server) handleDeletePost(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "bad post id")
		return
	}
	err = s.facade.DeletePost(r.Context(), id)
	if errors.Is(err, facade.ErrPostNotFound) {
		writeError(w, http.StatusNotFound, "post not found")
		return
	}
	if err != nil {
		log.Println("delete post:", err)
		writeError(w, http.StatusInternalServerError, "server error")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}
