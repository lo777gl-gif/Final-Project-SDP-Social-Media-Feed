package observer

import (
	"final/internal/structs"
	"log"
	"sync"
)

type EventType string

const (
	PostReactionEvent    EventType = "post_reacted"
	CommentReactionEvent EventType = "comment_reacted"
	CommentCreatedEvent  EventType = "comment_created"
)

type Event struct {
	Type      EventType
	ActorID   int
	PostID    int
	CommentID int
}

type Observer interface {
	Notify(e Event) error
}

type notificationStore interface {
	InsertNotification(structs.Notification) error
	PostAuthorID(postID int) (int, error)
	CommentAuthorID(commentID int) (int, error)
}

type Subject struct {
	mu        sync.RWMutex
	wg        sync.WaitGroup
	observers []Observer
}

func NewSubject() *Subject {
	return &Subject{}
}

func (s *Subject) Subscribe(o Observer) {
	s.mu.Lock()
	s.observers = append(s.observers, o)
	s.mu.Unlock()
}

func (s *Subject) Publish(e Event) {
	s.mu.RLock()
	observers := append([]Observer(nil), s.observers...)
	s.mu.RUnlock()
	for _, observer := range observers {
		s.wg.Add(1)
		go func(o Observer) {
			defer s.wg.Done()
			defer func() {
				if r := recover(); r != nil {
					log.Printf("observer panicked")
				}
			}()
			if err := observer.Notify(e); err != nil {
				log.Printf("observer error: %v", err)
			}
		}(observer)
	}
}
