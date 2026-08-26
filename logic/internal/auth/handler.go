package auth

import (
	"errors"
	"net/http"

	"github.com/english/backend/internal/domain"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/bson"
)

const sessionCookieName = "session_id"

type Handler struct {
	service        Service
	oauthService   GoogleOAuthService
	isProd         bool
	sessionExpDays int
	frontendURL    string
}

func NewHandler(service Service, oauthService GoogleOAuthService, env string, sessionExpDays int, frontendURL string) *Handler {
	return &Handler{
		service:        service,
		oauthService:   oauthService,
		isProd:         env == "production",
		sessionExpDays: sessionExpDays,
		frontendURL:    frontendURL,
	}
}

func (h *Handler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		domain.Fail(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}

	userAgent := c.Request.UserAgent()
	ip := c.ClientIP()

	resp, rawToken, err := h.service.Register(c.Request.Context(), &req, userAgent, ip)
	if err != nil {
		if errors.Is(err, ErrUserExists) {
			domain.Fail(c, http.StatusConflict, "CONFLICT", err.Error())
			return
		}
		domain.Fail(c, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}

	h.setSessionCookie(c, rawToken)
	domain.OK(c, resp)
}

func (h *Handler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		domain.Fail(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}

	userAgent := c.Request.UserAgent()
	ip := c.ClientIP()

	resp, rawToken, err := h.service.Login(c.Request.Context(), &req, userAgent, ip)
	if err != nil {
		if errors.Is(err, ErrInvalidCreds) {
			domain.Fail(c, http.StatusUnauthorized, "UNAUTHORIZED", err.Error())
			return
		}
		domain.Fail(c, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}

	h.setSessionCookie(c, rawToken)
	domain.OK(c, resp)
}

func (h *Handler) Logout(c *gin.Context) {
	rawToken, _ := c.Cookie(sessionCookieName)

	_ = h.service.Logout(c.Request.Context(), rawToken)
	h.clearSessionCookie(c)

	c.JSON(http.StatusNoContent, nil) // 204 No Content
}

func (h *Handler) LogoutAll(c *gin.Context) {
	userIDStr := c.GetString("user_id")
	userID, err := bson.ObjectIDFromHex(userIDStr)
	if err != nil {
		domain.Fail(c, http.StatusUnauthorized, "UNAUTHORIZED", "invalid user context")
		return
	}

	_ = h.service.LogoutAll(c.Request.Context(), userID)
	h.clearSessionCookie(c)

	c.JSON(http.StatusNoContent, nil) // 204 No Content
}

func (h *Handler) GetMe(c *gin.Context) {
	userIDStr := c.GetString("user_id")
	userID, err := bson.ObjectIDFromHex(userIDStr)
	if err != nil {
		domain.Fail(c, http.StatusUnauthorized, "UNAUTHORIZED", "invalid user context")
		return
	}

	resp, err := h.service.GetMe(c.Request.Context(), userID)
	if err != nil {
		domain.Fail(c, http.StatusUnauthorized, "UNAUTHORIZED", err.Error())
		return
	}

	domain.OK(c, resp)
}

func (h *Handler) GetSessions(c *gin.Context) {
	userIDStr := c.GetString("user_id")
	userID, err := bson.ObjectIDFromHex(userIDStr)
	if err != nil {
		domain.Fail(c, http.StatusUnauthorized, "UNAUTHORIZED", "invalid user context")
		return
	}

	rawToken, _ := c.Cookie(sessionCookieName)

	resp, err := h.service.GetActiveSessions(c.Request.Context(), userID, rawToken)
	if err != nil {
		domain.Fail(c, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}

	domain.OK(c, resp)
}

func (h *Handler) RevokeSession(c *gin.Context) {
	userIDStr := c.GetString("user_id")
	userID, err := bson.ObjectIDFromHex(userIDStr)
	if err != nil {
		domain.Fail(c, http.StatusUnauthorized, "UNAUTHORIZED", "invalid user context")
		return
	}

	sessionIDStr := c.Param("id")
	sessionID, err := bson.ObjectIDFromHex(sessionIDStr)
	if err != nil {
		domain.Fail(c, http.StatusBadRequest, "VALIDATION_ERROR", "invalid session id")
		return
	}

	err = h.service.RevokeSession(c.Request.Context(), sessionID, userID)
	if err != nil {
		domain.Fail(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to revoke session")
		return
	}

	domain.OK(c, gin.H{"message": "session revoked"})
}

func (h *Handler) GoogleLogin(c *gin.Context) {
	state, err := generateRandomToken()
	if err != nil {
		domain.Fail(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to generate state")
		return
	}

	c.SetCookie(
		"oauth_state",
		state,
		5*60, // 5 minutes
		"/",
		"",
		h.isProd,
		true,
	)
	c.SetSameSite(http.SameSiteLaxMode)

	url := h.oauthService.GetAuthorizationURL(state)
	c.Redirect(http.StatusTemporaryRedirect, url)
}

func (h *Handler) GoogleCallback(c *gin.Context) {
	state := c.Query("state")
	code := c.Query("code")

	cookieState, err := c.Cookie("oauth_state")
	if err != nil || state != cookieState {
		c.Redirect(http.StatusTemporaryRedirect, h.frontendURL+"/login?error=invalid_oauth_state")
		return
	}

	c.SetCookie("oauth_state", "", -1, "/", "", h.isProd, true)

	userInfo, err := h.oauthService.ExchangeCode(c.Request.Context(), code)
	if err != nil {
		c.Redirect(http.StatusTemporaryRedirect, h.frontendURL+"/login?error=oauth_exchange_failed")
		return
	}

	userAgent := c.Request.UserAgent()
	ip := c.ClientIP()

	rawSessionToken, _ := c.Cookie(sessionCookieName)

	// If a session exists, we attempt to link
	if rawSessionToken != "" {
		userID, err := h.service.GetUserIDFromToken(c.Request.Context(), rawSessionToken)
		if err == nil {
			err = h.service.LinkGoogleIdentity(c.Request.Context(), userID, userInfo)
			if err != nil {
				if errors.Is(err, ErrIdentityAlreadyLinked) {
					c.Redirect(http.StatusTemporaryRedirect, h.frontendURL+"/profile?error=identity_already_linked")
					return
				}
				c.Redirect(http.StatusTemporaryRedirect, h.frontendURL+"/profile?error=linking_failed")
				return
			}
			c.Redirect(http.StatusTemporaryRedirect, h.frontendURL+"/profile?success=google_linked")
			return
		}
	}

	// Login Flow
	_, rawToken, err := h.service.LoginWithGoogle(c.Request.Context(), userInfo, userAgent, ip)
	if err != nil {
		if errors.Is(err, ErrEmailAlreadyExists) {
			c.Redirect(http.StatusTemporaryRedirect, h.frontendURL+"/login?error=google_email_exists")
			return
		}
		if errors.Is(err, ErrEmailNotVerified) {
			c.Redirect(http.StatusTemporaryRedirect, h.frontendURL+"/login?error=google_email_not_verified")
			return
		}
		c.Redirect(http.StatusTemporaryRedirect, h.frontendURL+"/login?error=google_login_failed")
		return
	}

	h.setSessionCookie(c, rawToken)
	c.Redirect(http.StatusTemporaryRedirect, h.frontendURL+"/profile")
}

// -----------------------------------------------------------------------------
// Verification & Recovery
// -----------------------------------------------------------------------------

func (h *Handler) VerifyEmail(c *gin.Context) {
	var req VerifyEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		domain.Fail(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}

	if err := h.service.VerifyEmail(c.Request.Context(), req.Token); err != nil {
		domain.Fail(c, http.StatusBadRequest, "INVALID_TOKEN", err.Error())
		return
	}

	domain.OK(c, gin.H{"message": "Email verified successfully"})
}

func (h *Handler) ResendVerification(c *gin.Context) {
	var req ResendVerificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		domain.Fail(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}

	// In a real production app, apply rate limiting here.
	if err := h.service.ResendVerification(c.Request.Context(), req.Email); err != nil {
		domain.Fail(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to resend verification email")
		return
	}

	domain.OK(c, gin.H{"message": "Verification email sent if account exists and is unverified"})
}

func (h *Handler) ForgotPassword(c *gin.Context) {
	var req ForgotPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		domain.Fail(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}

	// In a real production app, apply rate limiting here.
	if err := h.service.ForgotPassword(c.Request.Context(), req.Email); err != nil {
		// Log error internally, but always return OK to prevent enumeration
		domain.OK(c, gin.H{"message": "If the email exists, a password reset link has been sent."})
		return
	}

	domain.OK(c, gin.H{"message": "If the email exists, a password reset link has been sent."})
}

func (h *Handler) ResetPassword(c *gin.Context) {
	var req ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		domain.Fail(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}

	if err := h.service.ResetPassword(c.Request.Context(), req.Token, req.NewPassword); err != nil {
		domain.Fail(c, http.StatusBadRequest, "RESET_FAILED", err.Error())
		return
	}

	// We revoke sessions, so clear the client's current session cookie too
	h.clearSessionCookie(c)

	domain.OK(c, gin.H{"message": "Password reset successfully. Please log in again."})
}

// -----------------------------------------------------------------------------
// Cookie Helpers
// -----------------------------------------------------------------------------

func (h *Handler) setSessionCookie(c *gin.Context, token string) {
	maxAge := h.sessionExpDays * 24 * 60 * 60
	c.SetCookie(
		sessionCookieName,
		token,
		maxAge,
		"/",      // Available everywhere
		"",       // Domain
		h.isProd, // Secure only in prod
		true,     // HttpOnly
	)
	c.SetSameSite(http.SameSiteLaxMode)
}

func (h *Handler) clearSessionCookie(c *gin.Context) {
	c.SetCookie(
		sessionCookieName,
		"",
		-1,
		"/",
		"",
		h.isProd,
		true,
	)
}
