package writing

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service WritingService
}

func NewHandler(service WritingService) *Handler {
	return &Handler{
		service: service,
	}
}

// GetNewMission handles GET /api/v1/writing/missions/new
func (h *Handler) GetNewMission(c *gin.Context) {
	userIDStr := c.GetString("user_id")
	if userIDStr == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	topic := c.Query("topic")
	difficulty := c.Query("difficulty")

	mission, err := h.service.GetNewMission(c.Request.Context(), userIDStr, topic, difficulty)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": mission})
}

// GetReviewMission handles GET /api/v1/writing/missions/review
func (h *Handler) GetReviewMission(c *gin.Context) {
	userIDStr := c.GetString("user_id")
	if userIDStr == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	topic := c.Query("topic")
	difficulty := c.Query("difficulty")

	mission, err := h.service.GetReviewMission(c.Request.Context(), userIDStr, topic, difficulty)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": mission})
}

// SubmitMission handles POST /api/v1/writing/missions/:id/submit
func (h *Handler) SubmitMission(c *gin.Context) {
	userIDStr := c.GetString("user_id")
	if userIDStr == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	missionIDStr := c.Param("id")
	if missionIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "mission id is required"})
		return
	}

	var req SubmissionRequestDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	resp, err := h.service.SubmitMission(c.Request.Context(), userIDStr, missionIDStr, req.Content)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": resp})
}

// GetDashboardStats handles GET /api/v1/writing/dashboard
func (h *Handler) GetDashboardStats(c *gin.Context) {
	userIDStr := c.GetString("user_id")
	if userIDStr == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	stats, err := h.service.GetDashboardStats(c.Request.Context(), userIDStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": stats})
}
