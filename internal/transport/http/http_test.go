package http_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/L-Rocket/my-project/internal/infrastructure/memory"
	"github.com/L-Rocket/my-project/internal/message"
	"github.com/L-Rocket/my-project/internal/session"
	transporthttp "github.com/L-Rocket/my-project/internal/transport/http"
	"github.com/L-Rocket/my-project/internal/user"
)

// full stack: real services over memory stores, exactly as main wires them
func newTestServer() *httptest.Server {
	gin.SetMode(gin.TestMode)
	sessionSvc := session.NewService(memory.NewSessionStore())
	engine := transporthttp.NewEngine(transporthttp.Deps{
		Logger:   slog.New(slog.NewTextHandler(io.Discard, nil)),
		Users:    user.NewService(memory.NewUserStore()),
		Sessions: sessionSvc,
		Messages: message.NewService(memory.NewMessageStore(), sessionSvc),
		Store:    "memory",
	})
	return httptest.NewServer(engine)
}

// envelope mirrors pkg/response: {"data": ...} / {"error": ...}
type envelope struct {
	Data  json.RawMessage `json:"data"`
	Error string          `json:"error"`
}

func do(t *testing.T, method, url string, body any) (int, envelope) {
	t.Helper()
	var rdr io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			t.Fatal(err)
		}
		rdr = bytes.NewReader(b)
	}
	req, err := http.NewRequest(method, url, rdr)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	var env envelope
	if err := json.NewDecoder(resp.Body).Decode(&env); err != nil {
		t.Fatalf("decode envelope: %v", err)
	}
	return resp.StatusCode, env
}

func TestHappyPathChain(t *testing.T) {
	srv := newTestServer()
	defer srv.Close()

	// create user
	code, env := do(t, "POST", srv.URL+"/api/users", map[string]string{"name": "ada"})
	if code != http.StatusCreated {
		t.Fatalf("create user: status %d, err %q", code, env.Error)
	}
	var u user.User
	if err := json.Unmarshal(env.Data, &u); err != nil {
		t.Fatal(err)
	}

	// create session for that user
	code, env = do(t, "POST", srv.URL+"/api/sessions", map[string]string{"user_id": u.ID, "title": "demo"})
	if code != http.StatusCreated {
		t.Fatalf("create session: status %d, err %q", code, env.Error)
	}
	var s session.Session
	if err := json.Unmarshal(env.Data, &s); err != nil {
		t.Fatal(err)
	}

	// post two messages
	for _, content := range []string{"hello", "world"} {
		code, env = do(t, "POST", srv.URL+"/api/sessions/"+s.ID+"/messages",
			map[string]string{"user_id": u.ID, "content": content})
		if code != http.StatusCreated {
			t.Fatalf("post message: status %d, err %q", code, env.Error)
		}
	}

	// list messages back, in order
	code, env = do(t, "GET", fmt.Sprintf("%s/api/sessions/%s/messages?user_id=%s", srv.URL, s.ID, u.ID), nil)
	if code != http.StatusOK {
		t.Fatalf("list messages: status %d, err %q", code, env.Error)
	}
	var ms []message.Message
	if err := json.Unmarshal(env.Data, &ms); err != nil {
		t.Fatal(err)
	}
	if len(ms) != 2 || ms[0].Content != "hello" || ms[1].Content != "world" {
		t.Fatalf("unexpected messages: %+v", ms)
	}

	// health (not wrapped in the envelope — it answers raw gin.H)
	resp, err := http.Get(srv.URL + "/health")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	var h map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&h); err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK || h["status"] != "up" || h["store"] != "memory" {
		t.Fatalf("health = %d %+v", resp.StatusCode, h)
	}
}

func TestForbiddenNotOwner(t *testing.T) {
	srv := newTestServer()
	defer srv.Close()

	_, env := do(t, "POST", srv.URL+"/api/users", map[string]string{"name": "owner"})
	var owner user.User
	if err := json.Unmarshal(env.Data, &owner); err != nil {
		t.Fatal(err)
	}
	_, env = do(t, "POST", srv.URL+"/api/sessions", map[string]string{"user_id": owner.ID, "title": "private"})
	var s session.Session
	if err := json.Unmarshal(env.Data, &s); err != nil {
		t.Fatal(err)
	}

	code, env := do(t, "POST", srv.URL+"/api/sessions/"+s.ID+"/end", map[string]string{"user_id": "user-someone-else"})
	if code != http.StatusForbidden {
		t.Fatalf("status = %d, want 403 (err %q)", code, env.Error)
	}
}

func TestNotFoundMissingUser(t *testing.T) {
	srv := newTestServer()
	defer srv.Close()

	code, env := do(t, "POST", srv.URL+"/api/sessions", map[string]string{"user_id": "user-ghost", "title": "x"})
	if code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404 (err %q)", code, env.Error)
	}
}

func TestConflictDuplicateName(t *testing.T) {
	srv := newTestServer()
	defer srv.Close()

	do(t, "POST", srv.URL+"/api/users", map[string]string{"name": "ada"})
	code, env := do(t, "POST", srv.URL+"/api/users", map[string]string{"name": "ada"})
	if code != http.StatusConflict {
		t.Fatalf("status = %d, want 409 (err %q)", code, env.Error)
	}
}
