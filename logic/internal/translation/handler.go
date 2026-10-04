package translation

import (
	"net/http"

	"github.com/english/backend/internal/domain"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	repo Repository
}

func NewHandler(repo Repository) *Handler {
	return &Handler{
		repo: repo,
	}
}

func (h *Handler) GetJobStats(c *gin.Context) {
	stats, err := h.repo.GetJobStats(c.Request.Context())
	if err != nil {
		domain.Fail(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to get job stats")
		return
	}

	domain.OK(c, stats)
}

func (h *Handler) RetryFailedJobs(c *gin.Context) {
	count, err := h.repo.RetryFailedJobs(c.Request.Context())
	if err != nil {
		domain.Fail(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to retry jobs")
		return
	}

	domain.OK(c, gin.H{
		"message": "Retried failed jobs successfully",
		"count":   count,
	})
}
