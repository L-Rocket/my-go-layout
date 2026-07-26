package postgrest

import (
	"context"
	"net/url"

	"github.com/L-Rocket/my-go-layout/internal/message"
)

// MessageStore implements message.Store over the "messages" table. IDs
// come from a DB sequence (bigserial), so Create reads the returned row
// back into the message.
type MessageStore struct {
	c *Client
}

func NewMessageStore(c *Client) *MessageStore {
	return &MessageStore{c: c}
}

func (s *MessageStore) Create(ctx context.Context, m *message.Message) error {
	var rows []message.Message
	if err := s.c.do(ctx, "POST", "/messages", m, &rows, "return=representation"); err != nil {
		return err
	}
	if len(rows) > 0 {
		m.ID, m.CreatedAt = rows[0].ID, rows[0].CreatedAt
	}
	return nil
}

func (s *MessageStore) ListBySession(ctx context.Context, sessionID string) ([]*message.Message, error) {
	var rows []*message.Message
	err := s.c.do(ctx, "GET", "/messages?session_id=eq."+url.QueryEscape(sessionID)+"&order=id.asc", nil, &rows, "")
	if err != nil {
		return nil, err
	}
	return rows, nil
}

var _ message.Store = (*MessageStore)(nil)
