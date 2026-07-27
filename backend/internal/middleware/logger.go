package middleware

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
)

// Logger logs method, path, status, and latency for every request.
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		slog.Info("request",
			"method",  c.Request.Method,
			"path",    c.Request.URL.Path,
			"status",  c.Writer.Status(),
			"latency", time.Since(start).String(),
			"ip",      c.ClientIP(),
		)
	}
}
