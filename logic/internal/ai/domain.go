package ai

import (
	"context"
	"time"

	"github.com/english/backend/internal/ai/dto"
	"github.com/english/backend/internal/dictionary"
)

// -----------------------------------------------------------------------
// AIClient — Provider-agnostic interface.
// Implementations: GeminiClient, OpenAIClient, ClaudeClient, OllamaClient.
// -----------------------------------------------------------------------

// AIClient generates structured JSON responses from an AI provider.
// The caller passes a prompt string and a pointer to the target schema; the
// implementation is responsible for instructing the provider to return JSON
// that matches the schema, then unmarshalling the result into dst.
type AIClient interface {
	// GenerateJSON sends prompt to the AI provider and unmarshals the response
	// into dst. schema describes the expected JSON structure to the provider
	// (provider-specific — Gemini uses *genai.Schema, OpenAI uses json-schema).
	GenerateJSON(ctx context.Context, prompt string, schema any, dst any) error
	// GenerateMultimodalJSON sends prompt and a file blob to the AI provider.
	GenerateMultimodalJSON(ctx context.Context, prompt string, mimeType string, data []byte, schema any, dst any) error
	// GenerateText sends prompt to the AI provider and returns the raw string response.
	GenerateText(ctx context.Context, prompt string) (string, error)
}

// -----------------------------------------------------------------------
// Cache — pluggable response cache (Memory for MVP → Redis in production).
// -----------------------------------------------------------------------

type Cache interface {
	// Get retrieves a cached value and unmarshals it into dst.
	// Returns (true, nil) on hit; (false, nil) on miss.
	Get(ctx context.Context, key string, dst any) (bool, error)
	// Set stores value with the given TTL.
	Set(ctx context.Context, key string, value any, ttl time.Duration) error
	// Delete removes a key (useful for invalidation).
	Delete(ctx context.Context, key string) error
}

// -----------------------------------------------------------------------
// RateLimiter — pluggable rate limiter (Memory → Redis).
// -----------------------------------------------------------------------

type RateLimiter interface {
	// Allow returns nil when the request is within quota, or ErrTooManyRequests.
	Allow(ctx context.Context, userID string) error
}

// -----------------------------------------------------------------------
// ExplanationRepo — database cache for word explanations.
// -----------------------------------------------------------------------

type ExplanationRepo interface {
	GetAIExplanation(ctx context.Context, word, pos, level, nativeLang string) (*dictionary.AIExplanation, error)
	SaveAIExplanation(ctx context.Context, expl *dictionary.AIExplanation) error
}

// -----------------------------------------------------------------------
// AITutorService — high-level AI features exposed to handlers.
// -----------------------------------------------------------------------

type AITutorService interface {
	// Word-level features
	ExplainWord(ctx context.Context, promptCtx PromptContext) (*dto.ExplainResponse, error)
	ExplainNuance(ctx context.Context, promptCtx PromptContext) (*dto.NuanceResponse, error)
	CompareWords(ctx context.Context, promptCtx PromptContext, word2 string) (*dto.CompareResponse, error)
	GenerateExamples(ctx context.Context, promptCtx PromptContext) (*dto.ExamplesResponse, error)
	CreateMnemonic(ctx context.Context, promptCtx PromptContext) (*dto.MnemonicResponse, error)

	// Writing features
	CheckGrammar(ctx context.Context, promptCtx PromptContext, text string) (*dto.GrammarResponse, error)
	SuggestWritingVocab(ctx context.Context, promptCtx PromptContext, text string) (*dto.VocabSuggestionResponse, error)

	// Quiz features
	GenerateFillBlank(ctx context.Context, promptCtx PromptContext, words []dictionary.WordDocument) (*dto.FillBlankResponse, error)

	// Speaking features
	GenerateSpeakingTask(ctx context.Context, topic, taskType, difficulty string) (*dto.SpeakingTaskResponse, error)
	EvaluateSpeaking(ctx context.Context, audioBytes []byte, mimeType string, taskSnapshot any) (*dto.SpeakingEvaluationResponse, error)
}

// -----------------------------------------------------------------------
// PromptContext — carries user-level metadata so prompts can be personalised.
// -----------------------------------------------------------------------

// PromptContext is immutable data passed alongside every service call.
// It allows prompts to adapt output for different proficiency levels and
// native languages without changing service signatures.
type PromptContext struct {
	// Word is the primary vocabulary item being discussed.
	Word string
	// POS is the Part-of-Speech tag (noun, verb, adjective…).
	POS string
	// UserLevel is the CEFR level of the learner (A1, A2, B1, B2, C1, C2).
	// Default: B1.
	UserLevel string
	// NativeLanguage is the user's native language (e.g. "Vietnamese").
	// Used to tailor explanations and common-mistake notes.
	NativeLanguage string
}
