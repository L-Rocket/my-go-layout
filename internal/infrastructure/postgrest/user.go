package postgrest

import (
	"context"
	"errors"
	"net/http"
	"net/url"

	"github.com/L-Rocket/my-go-layout/internal/user"
)

// UserStore implements user.Store over the "users" table. The table is
// expected to have a unique index on name.
type UserStore struct {
	c *Client
}

func NewUserStore(c *Client) *UserStore {
	return &UserStore{c: c}
}

func (s *UserStore) Create(ctx context.Context, u *user.User) error {
	err := s.c.do(ctx, "POST", "/users", u, nil, "return=representation")
	var he *HTTPError
	if errors.As(err, &he) && he.Status == http.StatusConflict {
		// unique-index violation on name
		return user.ErrDuplicateName
	}
	return err
}

func (s *UserStore) Get(ctx context.Context, id string) (*user.User, error) {
	var rows []user.User
	err := s.c.do(ctx, "GET", "/users?id=eq."+url.QueryEscape(id), nil, &rows, "")
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, user.ErrNotFound
	}
	return &rows[0], nil
}

func (s *UserStore) List(ctx context.Context) ([]*user.User, error) {
	var rows []*user.User
	err := s.c.do(ctx, "GET", "/users?order=created_at.asc", nil, &rows, "")
	if err != nil {
		return nil, err
	}
	return rows, nil
}

var _ user.Store = (*UserStore)(nil)
