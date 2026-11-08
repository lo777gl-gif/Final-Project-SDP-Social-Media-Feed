package strategy

import (
	"strings"
	"unicode"

	"final/internal/structs"
)

type Result struct {
	Allowed bool   `json:"allowed"`
	Reason  string `json:"reason"`
}

type ModerationStrategy interface {
	Name() string
	Check(p structs.Post) Result
}

type AllowAll struct{}

func (s AllowAll) Name() string {
	return "none"
}

func (s AllowAll) Check(p structs.Post) Result {
	return Result{Allowed: true}
}

type BannedWords struct {
	Words []string
}

func (s BannedWords) Name() string {
	return "banned_words"
}

func (s BannedWords) Check(p structs.Post) Result {
	text := strings.ToLower(p.Title + " " + p.Content)
	for _, w := range s.Words {
		if strings.Contains(text, w) {
			return Result{Allowed: false, Reason: "contains banned word: " + w}
		}
	}
	return Result{Allowed: true}
}

type CapsSpam struct {
	MaxPercent int
}

func (s CapsSpam) Name() string {
	return "caps_spam"
}

func (s CapsSpam) Check(p structs.Post) Result {
	letters := 0
	upper := 0
	for _, r := range p.Title + p.Content {
		if unicode.IsLetter(r) {
			letters++
			if unicode.IsUpper(r) {
				upper++
			}
		}
	}
	if letters < 10 {
		return Result{Allowed: true}
	}
	if upper*100/letters > s.MaxPercent {
		return Result{Allowed: false, Reason: "too many capital letters"}
	}
	return Result{Allowed: true}
}

type BlockedLinks struct {
	Domains []string
}

func (s BlockedLinks) Name() string {
	return "blocked_links"
}

func (s BlockedLinks) Check(p structs.Post) Result {
	text := strings.ToLower(p.MediaURL + " " + p.Content)
	for _, d := range s.Domains {
		if strings.Contains(text, d) {
			return Result{Allowed: false, Reason: "link to blocked site: " + d}
		}
	}
	return Result{Allowed: true}
}

type Strict struct {
	Strategies []ModerationStrategy
}

func (s Strict) Name() string {
	return "strict"
}

func (s Strict) Check(p structs.Post) Result {
	for _, st := range s.Strategies {
		res := st.Check(p)
		if !res.Allowed {
			return res
		}
	}
	return Result{Allowed: true}
}

// Context: uses one strategy, can change it at runtime.
type Moderator struct {
	strategy ModerationStrategy
}

func NewModerator(s ModerationStrategy) *Moderator {
	return &Moderator{strategy: s}
}

func (m *Moderator) SetStrategy(s ModerationStrategy) {
	m.strategy = s
}

func (m *Moderator) StrategyName() string {
	return m.strategy.Name()
}

func (m *Moderator) Check(p structs.Post) Result {
	return m.strategy.Check(p)
}

func (m *Moderator) Filter(posts []structs.Post) []structs.Post {
	result := []structs.Post{}
	for _, p := range posts {
		if m.Check(p).Allowed {
			result = append(result, p)
		}
	}
	return result
}

var bannedWords = BannedWords{Words: []string{"spam", "scam", "idiot", "stupid", "hate"}}
var capsSpam = CapsSpam{MaxPercent: 70}
var blockedLinks = BlockedLinks{Domains: []string{"bit.ly", "free-money", "casino"}}

var all = map[string]ModerationStrategy{
	"none":          AllowAll{},
	"banned_words":  bannedWords,
	"caps_spam":     capsSpam,
	"blocked_links": blockedLinks,
	"strict":        Strict{Strategies: []ModerationStrategy{bannedWords, capsSpam, blockedLinks}},
}

func Names() []string {
	return []string{"none", "banned_words", "caps_spam", "blocked_links", "strict"}
}

func ByName(name string) (ModerationStrategy, bool) {
	s, ok := all[name]
	return s, ok
}
