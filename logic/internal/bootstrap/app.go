package bootstrap

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/english/backend/internal/cache"
	"github.com/english/backend/internal/config"
	"github.com/english/backend/internal/database"
	"github.com/english/backend/internal/router"
	"github.com/rs/zerolog/log"
)

// Start wires all dependencies and runs the HTTP server.
// It is the single place where the entire application is assembled.
// main.go only calls this function.
func Start() error {
	// 1. Load configuration
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("config: %w", err)
	}
	log.Info().Str("env", cfg.Env).Str("port", cfg.Port).Msg("configuration loaded")

	// 2. Connect to MongoDB
	mongoClient, err := database.Connect(cfg.MongoURI)
	if err != nil {
		return fmt.Errorf("database: %w", err)
	}
	log.Info().Msg("mongodb connected")
	defer func() {
		if err := database.Disconnect(mongoClient); err != nil {
			log.Error().Err(err).Msg("mongodb disconnect error")
		}
	}()

	// 3. Connect to Redis
	cacheSvc, err := cache.NewService(cfg.RedisURI)
	if err != nil {
		log.Warn().Err(err).Msg("redis connection failed, cache disabled")
	}

	// Create a background context that gets cancelled on shutdown
	ctx, cancelApp := context.WithCancel(context.Background())
	defer cancelApp()

	// 4. Build router (all middleware + routes)
	r := router.Setup(ctx, cfg, mongoClient, cacheSvc)

	// 5. Start HTTP server with graceful shutdown
	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      r,
		ReadTimeout:  180 * time.Second,
		WriteTimeout: 180 * time.Second,
		IdleTimeout:  180 * time.Second,
	}

	// Run server in a goroutine so we can listen for shutdown signals concurrently.
	errCh := make(chan error, 1)
	go func() {
		log.Info().Str("addr", srv.Addr).Msg("server starting")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
	}()

	// Wait for OS signal or server error
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-errCh:
		return fmt.Errorf("server: %w", err)
	case sig := <-quit:
		log.Info().Str("signal", sig.String()).Msg("shutdown signal received")
	}

	// Graceful shutdown — give in-flight requests up to 10s to finish
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		return fmt.Errorf("server shutdown: %w", err)
	}

	log.Info().Msg("server stopped gracefully")
	return nil
}
