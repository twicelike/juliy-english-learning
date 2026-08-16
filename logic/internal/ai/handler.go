package ai

import (
	"errors"
	"net/http"

	"github.com/english/backend/internal/dictionary"
	"github.com/english/backend/internal/domain"
	"github.com/english/backend/internal/middleware"
	"github.com/gin-gonic/gin"
)

// Handler exposes AI Tutor features over HTTP.
// It follows the same layer conventions as other handlers: parse → delegate → respond.
type Handler struct {
	service     AITutorService
	rateLimiter RateLimiter
}

// NewHandler creates a new AI Handler.
func NewHandler(service AITutorService, rateLimiter RateLimiter) *Handler {
	return &Handler{service: service, rateLimiter: rateLimiter}
}

// -----------------------------------------------------------------------
// Request / Response types
// -----------------------------------------------------------------------

type explainSenseRequest struct {
	WordID         string `json:"wordId" binding:"required"`
	SenseIndex     int    `json:"senseIndex"`
	UserLevel      string `json:"userLevel"`
	NativeLanguage string `json:"nativeLanguage"`
}

type wordRequest struct {
	Word           string `json:"word" binding:"required"`
	POS            string `json:"pos"`
	Context        string `json:"context,omitempty"`
	UserLevel      string `json:"userLevel"`
	NativeLanguage string `json:"nativeLanguage"`
}



type compareRequest struct {
	Word1          string `json:"word1" binding:"required"`
	Word2          string `json:"word2" binding:"required"`
	POS            string `json:"pos"`
	UserLevel      string `json:"userLevel"`
	NativeLanguage string `json:"nativeLanguage"`
}

type textRequest struct {
	Text           string `json:"text" binding:"required,min=3"`
	UserLevel      string `json:"userLevel"`
	NativeLanguage string `json:"nativeLanguage"`
}

type fillBlankRequest struct {
	Words []struct {
		Word string `json:"word"`
		POS  string `json:"pos"`
	} `json:"words" binding:"required,min=1"`
	UserLevel      string `json:"userLevel"`
	NativeLanguage string `json:"nativeLanguage"`
}

// -----------------------------------------------------------------------
// Word handlers
// -----------------------------------------------------------------------

// POST /api/v1/ai/words/explain-sense
func (h *Handler) ExplainSense(c *gin.Context) {
	userID := c.GetString(middleware.UserIDKey)
	if err := h.rateLimiter.Allow(c.Request.Context(), userID); err != nil {
		domain.Fail(c, http.StatusTooManyRequests, "RATE_LIMITED", err.Error())
		return
	}

	var req explainSenseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		domain.Fail(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}

	resp, err := h.service.ExplainSense(c.Request.Context(), req.WordID, req.SenseIndex, req.UserLevel, req.NativeLanguage)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}
	domain.OK(c, resp)
}

// POST /api/v1/ai/words/nuance
func (h *Handler) ExplainNuance(c *gin.Context) {
	userID := c.GetString(middleware.UserIDKey)
	if err := h.rateLimiter.Allow(c.Request.Context(), userID); err != nil {
		domain.Fail(c, http.StatusTooManyRequests, "RATE_LIMITED", err.Error())
		return
	}

	var req wordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		domain.Fail(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}

	resp, err := h.service.ExplainNuance(c.Request.Context(), PromptContext{
		Word: req.Word, POS: req.POS,
		UserLevel: req.UserLevel, NativeLanguage: req.NativeLanguage,
	})
	if err != nil {
		h.handleServiceError(c, err)
		return
	}
	domain.OK(c, resp)
}

// POST /api/v1/ai/words/compare
func (h *Handler) CompareWords(c *gin.Context) {
	userID := c.GetString(middleware.UserIDKey)
	if err := h.rateLimiter.Allow(c.Request.Context(), userID); err != nil {
		domain.Fail(c, http.StatusTooManyRequests, "RATE_LIMITED", err.Error())
		return
	}

	var req compareRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		domain.Fail(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}

	resp, err := h.service.CompareWords(c.Request.Context(), PromptContext{
		Word: req.Word1, POS: req.POS,
		UserLevel: req.UserLevel, NativeLanguage: req.NativeLanguage,
	}, req.Word2)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}
	domain.OK(c, resp)
}

// POST /api/v1/ai/words/examples
func (h *Handler) GenerateExamples(c *gin.Context) {
	userID := c.GetString(middleware.UserIDKey)
	if err := h.rateLimiter.Allow(c.Request.Context(), userID); err != nil {
		domain.Fail(c, http.StatusTooManyRequests, "RATE_LIMITED", err.Error())
		return
	}

	var req wordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		domain.Fail(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}

	resp, err := h.service.GenerateExamples(c.Request.Context(), PromptContext{
		Word: req.Word, POS: req.POS,
		UserLevel: req.UserLevel, NativeLanguage: req.NativeLanguage,
	})
	if err != nil {
		h.handleServiceError(c, err)
		return
	}
	domain.OK(c, resp)
}


// -----------------------------------------------------------------------
// Writing handlers
// -----------------------------------------------------------------------

// POST /api/v1/ai/writing/check
func (h *Handler) CheckGrammar(c *gin.Context) {
	userID := c.GetString(middleware.UserIDKey)
	if err := h.rateLimiter.Allow(c.Request.Context(), userID); err != nil {
		domain.Fail(c, http.StatusTooManyRequests, "RATE_LIMITED", err.Error())
		return
	}

	var req textRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		domain.Fail(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}

	resp, err := h.service.CheckGrammar(c.Request.Context(), PromptContext{
		UserLevel: req.UserLevel, NativeLanguage: req.NativeLanguage,
	}, req.Text)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}
	domain.OK(c, resp)
}

// POST /api/v1/ai/writing/vocab
func (h *Handler) SuggestWritingVocab(c *gin.Context) {
	userID := c.GetString(middleware.UserIDKey)
	if err := h.rateLimiter.Allow(c.Request.Context(), userID); err != nil {
		domain.Fail(c, http.StatusTooManyRequests, "RATE_LIMITED", err.Error())
		return
	}

	var req textRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		domain.Fail(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}

	resp, err := h.service.SuggestWritingVocab(c.Request.Context(), PromptContext{
		UserLevel: req.UserLevel, NativeLanguage: req.NativeLanguage,
	}, req.Text)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}
	domain.OK(c, resp)
}

// -----------------------------------------------------------------------
// Quiz handlers
// -----------------------------------------------------------------------

// POST /api/v1/ai/quiz/fill-blank
func (h *Handler) GenerateFillBlank(c *gin.Context) {
	userID := c.GetString(middleware.UserIDKey)
	if err := h.rateLimiter.Allow(c.Request.Context(), userID); err != nil {
		domain.Fail(c, http.StatusTooManyRequests, "RATE_LIMITED", err.Error())
		return
	}

	var req fillBlankRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		domain.Fail(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}

	words := make([]dictionary.WordDocument, 0, len(req.Words))
	for _, w := range req.Words {
		words = append(words, dictionary.WordDocument{Word: w.Word, POS: w.POS})
	}

	resp, err := h.service.GenerateFillBlank(c.Request.Context(), PromptContext{
		UserLevel: req.UserLevel, NativeLanguage: req.NativeLanguage,
	}, words)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}
	domain.OK(c, resp)
}

// -----------------------------------------------------------------------
// Internal helpers
// -----------------------------------------------------------------------

func (h *Handler) handleServiceError(c *gin.Context, err error) {
	if errors.Is(err, ErrTooManyRequests) {
		domain.Fail(c, http.StatusTooManyRequests, "RATE_LIMITED", err.Error())
		return
	}
	if errors.Is(err, ErrInvalidAIResponse) {
		domain.Fail(c, http.StatusUnprocessableEntity, "AI_RESPONSE_INVALID", err.Error())
		return
	}
	if errors.Is(err, ErrAIUnavailable) {
		domain.Fail(c, http.StatusServiceUnavailable, "AI_UNAVAILABLE", err.Error())
		return
	}
	domain.Fail(c, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
}
