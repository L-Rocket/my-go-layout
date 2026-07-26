package memory

import (
	"context"
	"sync"
	"time"

	"github.com/L-Rocket/my-go-layout/internal/message"
)

// MessageStore is an in-memory message.Store. IDs auto-increment per
// session (like a per-session sequence) so post order is just ID order.
type MessageStore struct {
	mu      sync.Mutex
	byID    map[string][]*message.Message // sessionID → messages in order
	nextSeq map[string]int64              // sessionID → last issued ID
}

func NewMessageStore() *MessageStore {
	return &MessageStore{
		byID:    make(map[string][]*message.Message),
		nextSeq: make(map[string]int64),
	}
}

func (s *MessageStore) Create(_ context.Context, m *message.Message) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.nextSeq[m.SessionID]++
	cp := *m
	cp.ID = s.nextSeq[m.SessionID]
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
