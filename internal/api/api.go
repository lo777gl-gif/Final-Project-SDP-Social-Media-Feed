package api

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"final/internal/adapter"
	"final/internal/auth"
	"final/internal/facade"
	"final/internal/observer"
	"final/internal/structs"
)

const sessionCookie = "session_id"

type Server struct {
	db      *sql.DB
	auth    *auth.Service
	facade  *facade.Facade
	sources map[string]adapter.ContentSource
	social  *observer.Repo
	events  *observer.Subject
}

func NewServer(db *sql.DB) *Server {
	social := observer.NewRepo(db)
	events := observer.NewSubject()
	events.Subscribe(observer.NewNotifier(social))

	s := &Server{
		db:      db,
		auth:    auth.NewService(auth.NewRepo(db)),
		facade:  facade.New(db),
		sources: map[string]adapter.ContentSource{},
		social:  social,
		events:  events,
	}
	s.AddSource(adapter.NewJSONPlaceholderAdapter("https://jsonplaceholder.typicode.com/posts"))
	s.AddSource(adapter.NewRSSAdapter("bbc_news", "https://feeds.bbci.co.uk/news/rss.xml"))
	return s
}

func (s *Server) AddSource(src adapter.ContentSource) {
	s.sources[src.Name()] = src
}

func (s *Server) Routes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/register", s.handleRegister)
	mux.HandleFunc("POST /api/login", s.handleLogin)
	mux.HandleFunc("POST /api/logout", s.handleLogout)
	mux.HandleFunc("GET /api/me", s.handleMe)

	mux.HandleFunc("GET /api/posts", s.loginOnly(s.handleFeed))
	mux.HandleFunc("POST /api/posts", s.loginOnly(s.handleCreatePost))

	mux.HandleFunc("GET /api/external/sources", s.handleExternalSources)
	mux.HandleFunc("GET /api/external", s.handleExternal)

	mux.HandleFunc("GET /api/moderation/strategies", s.moderatorOnly(s.handleStrategies))
	mux.HandleFunc("GET /api/moderation/posts", s.moderatorOnly(s.handleModerationPosts))
	mux.HandleFunc("POST /api/moderation/check", s.moderatorOnly(s.handleModerationCheck))
	mux.HandleFunc("DELETE /api/moderation/posts/{id}", s.moderatorOnly(s.handleDeletePost))

	mux.HandleFunc("POST /api/posts/{id}/reactions", s.loginOnly(s.handleReactPost))
	mux.HandleFunc("POST /api/comments/{id}/reactions", s.loginOnly(s.handleReactComment))
	mux.HandleFunc("GET /api/posts/{id}/comments", s.loginOnly(s.handleListComments))
	mux.HandleFunc("POST /api/posts/{id}/comments", s.loginOnly(s.handleCreateComment))
	mux.HandleFunc("GET /api/notifications", s.loginOnly(s.handleListNotifications))
	mux.HandleFunc("GET /api/notifications/unread-count", s.loginOnly(s.handleUnreadCount))
	mux.HandleFunc("POST /api/notifications/{id}/read", s.loginOnly(s.handleMarkRead))
}

func (s *Server) CurrentUser(r *http.Request) *structs.User {
	c, err := r.Cookie(sessionCookie)
	if err != nil {
		return nil
	}
	u, err := s.auth.UserBySession(c.Value)
	if err != nil {
		return nil
	}
	return u
}

func (s *Server) loginOnly(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if s.CurrentUser(r) == nil {
			writeError(w, http.StatusUnauthorized, "please log in")
			return
		}
		next(w, r)
	}
}

func (s *Server) moderatorOnly(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		u := s.CurrentUser(r)
		if u == nil {
			writeError(w, http.StatusUnauthorized, "please log in")
			return
		}
		if u.Role != structs.RoleModerator {
			writeError(w, http.StatusForbidden, "only moderators can do this")
			return
		}
		next(w, r)
	}
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

func readJSON(r *http.Request, v any) error {
	return json.NewDecoder(r.Body).Decode(v)
}
