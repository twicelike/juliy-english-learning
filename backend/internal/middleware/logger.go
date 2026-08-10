package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

// Logger logs method, path, status, and latency for every request using zerolog.
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path

		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()

		// Get request ID and user ID if available
		reqID := c.Writer.Header().Get("X-Request-ID")
		userID := c.GetString("user_id")

		logger := log.Info()
		if status >= 400 && status < 500 {
			logger = log.Warn()
		} else if status >= 500 {
			logger = log.Error()
		}

		logger.
			Str("method", c.Request.Method).
			Str("path", path).
			Int("status", status).
			Dur("latency", latency).
			Str("ip", c.ClientIP()).
			Str("req_id", reqID).
			Str("user_id", userID).
			Msg("HTTP request")
	}
}
