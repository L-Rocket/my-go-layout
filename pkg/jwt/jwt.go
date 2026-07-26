// Package jwt is a stdlib-only HS256 JSON Web Token implementation
// (crypto/hmac + sha256 + base64url). It exists to show the mechanism;
// production code should use a vetted library.
package jwt

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

var (
	ErrMalformed    = errors.New("jwt: malformed token")
	ErrBadSignature = errors.New("jwt: signature mismatch")
	ErrExpired      = errors.New("jwt: token expired")
)

type claims struct {
	Sub string `json:"sub"`
	Exp int64  `json:"exp"`
	Iat int64  `json:"iat"`
}

var b64 = base64.RawURLEncoding

// Sign issues a token for subject that expires after ttl.
func Sign(secret, subject string, ttl time.Duration) (string, error) {
	now := time.Now()
	header := b64.EncodeToString([]byte(`{"alg":"HS256","typ":"JWT"}`))
	payload, err := json.Marshal(claims{Sub: subject, Iat: now.Unix(), Exp: now.Add(ttl).Unix()})
	if err != nil {
		return "", err
	}
	body := header + "." + b64.EncodeToString(payload)
	return body + "." + b64.EncodeToString(mac(secret, body)), nil
}

// Verify checks signature and expiry, returning the subject.
func Verify(secret, token string) (string, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return "", ErrMalformed
	}
	body := parts[0] + "." + parts[1]
	if !hmac.Equal(mac(secret, body), mustDecode(parts[2])) {
		return "", ErrBadSignature
	}
	payload, err := b64.DecodeString(parts[1])
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrMalformed, err)
	}
	var c claims
	if err := json.Unmarshal(payload, &c); err != nil {
		return "", fmt.Errorf("%w: %v", ErrMalformed, err)
	}
	if time.Now().Unix() >= c.Exp {
		return "", ErrExpired
	}
	return c.Sub, nil
}

func mac(secret, body string) []byte {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(body))
	return h.Sum(nil)
}

// mustDecode only feeds hmac.Equal; a bad encoding simply never matches.
func mustDecode(s string) []byte {
	b, _ := b64.DecodeString(s)
	return b
}
