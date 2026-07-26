package memory

import (
	"context"
	"sync"

	"github.com/L-Rocket/my-project/internal/session"
)

// SessionStore is an in-memory session.Store. Safe for concurrent use;
// data vanishes on restart, which is exactly what local runs want.
type SessionStore struct {
	mu   sync.RWMutex
	byID map[string]*session.Session
}

func NewSessionStore() *SessionStore {
	return &SessionStore{byID: make(map[string]*session.Session)}
}

func (s *SessionStore) Create(_ context.Context, sess *session.Session) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	cp := *sess // copy so callers can't mutate stored state behind the lock
	s.byID[sess.ID] = &cp
	return nil
}

func (s *SessionStore) Get(_ context.Context, id string) (*session.Session, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	sess, ok := s.byID[id]
	if !ok {
		return nil, session.ErrNotFound
	}
	cp := *sess
	return &cp, nil
}

func (s *SessionStore) ListByUser(_ context.Context, userID string) ([]*session.Session, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var out []*session.Session
	for _, sess := range s.byID {
		if sess.UserID == userID {
			cp := *sess
			out = append(out, &cp)
		}
	}
	// map iteration order is random; sort for a stable "oldest first" contract
	sortSessions(out)
	return out, nil
}

func (s *SessionStore) Update(_ context.Context, sess *session.Session) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.byID[sess.ID]; !ok {
		return session.ErrNotFound
	}
	cp := *sess
	s.byID[sess.ID] = &cp
	return nil
}
