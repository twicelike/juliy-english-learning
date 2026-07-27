package auth

import (
	"errors"
	"net/http"

	"github.com/english/backend/internal/domain"
	"github.com/gin-gonic/gin"
)

const refreshTokenCookieName = "refresh_token"

type Handler struct {
	service     Service
	isProd      bool
	refreshDays int
}

func NewHandler(service Service, env string, refreshDays int) *Handler {
	return &Handler{
		service:     service,
		isProd:      env == "production",
		refreshDays: refreshDays,
	}
}

func (h *Handler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		domain.Fail(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}

	resp, rt, err := h.service.Register(c.Request.Context(), &req)
	if err != nil {
		if errors.Is(err, ErrUserExists) {
			domain.Fail(c, http.StatusConflict, "CONFLICT", err.Error())
			return
		}
		domain.Fail(c, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}

	h.setRefreshTokenCookie(c, rt)
	domain.OK(c, resp)
}

func (h *Handler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		domain.Fail(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}

	resp, rt, err := h.service.Login(c.Request.Context(), &req)
	if err != nil {
		if errors.Is(err, ErrInvalidCreds) {
			domain.Fail(c, http.StatusUnauthorized, "UNAUTHORIZED", err.Error())
			return
		}
		domain.Fail(c, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}

	h.setRefreshTokenCookie(c, rt)
	domain.OK(c, resp)
}

func (h *Handler) Refresh(c *gin.Context) {
	rt, err := c.Cookie(refreshTokenCookieName)
	if err != nil || rt == "" {
		domain.Fail(c, http.StatusUnauthorized, "UNAUTHORIZED", "refresh token missing")
		return
	}

	resp, newRt, err := h.service.Refresh(c.Request.Context(), rt)
	if err != nil {
		h.clearRefreshTokenCookie(c)
		domain.Fail(c, http.StatusUnauthorized, "UNAUTHORIZED", ErrInvalidToken.Error())
		return
	}

	h.setRefreshTokenCookie(c, newRt)
	domain.OK(c, resp)
}

func (h *Handler) Logout(c *gin.Context) {
	rt, _ := c.Cookie(refreshTokenCookieName)
	
	_ = h.service.Logout(c.Request.Context(), rt)
	h.clearRefreshTokenCookie(c)

	c.JSON(http.StatusNoContent, nil) // 204 No Content
}

// -----------------------------------------------------------------------------
// Cookie Helpers
// -----------------------------------------------------------------------------

func (h *Handler) setRefreshTokenCookie(c *gin.Context, token string) {
	maxAge := h.refreshDays * 24 * 60 * 60
	c.SetCookie(
		refreshTokenCookieName,
		token,
		maxAge,
		"/api/v1/auth", // Restrict cookie to auth routes
		"",             // Domain
		h.isProd,       // Secure only in prod
		true,           // HttpOnly
	)
	c.SetSameSite(http.SameSiteLaxMode)
}

func (h *Handler) clearRefreshTokenCookie(c *gin.Context) {
	c.SetCookie(
		refreshTokenCookieName,
		"",
		-1,
		"/api/v1/auth",
		"",
		h.isProd,
		true,
	)
}
