package profile

import (
	"net/http"

	"github.com/english/backend/internal/events"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	svc       Service
	publisher events.Publisher
}

func NewHandler(svc Service, publisher events.Publisher) *Handler {
	return &Handler{svc: svc, publisher: publisher}
}

// GetDashboard returns the user's dashboard info including profile and AI recommendation.
func (h *Handler) GetDashboard(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	resp, err := h.svc.GetDashboard(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    resp,
	})
}

// GetPlacementTest returns the 15 hardcoded questions.
func (h *Handler) GetPlacementTest(c *gin.Context) {
	resp, err := h.svc.GetPlacementTest(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    resp,
	})
}

// SubmitPlacementTest grades the test and initializes the user profile.
func (h *Handler) SubmitPlacementTest(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req SubmitTestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.svc.SubmitPlacementTest(c.Request.Context(), userID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Publish event so that skills get refreshed
	if h.publisher != nil {
		_ = h.publisher.Publish(c.Request.Context(), events.Event{
			Type:   events.EventTypeProfileRefreshRequested,
			UserID: userID,
			Reason: "assessment_test_completed",
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    resp,
	})
}

// RequestRefreshSkills handles manual request to refresh profile skills.
func (h *Handler) RequestRefreshSkills() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetString("user_id")
		if userID == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		if h.publisher != nil {
			err := h.publisher.Publish(c.Request.Context(), events.Event{
				Type:   events.EventTypeProfileRefreshRequested,
				UserID: userID,
				Reason: "manual_api_request",
			})
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to publish event"})
				return
			}
		}

		c.JSON(http.StatusAccepted, gin.H{
			"status": "refresh_requested",
		})
	}
}

// UpdatePreferences handles updating user goals and focus areas.
func (h *Handler) UpdatePreferences(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req UpdatePreferencesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.svc.UpdatePreferences(c.Request.Context(), userID, &req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Publish event so that Recommendation Engine updates based on new preferences
	if h.publisher != nil {
		_ = h.publisher.Publish(c.Request.Context(), events.Event{
			Type:   events.EventTypeProfileRefreshRequested,
			UserID: userID,
			Reason: "preferences_updated",
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
	})
}
