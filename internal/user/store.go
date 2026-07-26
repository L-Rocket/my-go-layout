package user

import "context"

// Store is the persistence contract for users, owned by the domain (DIP).
type Store interface {
	// Create persists a new user; returns ErrDuplicateName if the name is taken.
	Create(ctx context.Context, u *User) error
	// Get returns one user or ErrNotFound.
	Get(ctx context.Context, id string) (*User, error)
	// List returns all users, oldest first.
	List(ctx context.Context) ([]*User, error)
}
