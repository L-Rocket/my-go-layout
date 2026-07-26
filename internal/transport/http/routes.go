package http

import (
	"log/slog"
	nethttp "net/http"

	"github.com/gin-gonic/gin"

	"github.com/L-Rocket/my-go-layout/internal/message"
	"github.com/L-Rocket/my-go-layout/internal/session"
)

// Deps is everything the routes need, injected from main so this package
// stays assembly-free.
type Deps struct {
	Logger   *slog.Logger
	Sessions *session.Service
	Messages *message.Service
	Store    string // "memory" | "postgrest", for the health payload
}

// NewEngine builds the gin engine: middleware first, then the route
// table. Kept separate from NewServer so tests can run the stack in-process.
func NewEngine(d Deps) *gin.Engine {
	r := gin.New()
	r.Use(RequestLogger(d.Logger), CORS())

	// Liveness: never fails — a down dependency must not hide the process
	// from the orchestrator. Reports which storage main wired in.
	r.GET("/health", func(c *gin.Context) {
		c.JSON(nethttp.StatusOK, gin.H{"status": "up", "store": d.Store})
	})

	api := r.Group("/api")
	api.POST("/sessions", createSession(d))
	api.GET("/sessions", listSessions(d))
	api.POST("/sessions/:id/end", endSession(d))
	api.POST("/sessions/:id/messages", postMessage(d))
	api.GET("/sessions/:id/messages", listMessages(d))
	return r
}

// --- request bodies: validation tags make gin reject bad input at binding ---

type createSessionReq struct {
	UserID string `json:"user_id" binding:"required"`
	Title  string `json:"title" binding:"required"`
}

type endSessionReq struct {
	UserID string `json:"user_id" binding:"required"`
}

type postMessageReq struct {
	UserID  string `json:"user_id" binding:"required"`
	Content string `json:"content" binding:"required"`
}

// --- handlers: parse → call domain service → map error → envelope ---
// No business rules here: ownership and lifecycle live in the services.

func createSession(d Deps) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req createSessionReq
		if !bindJSON(c, &req) {
			return
		}
		sess, err := d.Sessions.Create(c.Request.Context(), req.UserID, req.Title)
		if err != nil {
			handleError(c, err)
			return
		}
		ok(c, nethttp.StatusCreated, sess)
	}
}

func listSessions(d Deps) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.Query("user_id")
		if userID == "" {
			fail(c, nethttp.StatusBadRequest, "user_id query parameter is required")
			return
		}
		ss, err := d.Sessions.ListByUser(c.Request.Context(), userID)
		if err != nil {
			handleError(c, err)
			return
		}
		ok(c, nethttp.StatusOK, ss)
	}
}

func endSession(d Deps) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req endSessionReq
		if !bindJSON(c, &req) {
			return
		}
		sess, err := d.Sessions.End(c.Request.Context(), c.Param("id"), req.UserID)
		if err != nil {
			handleError(c, err)
			return
		}
		ok(c, nethttp.StatusOK, sess)
	}
}

func postMessage(d Deps) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req postMessageReq
		if !bindJSON(c, &req) {
			return
		}
		m, err := d.Messages.Post(c.Request.Context(), c.Param("id"), req.UserID, req.Content)
		if err != nil {
			handleError(c, err)
			return
		}
		ok(c, nethttp.StatusCreated, m)
	}
}

func listMessages(d Deps) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.Query("user_id")
		if userID == "" {
			fail(c, nethttp.StatusBadRequest, "user_id query parameter is required")
			return
		}
		ms, err := d.Messages.ListBySession(c.Request.Context(), c.Param("id"), userID)
		if err != nil {
			handleError(c, err)
			return
		}
		ok(c, nethttp.StatusOK, ms)
	}
}
