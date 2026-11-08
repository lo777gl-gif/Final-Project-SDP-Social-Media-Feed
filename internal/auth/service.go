package auth

import (
	"crypto/pbkdf2"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"net/mail"
	"strings"
	"time"
	"unicode/utf8"

	"final/internal/structs"
)

var (
	ErrBadInput   = errors.New("invalid input")
	ErrUserExists = errors.New("email or username already taken")
	ErrWrongLogin = errors.New("wrong login or password")
	ErrNoSession  = errors.New("not logged in")
)

const SessionLifetime = 24 * time.Hour

type Service struct {
	repo *Repo
}

func NewService(repo *Repo) *Service {
	return &Service{repo: repo}
}

func (s *Service) Register(email, username, password string) (*structs.User, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	username = strings.TrimSpace(username)

	if _, err := mail.ParseAddress(email); err != nil {
		return nil, fmt.Errorf("%w: bad email", ErrBadInput)
	}
	if utf8.RuneCountInString(username) < 3 || utf8.RuneCountInString(username) > 30 {
		return nil, fmt.Errorf("%w: username must be 3-30 characters", ErrBadInput)
	}
	if len(password) < 6 {
		return nil, fmt.Errorf("%w: password must be at least 6 characters", ErrBadInput)
	}

	taken, err := s.repo.EmailOrUsernameTaken(email, username)
	if err != nil {
		return nil, err
	}
	if taken {
		return nil, ErrUserExists
	}

	hash, err := HashPassword(password)
	if err != nil {
		return nil, err
	}

	// First user becomes moderator.
	role := structs.RoleUser
	count, err := s.repo.CountUsers()
	if err != nil {
		return nil, err
	}
	if count == 0 {
		role = structs.RoleModerator
	}

	u := &structs.User{Email: email, Username: username, PasswordHash: hash, Role: role}
	if err := s.repo.CreateUser(u); err != nil {
		return nil, err
	}
	return u, nil
}

func (s *Service) Login(login, password string) (*structs.User, *structs.Session, error) {
	u, err := s.repo.GetUserByLogin(strings.TrimSpace(login))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil, ErrWrongLogin
	}
	if err != nil {
		return nil, nil, err
	}
	if !CheckPassword(password, u.PasswordHash) {
		return nil, nil, ErrWrongLogin
	}

	sess := &structs.Session{
		ID:        rand.Text(),
		UserID:    u.ID,
		ExpiresAt: time.Now().UTC().Add(SessionLifetime),
		CreatedAt: time.Now().UTC(),
	}
	if err := s.repo.CreateSession(sess); err != nil {
		return nil, nil, err
	}
	return u, sess, nil
}

func (s *Service) Logout(sessionID string) error {
	return s.repo.DeleteSession(sessionID)
}

func (s *Service) UserBySession(sessionID string) (*structs.User, error) {
	sess, err := s.repo.GetSession(sessionID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNoSession
	}
	if err != nil {
		return nil, err
	}
	if time.Now().UTC().After(sess.ExpiresAt) {
		s.repo.DeleteSession(sessionID)
		return nil, ErrNoSession
	}
	u, err := s.repo.GetUserByID(sess.UserID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNoSession
	}
	return u, err
}

func HashPassword(password string) (string, error) {
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}
	key, err := pbkdf2.Key(sha256.New, password, salt, 100000, 32)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(salt) + ":" + hex.EncodeToString(key), nil
}

func CheckPassword(password, stored string) bool {
	parts := strings.Split(stored, ":")
	if len(parts) != 2 {
		return false
	}
	salt, err := hex.DecodeString(parts[0])
	if err != nil {
		return false
	}
	want, err := hex.DecodeString(parts[1])
	if err != nil {
		return false
	}
	got, err := pbkdf2.Key(sha256.New, password, salt, 100000, 32)
	if err != nil {
		return false
	}
	return subtle.ConstantTimeCompare(got, want) == 1
}
