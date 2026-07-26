// Package session is the Session domain module: types, business rules,
// and the storage contract (Store) it depends on. It knows nothing about
// HTTP, gRPC, or any concrete database.
package session

import (
	"crypto/rand"
	"encoding/hex"
	"time"
)

// Session lifecycle statuses.
const (
	StatusActive = "active"
	StatusEnded  = "ended"
)

// Session is a chat session owned by exactly one user.
type Session struct {
	ID        string     `json:"id"`
	UserID    string     `json:"user_id"`
	Title     string     `json:"title"`
	Status    string     `json:"status"`
	CreatedAt time.Time  `json:"created_at"`
	EndedAt   *time.Time `json:"ended_at,omitempty"`
}

// New builds an active session. The id is random (not sequential) so ids
// stay unguessable and storage backends never need to coordinate.
func New(userID, title string) *Session {
	return &Session{
		ID:        "sess-" + randHex(8), // 8 bytes = 16 hex chars
		UserID:    userID,
		Title:     title,
		Status:    StatusActive,
		CreatedAt: time.Now().UTC(),
	}
}

func randHex(n int) string {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		// crypto/rand only fails on a broken OS entropy source; there is no
		// sensible fallback, so crashing loudly is the honest behavior.
		panic("crypto/rand unavailable: " + err.Error())
	}
	return hex.EncodeToString(b)
}
