package message_test

import (
	"context"
	"errors"
	"testing"

	"github.com/L-Rocket/my-project/internal/infrastructure/memory"
	"github.com/L-Rocket/my-project/internal/message"
	"github.com/L-Rocket/my-project/internal/session"
)

// real session service over memory storage — the cross-module wiring
// under test is exactly what main does.
func setup() (*message.Service, *session.Service) {
	sessionSvc := session.NewService(memory.NewSessionStore())
	return message.NewService(memory.NewMessageStore(), sessionSvc), sessionSvc
}

func TestPostThenListInOrder(t *testing.T) {
	msgSvc, sessSvc := setup()
	ctx := context.Background()
	s, _ := sessSvc.Create(ctx, "user-a", "chat")

	for _, content := range []string{"one", "two", "three"} {
		if _, err := msgSvc.Post(ctx, s.ID, "user-a", content); err != nil {
			t.Fatalf("Post %q: %v", content, err)
		}
	}

	ms, err := msgSvc.ListBySession(ctx, s.ID, "user-a")
	if err != nil {
		t.Fatalf("ListBySession: %v", err)
	}
	if len(ms) != 3 {
		t.Fatalf("got %d messages, want 3", len(ms))
	}
	for i, want := range []string{"one", "two", "three"} {
		if ms[i].Content != want || ms[i].ID != int64(i+1) {
			t.Fatalf("message %d = %+v, want content %q id %d", i, ms[i], want, i+1)
		}
	}
}

func TestPostToEndedSessionRejected(t *testing.T) {
	msgSvc, sessSvc := setup()
	ctx := context.Background()
	s, _ := sessSvc.Create(ctx, "user-a", "chat")
	if _, err := sessSvc.End(ctx, s.ID, "user-a"); err != nil {
		t.Fatalf("End: %v", err)
	}
	if _, err := msgSvc.Post(ctx, s.ID, "user-a", "too late"); !errors.Is(err, session.ErrAlreadyEnded) {
		t.Fatalf("err = %v, want session.ErrAlreadyEnded", err)
	}
}

func TestPostByNonOwnerRejected(t *testing.T) {
	msgSvc, sessSvc := setup()
	ctx := context.Background()
	s, _ := sessSvc.Create(ctx, "user-a", "chat")

	if _, err := msgSvc.Post(ctx, s.ID, "user-b", "intruder"); !errors.Is(err, message.ErrNotOwner) {
		t.Fatalf("err = %v, want message.ErrNotOwner", err)
	}
	if _, err := msgSvc.ListBySession(ctx, s.ID, "user-b"); !errors.Is(err, message.ErrNotOwner) {
		t.Fatalf("err = %v, want message.ErrNotOwner", err)
	}
}

func TestPostToMissingSession(t *testing.T) {
	msgSvc, _ := setup()
	if _, err := msgSvc.Post(context.Background(), "sess-nope", "user-a", "hi"); !errors.Is(err, session.ErrNotFound) {
		t.Fatalf("err = %v, want session.ErrNotFound", err)
	}
}
