package session

import "context"

// Store is the persistence contract for sessions. The domain owns this
// interface (dependency inversion): infrastructure packages implement it,
// the domain never imports infrastructure.
type Store interface {
	// Create persists a new session.
	Create(ctx context.Context, s *Session) error
	// Get returns one session or ErrNotFound.
	Get(ctx context.Context, id string) (*Session, error)
	// ListByUser returns a user's sessions, oldest first.
	ListByUser(ctx context.Context, userID string) ([]*Session, error)
	// Update overwrites mutable fields (Status, EndedAt).
	Update(ctx context.Context, s *Session) error
}
