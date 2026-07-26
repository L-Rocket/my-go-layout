package session_test

import (
	"context"
	"errors"
	"testing"

	"github.com/L-Rocket/my-project/internal/infrastructure/memory"
	"github.com/L-Rocket/my-project/internal/session"
)

func newService() *session.Service {
	return session.NewService(memory.NewSessionStore())
}

func TestCreateGetList(t *testing.T) {
	svc := newService()
	ctx := context.Background()

	s1, err := svc.Create(ctx, "user-a", "first")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	s2, err := svc.Create(ctx, "user-a", "second")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if _, err := svc.Create(ctx, "user-b", "other user"); err != nil {
		t.Fatalf("Create: %v", err)
	}

	got, err := svc.Get(ctx, s1.ID)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got.Title != "first" || got.Status != session.StatusActive {
		t.Fatalf("unexpected session: %+v", got)
	}

	list, err := svc.ListByUser(ctx, "user-a")
	if err != nil {
		t.Fatalf("ListByUser: %v", err)
	}
	if len(list) != 2 || list[0].ID != s1.ID || list[1].ID != s2.ID {
		t.Fatalf("unexpected list: %+v", list)
	}

	if _, err := svc.Get(ctx, "sess-missing"); !errors.Is(err, session.ErrNotFound) {
		t.Fatalf("err = %v, want ErrNotFound", err)
	}
}

func TestEnd(t *testing.T) {
	svc := newService()
	ctx := context.Background()
	s, err := svc.Create(ctx, "user-a", "chat")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	ended, err := svc.End(ctx, s.ID, "user-a")
	if err != nil {
		t.Fatalf("End: %v", err)
	}
	if ended.Status != session.StatusEnded || ended.EndedAt == nil {
		t.Fatalf("session not ended: %+v", ended)
	}
}

func TestEndNotOwner(t *testing.T) {
	svc := newService()
	ctx := context.Background()
	s, _ := svc.Create(ctx, "user-a", "chat")

	if _, err := svc.End(ctx, s.ID, "user-b"); !errors.Is(err, session.ErrNotOwner) {
		t.Fatalf("err = %v, want ErrNotOwner", err)
	}
}

func TestDoubleEnd(t *testing.T) {
	svc := newService()
	ctx := context.Background()
	s, _ := svc.Create(ctx, "user-a", "chat")
	if _, err := svc.End(ctx, s.ID, "user-a"); err != nil {
		t.Fatalf("first End: %v", err)
	}
	if _, err := svc.End(ctx, s.ID, "user-a"); !errors.Is(err, session.ErrAlreadyEnded) {
		t.Fatalf("err = %v, want ErrAlreadyEnded", err)
	}
}
