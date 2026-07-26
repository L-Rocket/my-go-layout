package session

import "errors"

// Sentinel errors: transports map these to status codes in exactly one
// place (internal/transport/http/errors.go), so domain code never imports
// HTTP concepts.
var (
	ErrNotFound     = errors.New("session: not found")
	ErrNotOwner     = errors.New("session: not owned by user")
	ErrAlreadyEnded = errors.New("session: already ended")
)
