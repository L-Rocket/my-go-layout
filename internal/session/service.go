package session

import (
	"context"
	"time"
)

// Service holds the session business rules. The message module consumes
// this (not Store) so those rules — ownership, lifecycle — are enforced
// in one place.
type Service struct {
	store Store
}

func NewService(store Store) *Service {
	return &Service{store: store}
}

// Create starts a new active session for a user. Existence of the user is
// the caller's concern (the transport checks it and returns 404), keeping
// this module free of any user dependency.
func (s *Service) Create(ctx context.Context, userID, title string) (*Session, error) {
	sess := New(userID, title)
	if err := s.store.Create(ctx, sess); err != nil {
		return nil, err
	}
	return sess, nil
}

// Get returns one session or ErrNotFound.
func (s *Service) Get(ctx context.Context, id string) (*Session, error) {
	return s.store.Get(ctx, id)
}

// ListByUser returns a user's sessions, oldest first.
func (s *Service) ListByUser(ctx context.Context, userID string) ([]*Session, error) {
	return s.store.ListByUser(ctx, userID)
}

// End marks a session ended. Only the owner may end it, and ending an
// already-ended session is an error (not a silent no-op) so clients get
// honest feedback.
func (s *Service) End(ctx context.Context, id, userID string) (*Session, error) {
	sess, err := s.store.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	if sess.UserID != userID {
		return nil, ErrNotOwner
	}
	if sess.Status == StatusEnded {
		return nil, ErrAlreadyEnded
	}
	now := timeNow()
	sess.Status = StatusEnded
	sess.EndedAt = &now
	if err := s.store.Update(ctx, sess); err != nil {
		return nil, err
	}
	return sess, nil
}

// timeNow is a seam so tests could freeze time if needed.
var timeNow = func() time.Time { return time.Now().UTC() }
