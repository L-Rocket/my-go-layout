package message

import "context"

// Store is the persistence contract for messages, owned by the domain (DIP).
type Store interface {
	// Create persists a message, assigning ID and CreatedAt.
	Create(ctx context.Context, m *Message) error
	// ListBySession returns a session's messages ordered by ID (post order).
	ListBySession(ctx context.Context, sessionID string) ([]*Message, error)
}
