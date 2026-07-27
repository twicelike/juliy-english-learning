package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/english/backend/internal/domain"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

// HealthHandler holds the MongoDB client needed to verify DB connectivity.
type HealthHandler struct {
	mongo *mongo.Client
}

// NewHealthHandler creates a HealthHandler with the given mongo client.
func NewHealthHandler(client *mongo.Client) *HealthHandler {
	return &HealthHandler{mongo: client}
}

// Check handles GET /api/v1/health.
// It pings MongoDB and reports the result to the caller.
func (h *HealthHandler) Check(c *gin.Context) {
	dbStatus := "connected"

	ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
	defer cancel()

	if err := h.mongo.Ping(ctx, nil); err != nil {
		domain.Fail(c, http.StatusServiceUnavailable, "DB_UNREACHABLE",
			"Database is not reachable.")
		return
	}

	domain.OK(c, gin.H{"status": "ok", "db": dbStatus})
}
