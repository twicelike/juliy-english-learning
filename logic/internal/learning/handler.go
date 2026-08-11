package learning

import (

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

// POST /api/v1/learning/session
// Body: { "mode": "new_words" }
func (h *Handler) CreateSession(c *gin.Context) {
	userID := c.GetString(middleware.UserIDKey)

	var req struct {
		Mode string `json:"mode" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		domain.Fail(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}

	if req.Mode != "new_words" && req.Mode != "review" {
		domain.Fail(c, http.StatusBadRequest, "VALIDATION_ERROR", "only mode 'new_words' or 'review' is currently supported for stateful sessions")
		return
	}

	var session *SessionDTO
	var err error

	if req.Mode == "new_words" {
		session, err = h.service.CreateNewWordsSession(c.Request.Context(), userID)
	} else if req.Mode == "review" {
		session, err = h.service.CreateReviewSession(c.Request.Context(), userID)
	}

	if err != nil {
		if err.Error() == "no_new_words" {
			domain.Fail(c, http.StatusNotFound, "NO_NEW_WORDS", "You have no new bookmarked words to learn")
			return
		}
		if err.Error() == "no_due_words" {
			domain.Fail(c, http.StatusNotFound, "NO_DUE_WORDS", "No words are due for review.")
			return
		}
		domain.Fail(c, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}

	domain.OK(c, session)
}

// GET /api/v1/learning/session/:id
func (h *Handler) GetSessionByID(c *gin.Context) {
	sessionID := c.Param("id")
	if sessionID == "" {
		domain.Fail(c, http.StatusBadRequest, "VALIDATION_ERROR", "session id is required")
		return
	}

	session, err := h.service.GetSessionByID(c.Request.Context(), sessionID)
	if err != nil {
		if err.Error() == "session not found" {
			domain.Fail(c, http.StatusNotFound, "NOT_FOUND", "session not found")
			return
		}
		domain.Fail(c, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}

	domain.OK(c, session)
}

// POST /api/v1/learning/sessions/:id/submit
func (h *Handler) SubmitSessionResults(c *gin.Context) {
	userID := c.GetString(middleware.UserIDKey)
	sessionID := c.Param("id")

	var req SubmitSessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		domain.Fail(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}

	err := h.service.SubmitSessionResults(c.Request.Context(), userID, sessionID, &req)
	if err != nil {
		if err.Error() == "session not found" {
			domain.Fail(c, http.StatusNotFound, "NOT_FOUND", err.Error())
			return
		}
		if err.Error() == "session is not active" || err.Error() == "unauthorized session access" {
			domain.Fail(c, http.StatusForbidden, "FORBIDDEN", err.Error())
			return
		}
		domain.Fail(c, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}

	c.JSON(http.StatusNoContent, nil) // 204
}

// POST /api/v1/learning/debug/advance-time
func (h *Handler) DebugAdvanceTime(c *gin.Context) {
	userID := c.GetString(middleware.UserIDKey)

	var req struct {
		Days int `json:"days"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		req.Days = 1 // default 1 day
	}
	if req.Days <= 0 {
		req.Days = 1
	}

	if err := h.service.DebugAdvanceTime(c.Request.Context(), userID, req.Days); err != nil {
		domain.Fail(c, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}

	domain.OK(c, map[string]interface{}{"advanced_days": req.Days})
}
