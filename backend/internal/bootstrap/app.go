package bootstrap

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/english/backend/internal/config"
	"github.com/english/backend/internal/database"
	"github.com/english/backend/internal/router"
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
	slog.Info("configuration loaded", "env", cfg.Env, "port", cfg.Port)

	// 2. Connect to MongoDB
	mongoClient, err := database.Connect(cfg.MongoURI)
	if err != nil {
		return fmt.Errorf("database: %w", err)
	}
	slog.Info("mongodb connected")
	defer func() {
		if err := database.Disconnect(mongoClient); err != nil {
			slog.Error("mongodb disconnect error", "error", err)
		}
	}()

	// 3. Build router (all middleware + routes)
	r := router.Setup(cfg, mongoClient)

	// 4. Start HTTP server with graceful shutdown
	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Run server in a goroutine so we can listen for shutdown signals concurrently.
	errCh := make(chan error, 1)
	go func() {
		slog.Info("server starting", "addr", srv.Addr)
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
		slog.Info("shutdown signal received", "signal", sig)
	}

	// Graceful shutdown — give in-flight requests up to 10s to finish
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		return fmt.Errorf("server shutdown: %w", err)
	}

	slog.Info("server stopped gracefully")
	return nil
}
