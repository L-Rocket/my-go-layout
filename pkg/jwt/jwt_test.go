package jwt

import (
	"errors"
	"testing"
	"time"
)

func TestSignVerifyRoundTrip(t *testing.T) {
	tok, err := Sign("s3cret", "user-123", time.Minute)
	if err != nil {
		t.Fatalf("Sign: %v", err)
	}
	sub, err := Verify("s3cret", tok)
	if err != nil {
		t.Fatalf("Verify: %v", err)
	}
	if sub != "user-123" {
		t.Fatalf("subject = %q, want %q", sub, "user-123")
	}
}

func TestExpiredTokenRejected(t *testing.T) {
	tok, err := Sign("s3cret", "user-123", -time.Minute) // already expired
	if err != nil {
		t.Fatalf("Sign: %v", err)
	}
	if _, err := Verify("s3cret", tok); !errors.Is(err, ErrExpired) {
		t.Fatalf("err = %v, want ErrExpired", err)
	}
}

func TestWrongSecretRejected(t *testing.T) {
	tok, err := Sign("s3cret", "user-123", time.Minute)
	if err != nil {
		t.Fatalf("Sign: %v", err)
	}
	if _, err := Verify("wrong", tok); !errors.Is(err, ErrBadSignature) {
		t.Fatalf("err = %v, want ErrBadSignature", err)
	}
}

func TestMalformedTokenRejected(t *testing.T) {
	if _, err := Verify("s3cret", "not.a.valid.jwt.token"); !errors.Is(err, ErrMalformed) {
		t.Fatalf("err = %v, want ErrMalformed", err)
	}
}
