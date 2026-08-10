package learning

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/english/backend/internal/domain"
	"github.com/english/backend/internal/middleware"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

// GET /api/v1/learning/dashboard
func (h *Handler) GetDashboard(c *gin.Context) {
	userID := c.GetString(middleware.UserIDKey)

	dto, err := h.service.GetDashboard(c.Request.Context(), userID)
	if err != nil {
		domain.Fail(c, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}

	domain.OK(c, dto)
}

// GET /api/v1/learning/session?mode=flashcard&limit=10
func (h *Handler) GetSession(c *gin.Context) {
	userID := c.GetString(middleware.UserIDKey)

	modeStr := c.DefaultQuery("mode", string(ModeFlashcard))
	mode := SessionMode(modeStr)
	if mode != ModeFlashcard && mode != ModeQuiz {
		domain.Fail(c, http.StatusBadRequest, "VALIDATION_ERROR", "mode must be 'flashcard' or 'quiz'")
		return
	}

	limit := 10
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	session, err := h.service.GetSession(c.Request.Context(), userID, mode, limit)
	if err != nil {
		domain.Fail(c, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}

	domain.OK(c, session)
}

// POST /api/v1/learning/progress
func (h *Handler) SubmitProgress(c *gin.Context) {
	userID := c.GetString(middleware.UserIDKey)

	var req SubmitProgressRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		domain.Fail(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}

	err := h.service.SubmitProgress(c.Request.Context(), userID, &req)
	if err != nil {
		if errors.Is(err, ErrWordNotFound) {
			domain.Fail(c, http.StatusNotFound, "NOT_FOUND", err.Error())
			return
		}
		domain.Fail(c, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}

	c.JSON(http.StatusNoContent, nil) // 204
}
