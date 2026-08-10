package middleware

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	HttpRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests processed, partitioned by status code, method and path.",
		},
		[]string{"code", "method", "path"},
	)

	HttpRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Latency of HTTP requests in seconds.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path"},
	)

	HttpRequestsInFlight = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "http_requests_in_flight",
			Help: "Current number of HTTP requests being processed.",
		},
		[]string{"method", "path"},
	)

	HttpRequestsFailedTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_failed_total",
			Help: "Total number of failed HTTP requests (status >= 400).",
		},
		[]string{"code", "method", "path"},
	)

	// App-specific metrics
	CacheHitsTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "cache_hits_total",
		Help: "Total number of cache hits.",
	})
	
	CacheMissesTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "cache_misses_total",
		Help: "Total number of cache misses.",
	})
	
	RedisErrorsTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "redis_errors_total",
		Help: "Total number of redis errors.",
	})
)

// Metrics records Prometheus metrics for each request.
func Metrics() gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.Request.URL.Path
		method := c.Request.Method

		// Ignore metrics for the /metrics and /health endpoints
		if path == "/metrics" || path == "/live" || path == "/ready" || path == "/health" {
			c.Next()
			return
		}

		HttpRequestsInFlight.WithLabelValues(method, path).Inc()
		defer HttpRequestsInFlight.WithLabelValues(method, path).Dec()

		start := time.Now()
		c.Next()
		duration := time.Since(start).Seconds()
		status := strconv.Itoa(c.Writer.Status())

		HttpRequestsTotal.WithLabelValues(status, method, path).Inc()
		HttpRequestDuration.WithLabelValues(method, path).Observe(duration)

		if c.Writer.Status() >= 400 {
			HttpRequestsFailedTotal.WithLabelValues(status, method, path).Inc()
		}
	}
}
