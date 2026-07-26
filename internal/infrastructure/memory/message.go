package memory

import (
	"context"
	"sync"
	"time"

	"github.com/L-Rocket/my-go-layout/internal/message"
)

// MessageStore is an in-memory message.Store. IDs come from one global
// counter (matching the bigserial semantics of the PostgREST backend) so
// post order is just ID order and both backends behave identically.
type MessageStore struct {
	mu   sync.Mutex
	byID map[string][]*message.Message // sessionID → messages in order
	next int64                         // last issued ID, across all sessions
}

func NewMessageStore() *MessageStore {
	return &MessageStore{byID: make(map[string][]*message.Message)}
}

func (s *MessageStore) Create(_ context.Context, m *message.Message) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.next++
	cp := *m
	cp.ID = s.next
	cp.CreatedAt = time.Now().UTC()
	s.byID[m.SessionID] = append(s.byID[m.SessionID], &cp)
	m.ID, m.CreatedAt = cp.ID, cp.CreatedAt
	return nil
}

func (s *MessageStore) ListBySession(_ context.Context, sessionID string) ([]*message.Message, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	src := s.byID[sessionID]
	out := make([]*message.Message, 0, len(src))
	for _, m := range src {
		cp := *m
		out = append(out, &cp)
	}
	return out, nil
}
