// Package test holds whole-service integration tests: the app is
// assembled here exactly as cmd/server/main.go assembles it (memory
// stores → domain services → transports) and driven over its PUBLIC API
// on real loopback ports. Unit tests live next to the code they cover;
// this directory is for end-to-end behavior only. Fakes belong only at
// true external boundaries — there are none here, since memory is the
// default backend.
package test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net"
	"net/http"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	reflectpb "google.golang.org/grpc/reflection/grpc_reflection_v1"

	"github.com/L-Rocket/my-go-layout/internal/infrastructure/memory"
	"github.com/L-Rocket/my-go-layout/internal/message"
	"github.com/L-Rocket/my-go-layout/internal/session"
	transportgrpc "github.com/L-Rocket/my-go-layout/internal/transport/grpc"
	transporthttp "github.com/L-Rocket/my-go-layout/internal/transport/http"
)

// discardLogger keeps the stack identical to main's (a *slog.Logger in
// Deps) without flooding the test output with per-request JSON logs.
func discardLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

// startHTTP assembles the whole HTTP stack in-process exactly as
// cmd/server/main.go does with the default (memory) backend, and serves
// it on a real loopback port. Returns the base URL.
func startHTTP(t *testing.T) string {
	t.Helper()

	sessionSvc := session.NewService(memory.NewSessionStore())
	messageSvc := message.NewService(memory.NewMessageStore(), sessionSvc)
	engine := transporthttp.NewEngine(transporthttp.Deps{
		Logger:   discardLogger(),
		Sessions: sessionSvc,
		Messages: messageSvc,
		Store:    "memory",
	})

	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("http listen: %v", err)
	}
	srv := transporthttp.NewServer(lis.Addr().String(), engine)
	go func() { _ = srv.Serve(lis) }()
	t.Cleanup(func() { _ = srv.Close() })
	return "http://" + lis.Addr().String()
}

// do sends one JSON request against the running server, asserts the
// status code, and returns the decoded body for further assertions.
func do(t *testing.T, method, url string, body any, wantStatus int) map[string]any {
	t.Helper()

	var rdr io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("marshal body: %v", err)
		}
		rdr = bytes.NewReader(b)
	}
	req, err := http.NewRequest(method, url, rdr)
	if err != nil {
		t.Fatalf("build request %s %s: %v", method, url, err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("%s %s: %v", method, url, err)
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("%s %s: read body: %v", method, url, err)
	}

	if resp.StatusCode != wantStatus {
		t.Fatalf("%s %s: status = %d, want %d (body: %s)", method, url, resp.StatusCode, wantStatus, raw)
	}
	var decoded map[string]any
	if len(raw) > 0 {
		if err := json.Unmarshal(raw, &decoded); err != nil {
			t.Fatalf("%s %s: decode body %q: %v", method, url, raw, err)
		}
	}
	return decoded
}

// TestHTTPSessionLifecycle drives the full public HTTP API through one
// realistic flow: create → messages → ownership rules → lifecycle rules.
func TestHTTPSessionLifecycle(t *testing.T) {
	base := startHTTP(t)

	// create a session
	created := do(t, http.MethodPost, base+"/api/sessions",
		map[string]string{"user_id": "user-ada", "title": "demo"}, http.StatusCreated)
	sess, _ := created["data"].(map[string]any)
	sessID, _ := sess["id"].(string)
	if sessID == "" {
		t.Fatalf("create session: no id in response: %v", created)
	}

	// post two messages
	do(t, http.MethodPost, base+"/api/sessions/"+sessID+"/messages",
		map[string]string{"user_id": "user-ada", "content": "first"}, http.StatusCreated)
	do(t, http.MethodPost, base+"/api/sessions/"+sessID+"/messages",
		map[string]string{"user_id": "user-ada", "content": "second"}, http.StatusCreated)

	// list them back in insertion order
	listed := do(t, http.MethodGet, base+"/api/sessions/"+sessID+"/messages?user_id=user-ada",
		nil, http.StatusOK)
	msgs, _ := listed["data"].([]any)
	if len(msgs) != 2 {
		t.Fatalf("list messages: got %d, want 2: %v", len(msgs), listed)
	}
	for i, want := range []string{"first", "second"} {
		m, _ := msgs[i].(map[string]any)
		if m["content"] != want {
			t.Errorf("message %d: content = %v, want %q", i, m["content"], want)
		}
	}

	// ownership: a non-owner cannot end the session
	do(t, http.MethodPost, base+"/api/sessions/"+sessID+"/end",
		map[string]string{"user_id": "user-eve"}, http.StatusForbidden)
	// the owner can
	do(t, http.MethodPost, base+"/api/sessions/"+sessID+"/end",
		map[string]string{"user_id": "user-ada"}, http.StatusOK)
	// lifecycle: posting into an ended session conflicts
	do(t, http.MethodPost, base+"/api/sessions/"+sessID+"/messages",
		map[string]string{"user_id": "user-ada", "content": "too late"}, http.StatusConflict)

	// error mapping: unknown session → 404, missing user_id → 400
	do(t, http.MethodPost, base+"/api/sessions/sess-nope/end",
		map[string]string{"user_id": "user-ada"}, http.StatusNotFound)
	do(t, http.MethodGet, base+"/api/sessions", nil, http.StatusBadRequest)

	// liveness reports the wired storage backend
	health := do(t, http.MethodGet, base+"/health", nil, http.StatusOK)
	if health["status"] != "up" || health["store"] != "memory" {
		t.Errorf("health: got %v, want status=up store=memory", health)
	}
}

// TestGRPCReflection verifies the gRPC transport boots on a real port
// and answers a reflection ListServices call, as main serves it too.
func TestGRPCReflection(t *testing.T) {
	srv := transportgrpc.NewServer(discardLogger())
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("grpc listen: %v", err)
	}
	go func() { _ = srv.Serve(lis) }()
	t.Cleanup(srv.Stop)

	conn, err := grpc.NewClient(lis.Addr().String(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("grpc dial: %v", err)
	}
	t.Cleanup(func() { _ = conn.Close() })

	stream, err := reflectpb.NewServerReflectionClient(conn).ServerReflectionInfo(context.Background())
	if err != nil {
		t.Fatalf("reflection stream: %v", err)
	}
	if err := stream.Send(&reflectpb.ServerReflectionRequest{
		MessageRequest: &reflectpb.ServerReflectionRequest_ListServices{ListServices: "*"},
	}); err != nil {
		t.Fatalf("reflection send: %v", err)
	}
	resp, err := stream.Recv()
	if err != nil {
		t.Fatalf("reflection recv: %v", err)
	}
	if services := resp.GetListServicesResponse().GetService(); len(services) == 0 {
		t.Fatal("reflection: expected at least one registered service, got none")
	}
}
