package middleware

import (
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

// CORS configures Cross-Origin Resource Sharing headers.
// Allowed origins are read from the CORS_ORIGINS environment variable
// (comma-separated). Defaults to localhost:3000 for local development.
func CORS() gin.HandlerFunc {
	rawOrigins := os.Getenv("CORS_ORIGINS")
	if rawOrigins == "" {
		rawOrigins = "http://localhost:3000"
	}
	allowed := strings.Split(rawOrigins, ",")

	allowedSet := make(map[string]struct{}, len(allowed))
	for _, o := range allowed {
		allowedSet[strings.TrimSpace(o)] = struct{}{}
	}

	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		if _, ok := allowedSet[origin]; ok {
			c.Header("Access-Control-Allow-Origin", origin)
		}
		c.Header("Access-Control-Allow-Methods", "GET,POST,PUT,PATCH,DELETE,OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type,Authorization")
		c.Header("Access-Control-Allow-Credentials", "true")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	}
}
