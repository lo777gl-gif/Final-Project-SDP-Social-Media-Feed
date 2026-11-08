package api

import (
	"log"
	"net/http"
	"strconv"

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

	posts, err := s.listPosts()
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
	res, err := s.db.Exec(`DELETE FROM posts WHERE id = ?`, id)
	if err != nil {
		log.Println("delete post:", err)
		writeError(w, http.StatusInternalServerError, "server error")
		return
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		writeError(w, http.StatusNotFound, "post not found")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

func (s *Server) listPosts() ([]structs.Post, error) {
	rows, err := s.db.Query(`SELECT p.id, p.user_id, u.username, p.kind, p.title, p.content, p.media_url, p.created_at, p.updated_at
	FROM posts p JOIN users u ON u.id = p.user_id ORDER BY p.created_at DESC, p.id DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	posts := []structs.Post{}
	for rows.Next() {
		p := structs.Post{}
		err := rows.Scan(&p.ID, &p.UserID, &p.AuthorUsername, &p.Kind, &p.Title, &p.Content, &p.MediaURL,
			&p.CreatedAt, &p.UpdatedAt)
		if err != nil {
			return nil, err
		}
		posts = append(posts, p)
	}
	return posts, rows.Err()
}
