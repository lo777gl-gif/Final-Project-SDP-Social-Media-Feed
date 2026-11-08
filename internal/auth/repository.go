package auth

import (
	"database/sql"
	"time"

	"final/internal/structs"
)

type Repo struct {
	db *sql.DB
}

func NewRepo(db *sql.DB) *Repo {
	return &Repo{db: db}
}

func (r *Repo) CountUsers() (int, error) {
	var n int
	err := r.db.QueryRow(`SELECT COUNT(*) FROM users`).Scan(&n)
	return n, err
}

func (r *Repo) CreateUser(u *structs.User) error {
	res, err := r.db.Exec(`INSERT INTO users (email, username, password_hash, role) VALUES (?, ?, ?, ?)`,
		u.Email, u.Username, u.PasswordHash, u.Role)
	if err != nil {
		return err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return err
	}
	u.ID = int(id)
	return nil
}

func (r *Repo) scanUser(row *sql.Row) (*structs.User, error) {
	u := &structs.User{}
	err := row.Scan(&u.ID, &u.Email, &u.Username, &u.PasswordHash, &u.Role, &u.CreatedAt)
	if err != nil {
		return nil, err
	}
	return u, nil
}

func (r *Repo) GetUserByID(id int) (*structs.User, error) {
	return r.scanUser(r.db.QueryRow(
		`SELECT id, email, username, password_hash, role, created_at FROM users WHERE id = ?`, id))
}

func (r *Repo) GetUserByLogin(login string) (*structs.User, error) {
	return r.scanUser(r.db.QueryRow(
		`SELECT id, email, username, password_hash, role, created_at FROM users WHERE email = ? OR username = ?`,
		login, login))
}

func (r *Repo) EmailOrUsernameTaken(email, username string) (bool, error) {
	var n int
	err := r.db.QueryRow(`SELECT COUNT(*) FROM users WHERE email = ? OR username = ?`, email, username).Scan(&n)
	return n > 0, err
}

func (r *Repo) CreateSession(s *structs.Session) error {
	_, err := r.db.Exec(`INSERT INTO sessions (id, user_id, expires_at) VALUES (?, ?, ?)`,
		s.ID, s.UserID, s.ExpiresAt)
	return err
}

func (r *Repo) GetSession(id string) (*structs.Session, error) {
	s := &structs.Session{}
	err := r.db.QueryRow(`SELECT id, user_id, expires_at, created_at FROM sessions WHERE id = ?`, id).
		Scan(&s.ID, &s.UserID, &s.ExpiresAt, &s.CreatedAt)
	if err != nil {
		return nil, err
	}
	return s, nil
}

func (r *Repo) DeleteSession(id string) error {
	_, err := r.db.Exec(`DELETE FROM sessions WHERE id = ?`, id)
	return err
}

func (r *Repo) DeleteExpiredSessions() error {
	_, err := r.db.Exec(`DELETE FROM sessions WHERE expires_at < ?`, time.Now().UTC())
	return err
}
