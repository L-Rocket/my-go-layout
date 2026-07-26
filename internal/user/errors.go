package user

import "errors"

var (
	ErrNotFound      = errors.New("user: not found")
	ErrDuplicateName = errors.New("user: name already taken")
)
