package postgrest

import (
	"context"
	"net/url"

	"github.com/L-Rocket/my-project/internal/session"
)

// SessionStore implements session.Store over the "sessions" table.
type SessionStore struct {
	c *Client
}

func NewSessionStore(c *Client) *SessionStore {
	return &SessionStore{c: c}
}

func (s *SessionStore) Create(ctx context.Context, sess *session.Session) error {
	// return=representation hands back the stored row; we ignore it since
	// all fields are client-generated.
	return s.c.do(ctx, "POST", "/sessions", sess, nil, "return=representation")
}

func (s *SessionStore) Get(ctx context.Context, id string) (*session.Session, error) {
	var rows []session.Session
	err := s.c.do(ctx, "GET", "/sessions?id=eq."+url.QueryEscape(id), nil, &rows, "")
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, session.ErrNotFound
	}
	return &rows[0], nil
}

func (s *SessionStore) ListByUser(ctx context.Context, userID string) ([]*session.Session, error) {
	var rows []*session.Session
	err := s.c.do(ctx, "GET", "/sessions?user_id=eq."+url.QueryEscape(userID)+"&order=created_at.asc", nil, &rows, "")
	if err != nil {
		return nil, err
	}
	return rows, nil
}

func (s *SessionStore) Update(ctx context.Context, sess *session.Session) error {
	// Only lifecycle fields are mutable; PATCH the row by id and require
	// a row to actually change (rows=len 0 would mean a vanished session).
	var rows []session.Session
	patch := map[string]any{"status": sess.Status, "ended_at": sess.EndedAt}
	err := s.c.do(ctx, "PATCH", "/sessions?id=eq."+url.QueryEscape(sess.ID), patch, &rows, "return=representation")
	if err != nil {
		return err
	}
	if len(rows) == 0 {
		return session.ErrNotFound
	}
	return nil
}

// compile-time guard: this must satisfy the domain contract
var _ session.Store = (*SessionStore)(nil)
