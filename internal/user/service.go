package user

import "context"

// Service holds the user business rules. Thin on purpose: the rules that
// exist (unique names) live here, not in transports or storage.
type Service struct {
	store Store
}

func NewService(store Store) *Service {
	return &Service{store: store}
}

// Create registers a user; duplicate names are rejected with ErrDuplicateName.
func (s *Service) Create(ctx context.Context, name string) (*User, error) {
	u := New(name)
	if err := s.store.Create(ctx, u); err != nil {
		return nil, err
	}
	return u, nil
}

// Get returns one user or ErrNotFound.
func (s *Service) Get(ctx context.Context, id string) (*User, error) {
	return s.store.Get(ctx, id)
}

// List returns all users, oldest first.
func (s *Service) List(ctx context.Context) ([]*User, error) {
	return s.store.List(ctx)
}
