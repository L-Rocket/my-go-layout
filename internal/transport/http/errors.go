package http

import (
	"errors"
	nethttp "net/http"

	"github.com/gin-gonic/gin"

	"github.com/L-Rocket/my-go-layout/internal/message"
	"github.com/L-Rocket/my-go-layout/internal/session"
)

// handleError is the ONLY place domain errors become HTTP statuses.
// Handlers never pick status codes themselves, so the mapping stays
// consistent across every endpoint.
func handleError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, session.ErrNotFound):
		fail(c, nethttp.StatusNotFound, err.Error())
	case errors.Is(err, session.ErrNotOwner),
		errors.Is(err, message.ErrNotOwner):
		fail(c, nethttp.StatusForbidden, err.Error())
	case errors.Is(err, session.ErrAlreadyEnded):
		fail(c, nethttp.StatusConflict, err.Error())
	default:
		// unknown errors are 500 with a generic body — don't leak internals
		fail(c, nethttp.StatusInternalServerError, "internal error")
	}
}
