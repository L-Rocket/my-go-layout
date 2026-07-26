package message

import (
	"context"

	"github.com/L-Rocket/my-project/internal/session"
)

// SessionService is the slice of session.Service this module needs. It is
// declared as a narrow interface (not the concrete type) so tests can fake
// it and the dependency direction stays explicit: message → session rules.
// The concrete *session.Service satisfies it.
type SessionService interface {
	Get(ctx context.Context, id string) (*session.Session, error)
}

// Service posts and lists messages. It deliberately talks to
// session.Service rather than session.Store: the "only the owner, and only
// while active" rules are session business logic and must not be
// re-implemented (or bypassed) here.
type Service struct {
	store    Store
	sessions SessionService
}

func NewService(store Store, sessions SessionService) *Service {
	return &Service{store: store, sessions: sessions}
}

// Post appends a message to a session. Fails with session.ErrNotFound if
// the session is gone, ErrNotOwner if another user posts, and
// session.ErrAlreadyEnded once the session is ended.
func (s *Service) Post(ctx context.Context, sessionID, userID, content string) (*Message, error) {
	sess, err := s.sessions.Get(ctx, sessionID)
	if err != nil {
		return nil, err
	}
	if sess.UserID != userID {
		return nil, ErrNotOwner
	}
	if sess.Status == session.StatusEnded {
		return nil, session.ErrAlreadyEnded
	}
	m := &Message{SessionID: sessionID, UserID: userID, Content: content}
	if err := s.store.Create(ctx, m); err != nil {
		return nil, err
	}
	return m, nil
}

// ListBySession returns a session's messages in post order; only the
// session owner may read them.
func (s *Service) ListBySession(ctx context.Context, sessionID, userID string) ([]*Message, error) {
	sess, err := s.sessions.Get(ctx, sessionID)
	if err != nil {
		return nil, err
	}
	if sess.UserID != userID {
		return nil, ErrNotOwner
	}
	return s.store.ListBySession(ctx, sessionID)
}
