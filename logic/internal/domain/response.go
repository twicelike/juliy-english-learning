package domain

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// ApiResponse is the single standard envelope for every API response.
type ApiResponse struct {
	Success bool      `json:"success"`
	Data    any       `json:"data,omitempty"`
	Error   *ApiError `json:"error,omitempty"`
}

// ApiError carries a machine-readable code and a human-readable message.
type ApiError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// OK writes a 200 success response.
func OK(c *gin.Context, data any) {
	c.JSON(http.StatusOK, ApiResponse{Success: true, Data: data})
}

// Created writes a 201 success response.
func Created(c *gin.Context, data any) {
	c.JSON(http.StatusCreated, ApiResponse{Success: true, Data: data})
}

// Fail writes an error response with the given HTTP status code.
func Fail(c *gin.Context, status int, code, message string) {
	c.JSON(status, ApiResponse{
		Success: false,
		Error:   &ApiError{Code: code, Message: message},
	})
}

// InternalError writes a generic 500 response without leaking internal details.
func InternalError(c *gin.Context) {
	Fail(c, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "An unexpected error occurred.")
}
