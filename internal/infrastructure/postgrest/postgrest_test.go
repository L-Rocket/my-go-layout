package postgrest_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/L-Rocket/my-go-layout/internal/infrastructure/postgrest"
	"github.com/L-Rocket/my-go-layout/internal/message"
	"github.com/L-Rocket/my-go-layout/internal/session"
)

// fakeSupabase records the last request and replies with the canned body.
type fakeSupabase struct {
	srv        *httptest.Server
	lastMethod string
	lastPath   string
	lastAPIKey string
	lastAuth   string
}

func newFake(t *testing.T, status int, body string) *fakeSupabase {
	f := &fakeSupabase{}
	f.srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		f.lastMethod = r.Method
		f.lastPath = r.URL.RequestURI()
		f.lastAPIKey = r.Header.Get("apikey")
		f.lastAuth = r.Header.Get("Authorization")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(f.srv.Close)
	return f
}

func (f *fakeSupabase) assertAuth(t *testing.T) {
	t.Helper()
	if f.lastAPIKey != "svc-key" || f.lastAuth != "Bearer svc-key" {
		t.Fatalf("auth headers = %q / %q", f.lastAPIKey, f.lastAuth)
	}
}

func TestSessionStoreGet(t *testing.T) {
	f := newFake(t, http.StatusOK, `[{"id":"sess-1","user_id":"user-1","title":"t","status":"active","created_at":"2026-01-01T00:00:00Z"}]`)
	store := postgrest.NewSessionStore(postgrest.New(f.srv.URL, "svc-key"))

	s, err := store.Get(context.Background(), "sess-1")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if s.ID != "sess-1" || s.Status != session.StatusActive {
		t.Fatalf("unexpected session: %+v", s)
	}
	if f.lastMethod != "GET" || f.lastPath != "/rest/v1/sessions?id=eq.sess-1" {
		t.Fatalf("request = %s %s", f.lastMethod, f.lastPath)
	}
	f.assertAuth(t)
}

func TestSessionStoreGetNotFound(t *testing.T) {
	f := newFake(t, http.StatusOK, `[]`)
	store := postgrest.NewSessionStore(postgrest.New(f.srv.URL, "svc-key"))
	if _, err := store.Get(context.Background(), "sess-nope"); err != session.ErrNotFound {
		t.Fatalf("err = %v, want session.ErrNotFound", err)
	}
}

func TestMessageStoreListBySession(t *testing.T) {
	f := newFake(t, http.StatusOK, `[
		{"id":1,"session_id":"sess-1","user_id":"user-1","content":"hi","created_at":"2026-01-01T00:00:00Z"},
		{"id":2,"session_id":"sess-1","user_id":"user-1","content":"yo","created_at":"2026-01-01T00:00:01Z"}
	]`)
	store := postgrest.NewMessageStore(postgrest.New(f.srv.URL, "svc-key"))

	ms, err := store.ListBySession(context.Background(), "sess-1")
	if err != nil {
		t.Fatalf("ListBySession: %v", err)
	}
	if len(ms) != 2 || ms[0].Content != "hi" || ms[1].ID != 2 {
		t.Fatalf("unexpected messages: %+v", ms)
	}
	// PostgREST order param must request id order
	if f.lastMethod != "GET" || !strings.Contains(f.lastPath, "/rest/v1/messages?session_id=eq.sess-1") || !strings.Contains(f.lastPath, "order=id.asc") {
		t.Fatalf("request = %s %s", f.lastMethod, f.lastPath)
	}
	f.assertAuth(t)
}

// Create reads the DB-assigned ID back from Prefer: return=representation.
func TestMessageStoreCreateReadsBackID(t *testing.T) {
	f := newFake(t, http.StatusCreated, `[{"id":7,"session_id":"sess-1","user_id":"user-1","content":"hi","created_at":"2026-01-01T00:00:00Z"}]`)
	store := postgrest.NewMessageStore(postgrest.New(f.srv.URL, "svc-key"))

	msg := &message.Message{SessionID: "sess-1", UserID: "user-1", Content: "hi"}
	if err := store.Create(context.Background(), msg); err != nil {
		t.Fatalf("Create: %v", err)
	}
	if msg.ID != 7 {
		t.Fatalf("ID = %d, want 7 (read back from representation)", msg.ID)
	}
	if f.lastMethod != "POST" || f.lastPath != "/rest/v1/messages" {
		t.Fatalf("request = %s %s", f.lastMethod, f.lastPath)
	}
	f.assertAuth(t)
}
