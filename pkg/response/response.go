// Package response is the single JSON envelope for the HTTP API:
// success is {"data": ...}, failure is {"error": "..."}. One shape keeps
// clients trivial.
package response

import "github.com/gin-gonic/gin"

// OK writes a success envelope with the given status code.
func OK(c *gin.Context, code int, data any) {
	c.JSON(code, gin.H{"data": data})
}

// Error writes a failure envelope with the given status code.
func Error(c *gin.Context, code int, message string) {
	c.JSON(code, gin.H{"error": message})
}
