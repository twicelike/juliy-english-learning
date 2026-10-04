package middleware

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"time"

	"github.com/english/backend/internal/auth"
	"github.com/english/backend/internal/domain"
	"github.com/gin-gonic/gin"
)

const UserIDKey = "user_id"
const sessionCookieName = "session_id"

// RequireAuth checks the session_id cookie and validates it against the database.
func RequireAuth(repo auth.Repository) gin.HandlerFunc {
	return func(c *gin.Context) {
		rawToken, err := c.Cookie(sessionCookieName)
		if err != nil || rawToken == "" {
			domain.Fail(c, http.StatusUnauthorized, "UNAUTHORIZED", "Missing session cookie")
			c.Abort()
			return
		}

		// Hash the token
		hash := sha256.Sum256([]byte(rawToken))
		tokenHash := hex.EncodeToString(hash[:])

		// Find in DB
		session, err := repo.GetSessionByHash(c.Request.Context(), tokenHash)
		if err != nil {
			domain.Fail(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to validate session")
			c.Abort()
			return
		}
		if session == nil {
			domain.Fail(c, http.StatusUnauthorized, "UNAUTHORIZED", "Invalid session")
			c.Abort()
			return
		}

		// Check revoked
		if session.RevokedAt != nil {
			domain.Fail(c, http.StatusUnauthorized, "UNAUTHORIZED", "Session revoked")
			c.Abort()
			return
		}

		// Check expired
		if session.ExpiresAt.Before(time.Now()) {
			domain.Fail(c, http.StatusUnauthorized, "UNAUTHORIZED", "Session expired")
			c.Abort()
			return
		}

		user, err := repo.GetUserByID(c.Request.Context(), session.UserID)
		if err != nil || user == nil {
			domain.Fail(c, http.StatusUnauthorized, "UNAUTHORIZED", "User not found")
			c.Abort()
			return
		}
		
		if user.Status == "banned" {
			domain.Fail(c, http.StatusForbidden, "FORBIDDEN", "Your account is banned")
			c.Abort()
			return
		}

		// (Optional) We could update last_used_at here, but doing it on every request
		// might add unnecessary load to DB. A background job or sampling could be better.

		// Inject user ID into context
		c.Set(UserIDKey, session.UserID.Hex())
		c.Next()
	}
}

// RequireAdmin checks session and verifies if the user has the 'admin' role.
func RequireAdmin(repo auth.Repository) gin.HandlerFunc {
	return func(c *gin.Context) {
		rawToken, err := c.Cookie(sessionCookieName)
		if err != nil || rawToken == "" {
			domain.Fail(c, http.StatusUnauthorized, "UNAUTHORIZED", "Missing session cookie")
			c.Abort()
			return
		}

		hash := sha256.Sum256([]byte(rawToken))
		tokenHash := hex.EncodeToString(hash[:])

		session, err := repo.GetSessionByHash(c.Request.Context(), tokenHash)
		if err != nil || session == nil || session.RevokedAt != nil || session.ExpiresAt.Before(time.Now()) {
			domain.Fail(c, http.StatusUnauthorized, "UNAUTHORIZED", "Invalid session")
			c.Abort()
			return
		}

		user, err := repo.GetUserByID(c.Request.Context(), session.UserID)
		if err != nil || user == nil {
			domain.Fail(c, http.StatusUnauthorized, "UNAUTHORIZED", "User not found")
			c.Abort()
			return
		}
		
		if user.Status == "banned" {
			domain.Fail(c, http.StatusForbidden, "FORBIDDEN", "Your account is banned")
			c.Abort()
			return
		}

		if user.Role != auth.RoleAdmin && user.Role != auth.RoleSuperior {
			domain.Fail(c, http.StatusForbidden, "FORBIDDEN", "Require admin privileges")
			c.Abort()
			return
		}

		c.Set(UserIDKey, session.UserID.Hex())
		c.Next()
	}
}
