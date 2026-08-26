package middleware

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

// SecurityHeaders applies basic security headers to every response.
func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Prevent MIME-sniffing
		c.Writer.Header().Set("X-Content-Type-Options", "nosniff")
		// Prevent Clickjacking
		c.Writer.Header().Set("X-Frame-Options", "DENY")
		// Cross-Site Scripting (XSS) protection
		c.Writer.Header().Set("X-XSS-Protection", "1; mode=block")
		// Referrer policy
		c.Writer.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")

		c.Next()
	}
}

// CacheControl is a middleware generator for setting cache headers.
func CacheControl(maxAgeSeconds int) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Cache-Control", "public, max-age="+strconv.Itoa(maxAgeSeconds))
		c.Next()
	}
}
