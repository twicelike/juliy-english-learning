package ai

import "errors"

var (
	// ErrTooManyRequests is returned by RateLimiter.Allow when the user has
	// exceeded the hourly quota.
	ErrTooManyRequests = errors.New("too many AI requests, please try again later")

	// ErrAIUnavailable is returned when the upstream AI provider is unreachable
	// after all retries have been exhausted.
	ErrAIUnavailable = errors.New("AI service is temporarily unavailable")

	// ErrInvalidAIResponse is returned when the AI response fails validation.
	ErrInvalidAIResponse = errors.New("AI returned an invalid response")
)
