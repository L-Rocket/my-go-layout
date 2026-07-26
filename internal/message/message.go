// Package message is the Message domain module: chat messages inside a
// session. It enforces session lifecycle rules by depending on
// session.Service, never on session storage.
package message

import "time"

// Message is one chat message posted by a user into a session.
type Message struct {
	ID        int64     `json:"id"`
	SessionID string    `json:"session_id"`
	UserID    string    `json:"user_id"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}
