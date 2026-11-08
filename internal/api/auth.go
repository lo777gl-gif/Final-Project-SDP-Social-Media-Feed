package api

import (
	"errors"
	"log"
	"net/http"

	"final/internal/auth"
	"final/internal/structs"
)

type userResponse struct {
	ID       int    `json:"id"`
	Email    string `json:"email"`
	Username string `json:"username"`
	Role     string `json:"role"`
}

func toUserResponse(u *structs.User) userResponse {
	return userResponse{ID: u.ID, Email: u.Email, Username: u.Username, Role: string(u.Role)}
}

func (s *Server) handleRegister(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "bad request body")
		return
	}

	u, err := s.auth.Register(req.Email, req.Username, req.Password)
	if errors.Is(err, auth.ErrBadInput) {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if errors.Is(err, auth.ErrUserExists) {
		writeError(w, http.StatusConflict, err.Error())
		return
	}
	if err != nil {
		log.Println("register:", err)
		writeError(w, http.StatusInternalServerError, "server error")
		return
	}
	writeJSON(w, http.StatusCreated, toUserResponse(u))
}

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Login    string `json:"login"`
		Password string `json:"password"`
	}
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "bad request body")
		return
	}

	u, sess, err := s.auth.Login(req.Login, req.Password)
	if errors.Is(err, auth.ErrWrongLogin) {
		writeError(w, http.StatusUnauthorized, err.Error())
		return
	}
	if err != nil {
		log.Println("login:", err)
		writeError(w, http.StatusInternalServerError, "server error")
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookie,
		Value:    sess.ID,
		Path:     "/",
		Expires:  sess.ExpiresAt,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
	writeJSON(w, http.StatusOK, toUserResponse(u))
}

func (s *Server) handleLogout(w http.ResponseWriter, r *http.Request) {
	c, err := r.Cookie(sessionCookie)
	if err == nil {
		s.auth.Logout(c.Value)
	}
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookie,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
	})
	writeJSON(w, http.StatusOK, map[string]string{"status": "logged out"})
}

func (s *Server) handleMe(w http.ResponseWriter, r *http.Request) {
	u := s.CurrentUser(r)
	if u == nil {
		writeError(w, http.StatusUnauthorized, "please log in")
		return
	}
	writeJSON(w, http.StatusOK, toUserResponse(u))
}
