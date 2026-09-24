package ai

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"math/rand"
	"strings"
	"time"

	"github.com/english/backend/internal/ai/dto"
	"github.com/english/backend/internal/ai/prompts"
	"github.com/english/backend/internal/dictionary"
	"go.mongodb.org/mongo-driver/v2/bson"
)

// cacheTTLGlobal is used for AI responses that do not depend on user identity.
// Example: explain:en:book — always the same regardless of who asks.
const cacheTTLGlobal = 24 * time.Hour

type service struct {
	client      AIClient
	cache       Cache
	explRepo    ExplanationRepo
	dictRepo    dictionary.Repository
	rateLimiter RateLimiter
}

// NewService constructs an AITutorService with the given dependencies.
func NewService(client AIClient, cache Cache, explRepo ExplanationRepo, dictRepo dictionary.Repository, rateLimiter RateLimiter) AITutorService {
	return &service{
		client:      client,
		cache:       cache,
		explRepo:    explRepo,
		dictRepo:    dictRepo,
		rateLimiter: rateLimiter,
	}
}

// -----------------------------------------------------------------------
// Word features
// -----------------------------------------------------------------------

func (s *service) ExplainSense(ctx context.Context, wordID string, senseIndex int, userLevel, nativeLang string) (*dto.BatchSenseExplanation, error) {
	const ExplainPromptVersion = 3

	// 1. Try to fetch from database cache first
	cached, err := s.explRepo.GetSenseExplanation(ctx, wordID, senseIndex, userLevel, nativeLang, ExplainPromptVersion)
	if err == nil && cached != nil {
		slog.Info("AI DB cache hit", "wordId", wordID, "senseIndex", senseIndex, "level", userLevel, "version", ExplainPromptVersion)
		return &dto.BatchSenseExplanation{
			SenseIndex:     cached.SenseIndex,
			Explanation:    cached.Explanation,
			Examples:       cached.Examples,
			CommonMistakes: cached.CommonMistakes,
			UsageNote:      cached.UsageNote,
		}, nil
	}

	// 2. Fetch dictionary document
	objID, err := bson.ObjectIDFromHex(wordID)
	if err != nil {
		return nil, fmt.Errorf("invalid word id: %w", err)
	}
	wordDoc, err := s.dictRepo.GetByID(ctx, objID)
	if err != nil {
		return nil, fmt.Errorf("failed to get word from dictionary: %w", err)
	}
	if wordDoc == nil {
		return nil, errors.New("word not found in dictionary")
	}

	if senseIndex < 0 || senseIndex >= len(wordDoc.Senses) {
		return nil, errors.New("invalid sense index for word")
	}
	targetSense := wordDoc.Senses[senseIndex]
	if len(targetSense.Glosses) == 0 {
		return nil, errors.New("target sense has no definition")
	}

	// 3. Fetch cached sense indices to find unexplained senses
	cachedIndices, err := s.explRepo.GetCachedSenseIndices(ctx, wordID, userLevel, nativeLang, ExplainPromptVersion)
	if err != nil {
		slog.Warn("Failed to get cached sense indices", "err", err)
		cachedIndices = make(map[int]bool)
	}

	// 4. Select candidate senses (max 4 + 1 target)
	var candidates []prompts.SensePromptData
	for i, sense := range wordDoc.Senses {
		if i == senseIndex || cachedIndices[i] || len(sense.Glosses) == 0 {
			continue
		}
		// Skip ineligible senses (NICHE, RARE, IGNORE) — unknown quality is allowed
		if sense.Quality != nil {
			c := sense.Quality.Category
			if c != dictionary.CategoryPrimary && c != dictionary.CategoryCore && c != dictionary.CategoryUseful {
				continue
			}
		}
		candidates = append(candidates, prompts.SensePromptData{
			Index:      i,
			Definition: strings.Join(sense.Glosses, " | "), // Join all glosses for full context
		})
	}

	// Shuffle candidates and pick max 4
	if len(candidates) > 4 {
		rand.Shuffle(len(candidates), func(i, j int) {
			candidates[i], candidates[j] = candidates[j], candidates[i]
		})
		candidates = candidates[:4]
	}

	// Prepend the target sense
	batchSenses := append([]prompts.SensePromptData{{
		Index:      senseIndex,
		Definition: strings.Join(targetSense.Glosses, " | "),
	}}, candidates...)

	var requestedIndices []int
	for _, s := range batchSenses {
		requestedIndices = append(requestedIndices, s.Index)
	}

	slog.Info("AI generating batch explanations", "wordId", wordID, "batchSize", len(batchSenses), "indices", requestedIndices)

	// 5. Generate with AI
	pctx := prompts.PromptContext{
		Word:           wordDoc.Word,
		POS:            wordDoc.POS,
		UserLevel:      userLevel,
		NativeLanguage: nativeLang,
	}
	prompt := prompts.BuildBatchExplainPrompt(pctx, batchSenses)
	var resp dto.BatchExplainResponse
	if err := s.client.GenerateJSON(ctx, prompt, nil, &resp); err != nil {
		return nil, fmt.Errorf("ai: generate batch explanation: %w", err)
	}

	// 6. Validate exactly matches the requested indices
	if err := resp.Validate(requestedIndices); err != nil {
		// DO NOT save to cache if validation fails
		return nil, fmt.Errorf("%w: %s", ErrInvalidAIResponse, err)
	}

	// 7. Save to database for future requests
	var toSave []SenseExplanation
	var targetResponse *dto.BatchSenseExplanation

	for _, exp := range resp.Explanations {
		toSave = append(toSave, SenseExplanation{
			WordID:         wordID,
			SenseIndex:     exp.SenseIndex,
			UserLevel:      userLevel,
			NativeLanguage: nativeLang,
			Explanation:    exp.Explanation,
			Examples:       exp.Examples,
			CommonMistakes: exp.CommonMistakes,
			UsageNote:      exp.UsageNote,
			PromptVersion:  ExplainPromptVersion,
		})

		if exp.SenseIndex == senseIndex {
			expCopy := exp // Copy to avoid pointer issues
			targetResponse = &expCopy
		}
	}

	if err := s.explRepo.SaveBatchSenseExplanations(ctx, toSave); err != nil {
		slog.Error("Failed to save batch explanations", "err", err)
	}

	if targetResponse == nil {
		// Should not happen due to validation, but just in case
		return nil, fmt.Errorf("ai response missing target sense %d", senseIndex)
	}

	return targetResponse, nil
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
		Word:            p.Word,
		POS:             p.POS,
		UserLevel:       p.UserLevel,
		NativeLanguage:  p.NativeLanguage,
		ContextSentence: p.ContextSentence,
	}
}

// -----------------------------------------------------------------------
// Speaking features
// -----------------------------------------------------------------------

func (s *service) GenerateSpeakingTask(ctx context.Context, topic, taskType, difficulty string) (*dto.SpeakingTaskResponse, error) {
	prompt := prompts.BuildSpeakingTaskPrompt(topic, taskType, difficulty)
	var resp dto.SpeakingTaskResponse
	if err := s.client.GenerateJSON(ctx, prompt, nil, &resp); err != nil {
		return nil, fmt.Errorf("ai: generate speaking task: %w", err)
	}
	if err := resp.Validate(); err != nil {
		return nil, fmt.Errorf("%w: %s", ErrInvalidAIResponse, err)
	}

	// Business validation
	if taskType == "scenario" && resp.Scenario == "" {
		return nil, fmt.Errorf("%w: scenario is required for scenario task type", ErrInvalidAIResponse)
	}

	return &resp, nil
}

func (s *service) EvaluateSpeaking(ctx context.Context, audioBytes []byte, mimeType string, taskSnapshot any) (*dto.SpeakingEvaluationResponse, error) {
	snapshotBytes, err := json.Marshal(taskSnapshot)
	if err != nil {
		return nil, fmt.Errorf("ai: failed to marshal task snapshot: %w", err)
	}

	prompt := prompts.BuildSpeakingEvaluationPrompt(string(snapshotBytes))
	var resp dto.SpeakingEvaluationResponse

	err = s.client.GenerateMultimodalJSON(ctx, prompt, mimeType, audioBytes, nil, &resp)
	if err != nil {
		return nil, fmt.Errorf("ai: evaluate speaking: %w", err)
	}

	return &resp, nil
}
