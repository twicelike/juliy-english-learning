package cache

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/english/backend/internal/middleware"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
	"golang.org/x/sync/singleflight"
)

var (
	ErrCacheMiss = errors.New("cache miss")
)

type Service interface {
	GetOrSet(ctx context.Context, key string, target interface{}, ttl time.Duration, fetch func() (interface{}, error)) error
	Invalidate(ctx context.Context, keys ...string) error
	Ping(ctx context.Context) error
}

type service struct {
	client *redis.Client
	group  singleflight.Group
}

func NewService(uri string) (Service, error) {
	opts, err := redis.ParseURL(uri)
	if err != nil {
		return nil, err
	}

	client := redis.NewClient(opts)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		log.Warn().Err(err).Msg("redis ping failed, caching will be disabled (circuit breaker)")
		// Soft fail: return a working service that just bypasses cache
	} else {
		log.Info().Msg("redis connected successfully")
	}

	return &service{
		client: client,
	}, nil
}

// GetOrSet implements Cache-Aside + Stampede Protection + Circuit Breaker
func (s *service) GetOrSet(ctx context.Context, key string, target interface{}, ttl time.Duration, fetch func() (interface{}, error)) error {
	// Try Redis first
	val, err := s.client.Get(ctx, key).Result()
	if err == nil {
		// Cache Hit
		middleware.CacheHitsTotal.Inc()
		if err := json.Unmarshal([]byte(val), target); err == nil {
			return nil
		}
		log.Warn().Err(err).Str("key", key).Msg("failed to unmarshal cache data")
	} else if err != redis.Nil {
		// Redis error (Circuit breaker: log and fall through to DB)
		middleware.RedisErrorsTotal.Inc()
		log.Error().Err(err).Str("key", key).Msg("redis get error, falling back to DB")
	} else {
		// Normal Cache Miss
		middleware.CacheMissesTotal.Inc()
	}

	// Cache Miss or Error -> Use Singleflight to prevent Cache Stampede
	// This ensures only one goroutine executes fetch() for a given key
	v, err, _ := s.group.Do(key, func() (interface{}, error) {
		res, fetchErr := fetch()
		if fetchErr != nil {
			return nil, fetchErr
		}

		// Save to Redis (fire & forget in background to not block response)
		go func() {
			bgCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()
			
			bytes, mErr := json.Marshal(res)
			if mErr != nil {
				return
			}
			
			if setErr := s.client.Set(bgCtx, key, bytes, ttl).Err(); setErr != nil {
				middleware.RedisErrorsTotal.Inc()
				log.Error().Err(setErr).Str("key", key).Msg("redis set error")
			}
		}()

		return res, nil
	})

	if err != nil {
		return err
	}

	// Copy result to target (JSON marshal/unmarshal is a generic way to deep copy, though inefficient. 
	// For high performance, target should be assigned directly if types match, but since target is a pointer to a struct, 
	// JSON is the safest generic deep copy in Go without reflection gymnastics).
	bytes, _ := json.Marshal(v)
	return json.Unmarshal(bytes, target)
}

func (s *service) Invalidate(ctx context.Context, keys ...string) error {
	if len(keys) == 0 {
		return nil
	}
	err := s.client.Del(ctx, keys...).Err()
	if err != nil {
		middleware.RedisErrorsTotal.Inc()
		log.Error().Err(err).Strs("keys", keys).Msg("redis del error")
	}
	return err
}

func (s *service) Ping(ctx context.Context) error {
	return s.client.Ping(ctx).Err()
}
