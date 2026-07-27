package middleware

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/english/backend/internal/domain"
)

// Recovery catches panics, logs them, and returns a standard 500 response
// so the server never crashes on an unhandled error.
func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				slog.Error("panic recovered", "error", r)
				domain.Fail(c, http.StatusInternalServerError,
					"INTERNAL_SERVER_ERROR", "An unexpected error occurred.")
				c.Abort()
			}
		}()
		c.Next()
	}
}
