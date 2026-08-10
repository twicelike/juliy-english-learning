package user

import (
	"errors"
	"net/http"

	"github.com/english/backend/internal/domain"
	"github.com/english/backend/internal/middleware"
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

func (h *Handler) GetProfile(c *gin.Context) {
	userID := c.GetString(middleware.UserIDKey)
	
	profile, err := h.service.GetProfile(c.Request.Context(), userID)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			domain.Fail(c, http.StatusNotFound, "NOT_FOUND", "User not found")
			return
		}
		domain.Fail(c, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}
	
	domain.OK(c, profile)
}

func (h *Handler) AddBookmark(c *gin.Context) {
	userID := c.GetString(middleware.UserIDKey)
	var req BookmarkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		domain.Fail(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}

	if err := h.service.AddBookmark(c.Request.Context(), userID, req.Word, req.Type); err != nil {
		domain.Fail(c, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": "Bookmark added",
	})
}

func (h *Handler) RemoveBookmark(c *gin.Context) {
	userID := c.GetString(middleware.UserIDKey)
	word := c.Param("word")
	if word == "" {
		domain.Fail(c, http.StatusBadRequest, "VALIDATION_ERROR", "Word is required")
		return
	}

	if err := h.service.RemoveBookmark(c.Request.Context(), userID, word); err != nil {
		domain.Fail(c, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}

	c.JSON(http.StatusNoContent, nil) // 204
}

func (h *Handler) ListBookmarks(c *gin.Context) {
	userID := c.GetString(middleware.UserIDKey)
	
	bookmarks, err := h.service.ListBookmarks(c.Request.Context(), userID)
	if err != nil {
		domain.Fail(c, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
		return
	}
	
	domain.OK(c, bookmarks)
}
