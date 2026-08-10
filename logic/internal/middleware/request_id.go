package middleware

import (
	"crypto/rand"
	"encoding/hex"

	"github.com/gin-gonic/gin"
)

// RequestID generates a unique ID for each request and attaches it to the response headers and context.
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		reqID := c.GetHeader("X-Request-ID")
		if reqID == "" {
			b := make([]byte, 16)
			_, _ = rand.Read(b)
			reqID = hex.EncodeToString(b)
		}

		c.Set("X-Request-ID", reqID)
		c.Writer.Header().Set("X-Request-ID", reqID)

		c.Next()
	}
}
