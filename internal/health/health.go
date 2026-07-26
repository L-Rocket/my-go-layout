// Package health exposes a liveness payload. It never fails: a down
// dependency must not hide the process from the orchestrator.
package health

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Handler reports the process is up and which storage backend main wired
// ("memory" or "postgrest") — handy when eyeballing a deployment.
func Handler(store string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "up", "store": store})
	}
}
