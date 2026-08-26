package client

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"time"

	"github.com/english/backend/internal/ai"
)

const (
	defaultMaxRetries = 2
	defaultBaseDelay  = 500 * time.Millisecond
)

// retryClient wraps an AIClient and retries on transient errors.
type retryClient struct {
	inner      ai.AIClient
	maxRetries int
	baseDelay  time.Duration
}

// WithRetry wraps an AIClient with exponential-backoff retry logic.
// Retries up to maxRetries times on 429 / 500 / 503 / timeout errors.
func WithRetry(inner ai.AIClient, maxRetries int) ai.AIClient {
	if maxRetries <= 0 {
		maxRetries = defaultMaxRetries
	}
	return &retryClient{inner: inner, maxRetries: maxRetries, baseDelay: defaultBaseDelay}
}

func (r *retryClient) GenerateJSON(ctx context.Context, prompt string, schema any, dst any) error {
	var lastErr error
	delay := r.baseDelay

	for attempt := 0; attempt <= r.maxRetries; attempt++ {
		err := r.inner.GenerateJSON(ctx, prompt, schema, dst)
		if err == nil {
			return nil
		}
		lastErr = err

		if !isRetryable(err) {
			return err
		}

		if attempt < r.maxRetries {
			slog.Warn("AI request failed, retrying", "attempt", attempt+1, "delay_ms", delay.Milliseconds(), "error", err)
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(delay):
			}
			delay *= 2 // exponential back-off
		}
	}
	return lastErr
}

func (r *retryClient) GenerateMultimodalJSON(ctx context.Context, prompt string, mimeType string, data []byte, schema any, dst any) error {
	var lastErr error
	delay := r.baseDelay

	for attempt := 0; attempt <= r.maxRetries; attempt++ {
		err := r.inner.GenerateMultimodalJSON(ctx, prompt, mimeType, data, schema, dst)
		if err == nil {
			return nil
		}
		lastErr = err

		if !isRetryable(err) {
			return err
		}

		if attempt < r.maxRetries {
			slog.Warn("AI multimodal request failed, retrying", "attempt", attempt+1, "delay_ms", delay.Milliseconds(), "error", err)
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(delay):
			}
			delay *= 2 // exponential back-off
		}
	}
	return lastErr
}

// isRetryable returns true for transient errors worth retrying.
func isRetryable(err error) bool {
	if err == nil {
		return false
	}
	// Context cancellation is not retryable.
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return false
	}
	// Network timeout → retry
	var netErr net.Error
	if errors.As(err, &netErr) && netErr.Timeout() {
		return true
	}
	// Rely on error message heuristics for HTTP-level codes embedded by the SDK.
	msg := err.Error()
	for _, marker := range []string{"429", "500", "503", "deadline exceeded", "timeout"} {
		if containsFold(msg, marker) {
			return true
		}
	}
	return false
}

func containsFold(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(s) > 0 && searchInsensitive(s, sub))
}

func searchInsensitive(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if equalFold(s[i:i+len(sub)], sub) {
			return true
		}
	}
	return false
}

func equalFold(a, b string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := 0; i < len(a); i++ {
		ca, cb := a[i], b[i]
		if ca >= 'A' && ca <= 'Z' {
			ca += 32
		}
		if cb >= 'A' && cb <= 'Z' {
			cb += 32
		}
		if ca != cb {
			return false
		}
	}
	return true
}

func (r *retryClient) GenerateText(ctx context.Context, prompt string) (string, error) {
	var resp string
	var err error
	for i := 0; i <= r.maxRetries; i++ {
		resp, err = r.inner.GenerateText(ctx, prompt)
		if err == nil {
			return resp, nil
		}
		if i < r.maxRetries {
			time.Sleep(time.Second * time.Duration(1<<i))
		}
	}
	return resp, err
}
