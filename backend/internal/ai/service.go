package ai

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/english/backend/internal/ai/dto"
	"github.com/english/backend/internal/ai/prompts"
	"github.com/english/backend/internal/dictionary"
)

// cacheTTLGlobal is used for AI responses that do not depend on user identity.
// Example: explain:en:book — always the same regardless of who asks.
const cacheTTLGlobal = 24 * time.Hour

type service struct {
	client      AIClient
	cache       Cache
	rateLimiter RateLimiter
}

// NewService constructs an AITutorService with the given dependencies.
func NewService(client AIClient, cache Cache, rateLimiter RateLimiter) AITutorService {
	return &service{
		client:      client,
		cache:       cache,
		rateLimiter: rateLimiter,
	}
}

// -----------------------------------------------------------------------
// Word features
// -----------------------------------------------------------------------

func (s *service) ExplainWord(ctx context.Context, pctx PromptContext) (*dto.ExplainResponse, error) {
	key := globalKey("explain", pctx)
	var resp dto.ExplainResponse
	if hit, _ := s.cache.Get(ctx, key, &resp); hit {
		slog.Info("AI cache hit", "key", key)
		return &resp, nil
	}

	prompt := prompts.BuildExplainPrompt(toPromptCtx(pctx))
	if err := s.client.GenerateJSON(ctx, prompt, nil, &resp); err != nil {
		return nil, fmt.Errorf("ai: explain word: %w", err)
	}
	if err := resp.Validate(); err != nil {
		return nil, fmt.Errorf("%w: %s", ErrInvalidAIResponse, err)
	}

	_ = s.cache.Set(ctx, key, &resp, cacheTTLGlobal)
	return &resp, nil
}

func (s *service) ExplainNuance(ctx context.Context, pctx PromptContext) (*dto.NuanceResponse, error) {
	key := globalKey("nuance", pctx)
	var resp dto.NuanceResponse
	if hit, _ := s.cache.Get(ctx, key, &resp); hit {
		slog.Info("AI cache hit", "key", key)
		return &resp, nil
	}

	prompt := prompts.BuildNuancePrompt(toPromptCtx(pctx))
	if err := s.client.GenerateJSON(ctx, prompt, nil, &resp); err != nil {
		return nil, fmt.Errorf("ai: explain nuance: %w", err)
	}
	if err := resp.Validate(); err != nil {
		return nil, fmt.Errorf("%w: %s", ErrInvalidAIResponse, err)
	}

	_ = s.cache.Set(ctx, key, &resp, cacheTTLGlobal)
	return &resp, nil
}

func (s *service) CompareWords(ctx context.Context, pctx PromptContext, word2 string) (*dto.CompareResponse, error) {
	key := fmt.Sprintf("compare:en:%s:%s", strings.ToLower(pctx.Word), strings.ToLower(word2))
	var resp dto.CompareResponse
	if hit, _ := s.cache.Get(ctx, key, &resp); hit {
		slog.Info("AI cache hit", "key", key)
		return &resp, nil
	}

	prompt := prompts.BuildComparePrompt(toPromptCtx(pctx), word2)
	if err := s.client.GenerateJSON(ctx, prompt, nil, &resp); err != nil {
		return nil, fmt.Errorf("ai: compare words: %w", err)
	}
	if err := resp.Validate(); err != nil {
		return nil, fmt.Errorf("%w: %s", ErrInvalidAIResponse, err)
	}

	_ = s.cache.Set(ctx, key, &resp, cacheTTLGlobal)
	return &resp, nil
}

func (s *service) GenerateExamples(ctx context.Context, pctx PromptContext) (*dto.ExamplesResponse, error) {
	key := globalKey("examples", pctx)
	var resp dto.ExamplesResponse
	if hit, _ := s.cache.Get(ctx, key, &resp); hit {
		slog.Info("AI cache hit", "key", key)
		return &resp, nil
	}

	prompt := prompts.BuildExamplesPrompt(toPromptCtx(pctx))
	if err := s.client.GenerateJSON(ctx, prompt, nil, &resp); err != nil {
		return nil, fmt.Errorf("ai: generate examples: %w", err)
	}
	if err := resp.Validate(); err != nil {
		return nil, fmt.Errorf("%w: %s", ErrInvalidAIResponse, err)
	}

	_ = s.cache.Set(ctx, key, &resp, cacheTTLGlobal)
	return &resp, nil
}

func (s *service) CreateMnemonic(ctx context.Context, pctx PromptContext) (*dto.MnemonicResponse, error) {
	key := globalKey("mnemonic", pctx)
	var resp dto.MnemonicResponse
	if hit, _ := s.cache.Get(ctx, key, &resp); hit {
		slog.Info("AI cache hit", "key", key)
		return &resp, nil
	}

	prompt := prompts.BuildMnemonicPrompt(toPromptCtx(pctx))
	if err := s.client.GenerateJSON(ctx, prompt, nil, &resp); err != nil {
		return nil, fmt.Errorf("ai: create mnemonic: %w", err)
	}
	if err := resp.Validate(); err != nil {
		return nil, fmt.Errorf("%w: %s", ErrInvalidAIResponse, err)
	}

	_ = s.cache.Set(ctx, key, &resp, cacheTTLGlobal)
	return &resp, nil
}

// -----------------------------------------------------------------------
// Writing features (no global cache — text is user-specific)
// -----------------------------------------------------------------------

func (s *service) CheckGrammar(ctx context.Context, pctx PromptContext, text string) (*dto.GrammarResponse, error) {
	prompt := prompts.BuildGrammarPrompt(toPromptCtx(pctx), text)
	var resp dto.GrammarResponse
	if err := s.client.GenerateJSON(ctx, prompt, nil, &resp); err != nil {
		return nil, fmt.Errorf("ai: check grammar: %w", err)
	}
	if err := resp.Validate(); err != nil {
		return nil, fmt.Errorf("%w: %s", ErrInvalidAIResponse, err)
	}
	return &resp, nil
}

func (s *service) SuggestWritingVocab(ctx context.Context, pctx PromptContext, text string) (*dto.VocabSuggestionResponse, error) {
	prompt := prompts.BuildVocabSuggestionPrompt(toPromptCtx(pctx), text)
	var resp dto.VocabSuggestionResponse
	if err := s.client.GenerateJSON(ctx, prompt, nil, &resp); err != nil {
		return nil, fmt.Errorf("ai: suggest vocab: %w", err)
	}
	if err := resp.Validate(); err != nil {
		return nil, fmt.Errorf("%w: %s", ErrInvalidAIResponse, err)
	}
	return &resp, nil
}

// -----------------------------------------------------------------------
// Quiz features (user-specific — no global cache)
// -----------------------------------------------------------------------

func (s *service) GenerateFillBlank(ctx context.Context, pctx PromptContext, words []dictionary.WordDocument) (*dto.FillBlankResponse, error) {
	wordInputs := make([]prompts.WordInput, 0, len(words))
	for _, w := range words {
		wordInputs = append(wordInputs, prompts.WordInput{
			Word: w.Word,
			POS:  w.POS,
		})
	}

	prompt := prompts.BuildFillBlankPrompt(toPromptCtx(pctx), wordInputs)
	var resp dto.FillBlankResponse
	if err := s.client.GenerateJSON(ctx, prompt, nil, &resp); err != nil {
		return nil, fmt.Errorf("ai: generate fill blank: %w", err)
	}
	if err := resp.Validate(); err != nil {
		return nil, fmt.Errorf("%w: %s", ErrInvalidAIResponse, err)
	}
	return &resp, nil
}

// -----------------------------------------------------------------------
// Helpers
// -----------------------------------------------------------------------

// globalKey builds a cache key that is independent of the user identity.
// Format: action:lang:word[:pos]
func globalKey(action string, pctx PromptContext) string {
	parts := []string{action, "en", strings.ToLower(pctx.Word)}
	if pctx.POS != "" {
		parts = append(parts, strings.ToLower(pctx.POS))
	}
	return strings.Join(parts, ":")
}

// toPromptCtx converts an ai.PromptContext to prompts.PromptContext.
func toPromptCtx(p PromptContext) prompts.PromptContext {
	return prompts.PromptContext{
		Word:           p.Word,
		POS:            p.POS,
		UserLevel:      p.UserLevel,
		NativeLanguage: p.NativeLanguage,
	}
}
