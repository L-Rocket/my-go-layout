// Package validator bridges gin binding errors to the response envelope,
// so handlers never write ad-hoc 400 bodies.
package validator

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/L-Rocket/my-project/pkg/response"
)

// BindJSON binds the request body into v. On failure it writes a 400
// envelope and returns false — handlers should just `return`.
func BindJSON(c *gin.Context, v any) bool {
	if err := c.ShouldBindJSON(v); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request body: "+err.Error())
		return false
	}
	return true
}
