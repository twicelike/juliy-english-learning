package client

import (
	"context"
	"log/slog"
	"time"

	"github.com/english/backend/internal/ai"
)

// metricsClient wraps an AIClient and emits structured log-based metrics.
// Replace log lines with Prometheus / OpenTelemetry counters when ready.
type metricsClient struct {
	inner ai.AIClient
}

// WithMetrics wraps an AIClient and tracks latency, errors, and call count.
func WithMetrics(inner ai.AIClient) ai.AIClient {
	return &metricsClient{inner: inner}
}

func (m *metricsClient) GenerateJSON(ctx context.Context, prompt string, schema any, dst any) error {
	start := time.Now()

	err := m.inner.GenerateJSON(ctx, prompt, schema, dst)

	latency := time.Since(start)
	if err != nil {
		slog.Error("AI request failed",
			"latency_ms", latency.Milliseconds(),
			"error", err,
		)
	} else {
		slog.Info("AI request completed",
			"latency_ms", latency.Milliseconds(),
		)
	}

	return err
}

func (m *metricsClient) GenerateMultimodalJSON(ctx context.Context, prompt string, mimeType string, data []byte, schema any, dst any) error {
	start := time.Now()

	err := m.inner.GenerateMultimodalJSON(ctx, prompt, mimeType, data, schema, dst)

	latency := time.Since(start)
	if err != nil {
		slog.Error("AI multimodal request failed",
			"latency_ms", latency.Milliseconds(),
			"error", err,
		)
	} else {
		slog.Info("AI multimodal request completed",
			"latency_ms", latency.Milliseconds(),
		)
	}

	return err
}

func (m *metricsClient) GenerateText(ctx context.Context, prompt string) (string, error) {
	// Simple implementation
	return m.inner.GenerateText(ctx, prompt)
}
