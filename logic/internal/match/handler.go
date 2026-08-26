package match

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/english/backend/internal/domain"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{
		service: service,
	}
}

func (h *Handler) GetRecommendations(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		domain.Fail(c, http.StatusUnauthorized, "UNAUTHORIZED", "Unauthorized")
		return
	}

	limitStr := c.DefaultQuery("limit", "10")
	limit, _ := strconv.Atoi(limitStr)
	if limit <= 0 {
		limit = 10
	}

	cards, err := h.service.GetRecommendations(c.Request.Context(), userID.(string), limit)
	if err != nil {
		slog.Error("GetRecommendations failed", "user_id", userID, "error", err)
		domain.InternalError(c)
		return
	}

	domain.OK(c, cards)
}

func (h *Handler) Swipe(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		domain.Fail(c, http.StatusUnauthorized, "UNAUTHORIZED", "Unauthorized")
		return
	}

	var req SwipeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		domain.Fail(c, http.StatusBadRequest, "BAD_REQUEST", "Invalid request body")
		return
	}

	if req.WordID == "" || (req.Action != SwipeKeep && req.Action != SwipePass) {
		domain.Fail(c, http.StatusBadRequest, "BAD_REQUEST", "Invalid word_id or action")
		return
	}

	if err := h.service.Swipe(c.Request.Context(), userID.(string), &req); err != nil {
		domain.InternalError(c)
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
