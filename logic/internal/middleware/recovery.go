package middleware

import (
	"fmt"
	"net/http"
	"runtime/debug"

	"github.com/english/backend/internal/domain"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

// Recovery catches panics, logs them with full stack trace, and returns a standard 500 response.
func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				reqID := c.GetHeader("X-Request-ID")
				stack := string(debug.Stack())

				log.Error().
					Str("req_id", reqID).
					Str("panic", fmt.Sprintf("%v", r)).
					Str("stack", stack).
					Msg("panic recovered")

				domain.Fail(c, http.StatusInternalServerError,
					"INTERNAL_SERVER_ERROR", "An unexpected error occurred.")
				c.Abort()
			}
		}()
		c.Next()
	}
}
