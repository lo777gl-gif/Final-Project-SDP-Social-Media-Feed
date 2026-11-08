package api

import (
	"log"
	"net/http"
	"strconv"

	"final/internal/strategy"
)

func (s *Server) handleExternalSources(w http.ResponseWriter, r *http.Request) {
	names := []string{}
	for name := range s.sources {
		names = append(names, name)
	}
	writeJSON(w, http.StatusOK, names)
}

func (s *Server) handleExternal(w http.ResponseWriter, r *http.Request) {
	src, ok := s.sources[r.URL.Query().Get("source")]
	if !ok {
		writeError(w, http.StatusBadRequest, "unknown source")
		return
	}

	limit, err := strconv.Atoi(r.URL.Query().Get("limit"))
	if err != nil || limit <= 0 || limit > 50 {
		limit = 10
	}

	posts, err := src.FetchPosts(limit)
	if err != nil {
		log.Println("external:", err)
		writeError(w, http.StatusBadGateway, "could not load external content")
		return
	}

	// External posts also go through moderation.
	st, ok := strategy.ByName(r.URL.Query().Get("strategy"))
	if !ok {
		st, _ = strategy.ByName("strict")
	}
	mod := strategy.NewModerator(st)
	writeJSON(w, http.StatusOK, mod.Filter(posts))
}
