package ai

import (
	"context"
	"sync"
	"time"
)

// MemoryRateLimiter implements RateLimiter with an in-process sliding window.
// Replace with a Redis-based limiter in production.
type MemoryRateLimiter struct {
	mu       sync.Mutex
	windows  map[string][]time.Time // userID -> sorted list of request times
	maxPerHr int
}

// NewMemoryRateLimiter creates a MemoryRateLimiter with the given hourly cap.
func NewMemoryRateLimiter(maxPerHour int) *MemoryRateLimiter {
	return &MemoryRateLimiter{
		windows:  make(map[string][]time.Time),
		maxPerHr: maxPerHour,
	}
}

// Allow returns nil when the request is within quota, or ErrTooManyRequests.
func (r *MemoryRateLimiter) Allow(_ context.Context, userID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	horizon := now.Add(-time.Hour)

	// Prune expired timestamps.
	times := r.windows[userID]
	start := 0
	for start < len(times) && times[start].Before(horizon) {
		start++
	}
	times = times[start:]

	if len(times) >= r.maxPerHr {
		r.windows[userID] = times
		return ErrTooManyRequests
	}

	r.windows[userID] = append(times, now)
	return nil
}
