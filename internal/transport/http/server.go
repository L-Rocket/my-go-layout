// Package http is the HTTP transport: gin engine, routes, middleware,
// and the single domain-error→status mapping. No business rules here.
package http

import (
	"net/http"
	"time"
)

// NewServer wraps the gin engine in an *http.Server. ReadHeaderTimeout
// guards against slowloris; there is deliberately NO blanket WriteTimeout
// because it would kill long-lived streaming responses (SSE-safe rule).
func NewServer(addr string, handler http.Handler) *http.Server {
	return &http.Server{
		Addr:              addr,
		Handler:           handler,
		ReadHeaderTimeout: 10 * time.Second,
	}
}
