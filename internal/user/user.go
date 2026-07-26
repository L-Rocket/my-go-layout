// Package user is the User domain module: account records only, no auth
// (this is a playground; identities are just names).
package user

import (
	"crypto/rand"
	"encoding/hex"
	"time"
)

// User is a registered account.
type User struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

// New builds a user with a uuid-style random id (128 bits, hex-encoded).
func New(name string) *User {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		panic("crypto/rand unavailable: " + err.Error())
	}
	return &User{
		ID:        "user-" + hex.EncodeToString(b),
		Name:      name,
		CreatedAt: time.Now().UTC(),
	}
}
