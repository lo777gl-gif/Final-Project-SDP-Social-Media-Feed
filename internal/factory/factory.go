package factory

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"
	"unicode/utf8"

	"final/internal/structs"
)

var ErrInvalidPost = errors.New("invalid post")

const MaxContentLength = 1000

type Factory struct{}

func NewFactory() *Factory {
	return &Factory{}
}

func (f *Factory) Create(kind structs.PostKind, title string, userID int, content, mediaURL string) (*structs.Post, error) {
	now := time.Now().UTC()
	p := &structs.Post{
		UserID:    userID,
		Kind:      kind,
		Title:     strings.TrimSpace(title),
		Content:   strings.TrimSpace(content),
		MediaURL:  strings.TrimSpace(mediaURL),
		CreatedAt: now,
		UpdatedAt: now,
	}

	if kind == structs.PostKindText {
		p.MediaURL = ""
	}

	if err := f.Validate(p); err != nil {
		return nil, err
	}
	return p, nil
}

func (f *Factory) Validate(p *structs.Post) error {
	if p == nil {
		return invalid("post is nil")
	}
	if p.UserID <= 0 {
		return invalid("user_id must be positive")
	}
	if utf8.RuneCountInString(p.Content) > MaxContentLength {
		return invalid("content is longer than %d characters", MaxContentLength)
	}

	switch p.Kind {
	case structs.PostKindText:
		return validateText(p)
	case structs.PostKindImage:
		return validateImage(p)
	case structs.PostKindLink:
		return validateLink(p)
	default:
		return invalid("unknown kind %q (allowed: text, image, link)", p.Kind)
	}
}

func validateText(p *structs.Post) error {
	if p.Content == "" {
		return invalid("text post must have content")
	}
	return nil
}

func validateImage(p *structs.Post) error {
	return validateURL(p.MediaURL, "image")
}

func validateLink(p *structs.Post) error {
	return validateURL(p.MediaURL, "link")
}

func validateURL(raw, what string) error {
	if raw == "" {
		return invalid("%s post must have media_url", what)
	}
	u, err := url.ParseRequestURI(raw)
	if err != nil {
		return invalid("media_url is not a valid URL")
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return invalid("media_url must start with http:// or https://")
	}
	if u.Host == "" {
		return invalid("media_url has no host")
	}
	return nil
}

func invalid(format string, args ...any) error {
	return fmt.Errorf("%w: %s", ErrInvalidPost, fmt.Sprintf(format, args...))
}
