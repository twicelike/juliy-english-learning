package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

// IPRateLimiter is a simple memory-based rate limiter per IP address
func IPRateLimiter(requestsPerMin int) gin.HandlerFunc {
	type client struct {
		limiter  *rate.Limiter
		lastSeen time.Time
	}

	var (
		mu      sync.Mutex
		clients = make(map[string]*client)
	)

	// Background routine to cleanup old clients every minute
	go func() {
		for {
			time.Sleep(time.Minute)
			mu.Lock()
			for ip, c := range clients {
				if time.Since(c.lastSeen) > 3*time.Minute {
					delete(clients, ip)
				}
			}
			mu.Unlock()
		}
	}()

	limit := rate.Every(time.Minute / time.Duration(requestsPerMin))

	return func(c *gin.Context) {
		ip := c.ClientIP()
		if ip == "" {
			ip = "unknown"
		}

		mu.Lock()
		if _, found := clients[ip]; !found {
			clients[ip] = &client{
				limiter: rate.NewLimiter(limit, requestsPerMin), // Burst allowed is same as request per min
			}
		}
		clients[ip].lastSeen = time.Now()
		limiter := clients[ip].limiter
		mu.Unlock()

		if !limiter.Allow() {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "Rate limit exceeded"})
			return
		}

		c.Next()
	}
}

// KeyRateLimiter is a memory-based rate limiter per extracted key (e.g., IP + email)
func KeyRateLimiter(requestsPerMin int, keyExtractor func(c *gin.Context) string) gin.HandlerFunc {
	type client struct {
		limiter  *rate.Limiter
		lastSeen time.Time
	}

	var (
		mu      sync.Mutex
		clients = make(map[string]*client)
	)

	// Background routine to cleanup old clients every minute
	go func() {
		for {
			time.Sleep(time.Minute)
			mu.Lock()
			for key, c := range clients {
				if time.Since(c.lastSeen) > 3*time.Minute {
					delete(clients, key)
				}
			}
			mu.Unlock()
		}
	}()

	limit := rate.Every(time.Minute / time.Duration(requestsPerMin))

	return func(c *gin.Context) {
		key := keyExtractor(c)
		if key == "" {
			key = "unknown"
		}

		mu.Lock()
		if _, found := clients[key]; !found {
			clients[key] = &client{
				limiter: rate.NewLimiter(limit, requestsPerMin),
			}
		}
		clients[key].lastSeen = time.Now()
		limiter := clients[key].limiter
		mu.Unlock()

		if !limiter.Allow() {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "Rate limit exceeded"})
			return
		}

		c.Next()
	}
}
