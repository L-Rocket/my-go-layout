package http

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// allowedOrigin is the local Vite dev server. Hand-rolled CORS (instead
// of gin-contrib/cors) keeps the dependency list small and the rules obvious.
const allowedOrigin = "http://localhost:5173"

// RequestLogger logs one line per request via the app logger. Hand-rolled
// (not gin.Logger) so output goes through slog JSON like everything else.
func RequestLogger(log *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		log.Info("http request",
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"status", c.Writer.Status(),
			"latency_ms", time.Since(start).Milliseconds(),
		)
	}
}

// CORS answers preflights and stamps the allow headers for the one
// trusted origin. No wildcards: this API will carry credentials later.
func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		h := c.Writer.Header()
		h.Set("Access-Control-Allow-Origin", allowedOrigin)
		h.Set("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS")
		h.Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}
