package speaking

import (
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service    Service
	recService RecommendationService
}

// NewHandler creates a new handler for the speaking domain.
func NewHandler(service Service, recService RecommendationService) *Handler {
	return &Handler{
		service:    service,
		recService: recService,
	}
}

// RegisterRoutes registers the speaking routes.
func (h *Handler) RegisterRoutes(router *gin.RouterGroup) {
	speaking := router.Group("/speaking")
	{
		speaking.GET("/tasks/random", h.GetRandomTask)
		speaking.GET("/tasks/recommended", h.GetRecommendedTask)
		speaking.POST("/sessions", h.StartSession)
		speaking.GET("/sessions", h.GetHistory)
		speaking.GET("/sessions/:id", h.GetSessionDetails)
		speaking.POST("/sessions/:id/attempts", h.SubmitAttempt)
		speaking.GET("/attempts/:id", h.GetAttempt)
		speaking.GET("/attempts/:id/audio", h.GetAttemptAudio)
	}
}

func (h *Handler) GetRecommendedTask(c *gin.Context) {
	userIDStr := c.GetString("user_id")
	if userIDStr == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	resp, err := h.recService.GetRecommendedTask(c.Request.Context(), userIDStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if resp == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "no recommendation available"})
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *Handler) GetHistory(c *gin.Context) {
	userIDStr := c.GetString("user_id")
	if userIDStr == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	items, err := h.service.GetHistory(c.Request.Context(), userIDStr, 20)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if items == nil {
		items = []SpeakingHistoryItem{}
	}

	c.JSON(http.StatusOK, gin.H{
		"items": items,
		"total": len(items),
	})
}

func (h *Handler) GetRandomTask(c *gin.Context) {
	topic := c.Query("topic")
	difficulty := c.Query("difficulty")
	taskType := c.Query("type")

	if taskType == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "type is required"})
		return
	}

	focus := c.Query("focus")

	task, err := h.service.GetRandomTask(c.Request.Context(), topic, difficulty, taskType, focus)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if task == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "no tasks found"})
		return
	}

	c.JSON(http.StatusOK, task)
}

func (h *Handler) StartSession(c *gin.Context) {
	userIDStr := c.GetString("user_id")
	if userIDStr == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req struct {
		TaskID string `json:"task_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	session, err := h.service.StartSession(c.Request.Context(), userIDStr, req.TaskID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, session)
}

func (h *Handler) GetSessionDetails(c *gin.Context) {
	userIDStr := c.GetString("user_id")
	if userIDStr == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	sessionIDStr := c.Param("id")
	if sessionIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "session id is required"})
		return
	}

	session, attempts, err := h.service.GetSessionDetails(c.Request.Context(), userIDStr, sessionIDStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if session == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "session not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"session":  session,
		"attempts": attempts,
	})
}

func (h *Handler) SubmitAttempt(c *gin.Context) {
	userIDStr := c.GetString("user_id")
	if userIDStr == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	sessionIDStr := c.Param("id")
	if sessionIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "session id is required"})
		return
	}

	// Parse multipart form
	// We expect "audio" file and "duration_ms" string
	durationMsStr := c.PostForm("duration_ms")
	var durationMs int
	fmt.Sscanf(durationMsStr, "%d", &durationMs)

	fileHeader, err := c.FormFile("audio")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "audio file is required"})
		return
	}

	file, err := fileHeader.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to open audio file"})
		return
	}
	defer file.Close()

	audioBytes, err := io.ReadAll(file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read audio file"})
		return
	}
	
	mimeType := fileHeader.Header.Get("Content-Type")

	attempt, err := h.service.SubmitAttempt(c.Request.Context(), userIDStr, sessionIDStr, mimeType, durationMs, audioBytes)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"attempt_id": attempt.ID,
		"status":     attempt.Status,
	})
}

func (h *Handler) GetAttempt(c *gin.Context) {
	userIDStr := c.GetString("user_id")
	if userIDStr == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	attemptIDStr := c.Param("id")
	if attemptIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "attempt id is required"})
		return
	}

	attempt, err := h.service.GetAttemptByID(c.Request.Context(), userIDStr, attemptIDStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, attempt)
}

func (h *Handler) GetAttemptAudio(c *gin.Context) {
	userIDStr := c.GetString("user_id")
	if userIDStr == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	attemptIDStr := c.Param("id")
	if attemptIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "attempt id is required"})
		return
	}

	reader, contentLength, mimeType, err := h.service.GetAttemptAudio(c.Request.Context(), userIDStr, attemptIDStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer reader.Close()

	c.Header("Accept-Ranges", "bytes")
	c.DataFromReader(http.StatusOK, contentLength, mimeType, reader, nil)
}
