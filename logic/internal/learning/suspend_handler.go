package learning

import (
	"net/http"

	"github.com/english/backend/internal/domain"
	"github.com/english/backend/internal/middleware"
	"github.com/gin-gonic/gin"
)

// PUT /api/v1/learning/progress/suspend/:word_id
func (h *Handler) SuspendWord(c *gin.Context) {
	userID := c.GetString(middleware.UserIDKey)
	wordID := c.Param("word_id")

	if wordID == "" {
		domain.Fail(c, http.StatusBadRequest, "VALIDATION_ERROR", "word_id is required")
		return
	}

	if err := h.service.SuspendWord(c.Request.Context(), userID, wordID); err != nil {
		domain.Fail(c, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}

	domain.OK(c, gin.H{"message": "Word suspended successfully"})
}
