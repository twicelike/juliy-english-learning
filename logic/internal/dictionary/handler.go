package dictionary

import (
	"net/http"
	"time"

	"github.com/english/backend/internal/domain"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	service Service
}

// NewHandler creates a new dictionary HTTP handler.
func NewHandler(service Service) *Handler {
	return &Handler{
		service: service,
	}
}

// Search handles GET /api/v1/words/search?q=...
func (h *Handler) Search(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		domain.OK(c, []WordDTO{})
		return
	}

	results, err := h.service.Search(c.Request.Context(), query)
	if err != nil {
		domain.Fail(c, http.StatusInternalServerError, "SEARCH_ERROR", err.Error())
		return
	}

	domain.OK(c, results)
}

// GetDetail handles GET /api/v1/words/detail/:word
func (h *Handler) GetDetail(c *gin.Context) {
	word := c.Param("word")
	if word == "" {
		domain.Fail(c, http.StatusBadRequest, "INVALID_REQUEST", "Word parameter is required")
		return
	}

	dto, err := h.service.GetDetail(c.Request.Context(), word)
	if err != nil {
		domain.Fail(c, http.StatusInternalServerError, "DB_ERROR", err.Error())
		return
	}
	if dto == nil {
		// As requested, handle not found gracefully. 404 is correct for REST.
		domain.Fail(c, http.StatusNotFound, "NOT_FOUND", "Word not found")
		return
	}

	domain.OK(c, dto)
}

// GetRandom handles GET /api/v1/words/random
func (h *Handler) GetRandom(c *gin.Context) {
	dto, err := h.service.GetRandom(c.Request.Context())
	if err != nil {
		domain.Fail(c, http.StatusInternalServerError, "DB_ERROR", err.Error())
		return
	}
	if dto == nil {
		domain.Fail(c, http.StatusNotFound, "NOT_FOUND", "No words found")
		return
	}

	domain.OK(c, dto)
}

// GetWordOfTheDay handles GET /api/v1/words/word-of-the-day
func (h *Handler) GetWordOfTheDay(c *gin.Context) {
	// Use today's date as a deterministic seed string.
	// We'll base it on UTC to avoid timezone edge cases.
	dateStr := time.Now().UTC().Format("2006-01-02")

	dto, err := h.service.GetWordOfTheDay(c.Request.Context(), dateStr)
	if err != nil {
		domain.Fail(c, http.StatusInternalServerError, "DB_ERROR", err.Error())
		return
	}
	if dto == nil {
		domain.Fail(c, http.StatusNotFound, "NOT_FOUND", "No words found")
		return
	}

	domain.OK(c, dto)
}
