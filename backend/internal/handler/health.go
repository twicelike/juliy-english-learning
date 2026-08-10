package handler

import (
	"context"
	"net/http"
	"runtime"
	"time"

	"github.com/english/backend/internal/cache"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

var startTime time.Time

func init() {
	startTime = time.Now()
}

type HealthHandler struct {
	mongoClient *mongo.Client
	cacheSvc    cache.Service
	version     string
}

func NewHealthHandler(mongoClient *mongo.Client, cacheSvc cache.Service, version string) *HealthHandler {
	return &HealthHandler{
		mongoClient: mongoClient,
		cacheSvc:    cacheSvc,
		version:     version,
	}
}

// Live simply returns 200 OK. Used by Kubernetes/Docker to know if the process is alive.
func (h *HealthHandler) Live(c *gin.Context) {
	c.String(http.StatusOK, "OK")
}

// Ready checks if the application is ready to accept traffic (checks connections to DB and Redis).
func (h *HealthHandler) Ready(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := h.mongoClient.Ping(ctx, nil); err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"status": "error", "message": "database not ready"})
		return
	}

	if h.cacheSvc != nil {
		if err := h.cacheSvc.Ping(ctx); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "error", "message": "cache not ready"})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"status": "ready"})
}

// Health returns detailed information about the system.
func (h *HealthHandler) Health(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	mongoStatus := "ok"
	if err := h.mongoClient.Ping(ctx, nil); err != nil {
		mongoStatus = "down"
	}

	redisStatus := "ok"
	if h.cacheSvc != nil {
		if err := h.cacheSvc.Ping(ctx); err != nil {
			redisStatus = "down"
		}
	} else {
		redisStatus = "disabled"
	}

	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	c.JSON(http.StatusOK, gin.H{
		"version":    h.version,
		"uptime":     time.Since(startTime).String(),
		"mongo":      mongoStatus,
		"redis":      redisStatus,
		"goroutines": runtime.NumGoroutine(),
		"memory": gin.H{
			"alloc_mb":      m.Alloc / 1024 / 1024,
			"total_alloc_mb": m.TotalAlloc / 1024 / 1024,
			"sys_mb":        m.Sys / 1024 / 1024,
			"num_gc":        m.NumGC,
		},
	})
}
