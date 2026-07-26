package memory

import (
	"context"
	"sync"

	"github.com/L-Rocket/my-project/internal/user"
)

// UserStore is an in-memory user.Store. Name uniqueness is enforced here
// (a unique index would do it in Postgres).
type UserStore struct {
	mu     sync.RWMutex
	byID   map[string]*user.User
	byName map[string]string // name → id
}

func NewUserStore() *UserStore {
	return &UserStore{byID: make(map[string]*user.User), byName: make(map[string]string)}
}

func (s *UserStore) Create(_ context.Context, u *user.User) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, taken := s.byName[u.Name]; taken {
		return user.ErrDuplicateName
	}
	cp := *u
	s.byID[u.ID] = &cp
	s.byName[u.Name] = u.ID
	return nil
}

func (s *UserStore) Get(_ context.Context, id string) (*user.User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	u, ok := s.byID[id]
	if !ok {
		return nil, user.ErrNotFound
	}
	cp := *u
	return &cp, nil
}

func (s *UserStore) List(_ context.Context) ([]*user.User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]*user.User, 0, len(s.byID))
	for _, u := range s.byID {
		cp := *u
		out = append(out, &cp)
	}
	sortUsers(out)
	return out, nil
}
