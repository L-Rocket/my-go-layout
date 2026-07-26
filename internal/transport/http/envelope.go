package http

import (
	nethttp "net/http"

	"github.com/gin-gonic/gin"
)

// The single JSON envelope for the HTTP API: success is {"data": ...},
// failure is {"error": "..."}. One shape keeps clients trivial.

// ok writes a success envelope with the given status code.
func ok(c *gin.Context, code int, data any) {
	c.JSON(code, gin.H{"data": data})
}

// fail writes a failure envelope with the given status code.
func fail(c *gin.Context, code int, message string) {
	c.JSON(code, gin.H{"error": message})
}

// bindJSON binds the request body into v. On failure it writes a 400
// envelope and returns false — handlers should just `return`.
func bindJSON(c *gin.Context, v any) bool {
	if err := c.ShouldBindJSON(v); err != nil {
		fail(c, nethttp.StatusBadRequest, "invalid request body: "+err.Error())
		return false
	}
	return true
}
