package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// -----------------------------------------------------------------------
// MemoryCache
// -----------------------------------------------------------------------

type cacheEntry struct {
	data      []byte
	expiresAt time.Time
}

// MemoryCache is a simple in-process Cache implementation suitable for MVP.
// It is safe for concurrent use.
type MemoryCache struct {
	mu    sync.RWMutex
	store map[string]cacheEntry
}

// NewMemoryCache creates an empty MemoryCache.
func NewMemoryCache() *MemoryCache {
	return &MemoryCache{store: make(map[string]cacheEntry)}
}

// Get retrieves a cached value. Returns (true, nil) on hit; (false, nil) on miss.
func (c *MemoryCache) Get(_ context.Context, key string, dst any) (bool, error) {
	c.mu.RLock()
	entry, ok := c.store[key]
	c.mu.RUnlock()

	if !ok || time.Now().After(entry.expiresAt) {
		return false, nil
	}
	if err := json.Unmarshal(entry.data, dst); err != nil {
		return false, fmt.Errorf("cache: unmarshal: %w", err)
	}
	return true, nil
}

// Set stores value serialised as JSON with the given TTL.
func (c *MemoryCache) Set(_ context.Context, key string, value any, ttl time.Duration) error {
	b, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("cache: marshal: %w", err)
	}
	c.mu.Lock()
	c.store[key] = cacheEntry{data: b, expiresAt: time.Now().Add(ttl)}
	c.mu.Unlock()
	return nil
}

// Delete removes a key from the cache.
func (c *MemoryCache) Delete(_ context.Context, key string) error {
	c.mu.Lock()
	delete(c.store, key)
	c.mu.Unlock()
	return nil
}
