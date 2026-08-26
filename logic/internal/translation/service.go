package translation

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/english/backend/internal/ai"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type MeaningInput struct {
	SenseID    string
	Definition string
}

type TranslateWordInput struct {
	WordID         string
	Word           string
	PartOfSpeech   string
	Meanings       []MeaningInput
	TargetLanguage string
}

type TranslateBatchInput struct {
	Words          []TranslateWordInput
	TargetLanguage string
}

type Service interface {
	TranslateWord(ctx context.Context, input TranslateWordInput) ([]string, error)
	TranslateBatch(ctx context.Context, input TranslateBatchInput) ([]WordTranslation, error)
	EnqueueJobs(ctx context.Context, wordIDs []string, targetLang string) error
	GetTranslationsByWordIDs(ctx context.Context, wordIDs []bson.ObjectID, targetLang string) ([]WordTranslation, error)
}

type service struct {
	aiClient ai.AIClient
	repo     Repository
}

func NewService(aiClient ai.AIClient, repo Repository) Service {
	return &service{aiClient: aiClient, repo: repo}
}

func (s *service) TranslateWord(ctx context.Context, input TranslateWordInput) ([]string, error) {
	if s.aiClient == nil {
		return nil, fmt.Errorf("AI client is not configured")
	}
	if input.Word == "" || input.TargetLanguage == "" {
		return nil, fmt.Errorf("word and target language are required")
	}
	if !IsSupportedLanguage(input.TargetLanguage) {
		return nil, fmt.Errorf("unsupported target language: %s", input.TargetLanguage)
	}

	prompt := s.buildPrompt(input)
	resp, err := s.aiClient.GenerateText(ctx, prompt)
	if err != nil {
		return nil, fmt.Errorf("ai generate failed: %w", err)
	}
	if strings.TrimSpace(resp) == "" {
		return nil, fmt.Errorf("ai returned empty response")
	}

	candidates, err := extractJSONCandidates(resp)
	if err != nil {
		return nil, fmt.Errorf("failed to extract candidates from ai response: %w", err)
	}

	if len(candidates) == 0 {
		return nil, fmt.Errorf("ai returned empty candidates array")
	}

	return candidates, nil
}

func (s *service) buildPrompt(input TranslateWordInput) string {
	var meaningsStr strings.Builder
	for i, m := range input.Meanings {
		meaningsStr.WriteString(fmt.Sprintf("%d. %s\n", i+1, m.Definition))
	}

	return fmt.Sprintf(`Word: %s
Part of Speech: %s
Meanings:
%s
Target Language: %s

Instruction:
Generate the most useful target-language translations covering the provided meanings. Do not attempt to map each translation to a specific sense. Return unique candidates only. Return 1 to 5 concise translation candidates. Do not force additional candidates when fewer are appropriate. Output ONLY strict JSON array of strings. No markdown, no explanation.

Example output:
["translation1", "translation2"]`, input.Word, input.PartOfSpeech, meaningsStr.String(), input.TargetLanguage)
}

func extractJSONCandidates(response string) ([]string, error) {
	resp := strings.TrimSpace(response)

	// Remove <think>...</think> block if present
	if startThink := strings.Index(resp, "<think>"); startThink != -1 {
		if endThink := strings.Index(resp, "</think>"); endThink != -1 {
			resp = resp[:startThink] + resp[endThink+8:]
			resp = strings.TrimSpace(resp)
		}
	}

	if strings.HasPrefix(resp, "```json") {
		resp = strings.TrimPrefix(resp, "```json")
		if strings.HasSuffix(strings.TrimSpace(resp), "```") {
			resp = strings.TrimSuffix(strings.TrimSpace(resp), "```")
		}
	} else if strings.HasPrefix(resp, "```") {
		resp = strings.TrimPrefix(resp, "```")
		if strings.HasSuffix(strings.TrimSpace(resp), "```") {
			resp = strings.TrimSuffix(strings.TrimSpace(resp), "```")
		}
	}

	startIdx := strings.Index(resp, "[")
	endIdx := strings.LastIndex(resp, "]")
	if startIdx != -1 && endIdx != -1 && endIdx > startIdx {
		resp = resp[startIdx : endIdx+1]
	}

	var rawCandidates []string
	if err := json.Unmarshal([]byte(resp), &rawCandidates); err != nil {
		return nil, fmt.Errorf("invalid json format: %v", err)
	}

	var candidates []string
	seen := make(map[string]bool)

	for _, c := range rawCandidates {
		trimmed := strings.TrimSpace(c)
		if trimmed == "" {
			continue
		}

		lowerTrimmed := strings.ToLower(trimmed)
		if !seen[lowerTrimmed] {
			seen[lowerTrimmed] = true
			candidates = append(candidates, trimmed)
			if len(candidates) == 5 {
				break
			}
		}
	}

	return candidates, nil
}

func (s *service) EnqueueJobs(ctx context.Context, wordIDs []string, targetLang string) error {
	if !IsSupportedLanguage(targetLang) {
		return fmt.Errorf("unsupported target language: %s", targetLang)
	}

	for _, idStr := range wordIDs {
		wID, err := bson.ObjectIDFromHex(idStr)
		if err != nil {
			continue // skip invalid
		}
		// Enqueue job one by one for now (or implement EnqueueMany if needed)
		// We'll just call EnqueueJob on repo
		_ = s.repo.EnqueueJob(ctx, wID, targetLang)
	}

	return nil
}

func (s *service) TranslateBatch(ctx context.Context, input TranslateBatchInput) ([]WordTranslation, error) {
	if s.aiClient == nil {
		return nil, fmt.Errorf("AI client is not configured")
	}
	if len(input.Words) == 0 {
		return nil, nil
	}
	if !IsSupportedLanguage(input.TargetLanguage) {
		return nil, fmt.Errorf("unsupported target language: %s", input.TargetLanguage)
	}

	prompt := s.buildBatchPrompt(input)
	resp, err := s.aiClient.GenerateText(ctx, prompt)
	if err != nil {
		return nil, fmt.Errorf("ai generate failed: %w", err)
	}
	if strings.TrimSpace(resp) == "" {
		return nil, fmt.Errorf("ai returned empty response")
	}

	batchResult, err := extractJSONBatchCandidates(resp)
	if err != nil {
		return nil, fmt.Errorf("failed to extract batch candidates from ai response: %w", err)
	}

	var translations []WordTranslation
	now := time.Now()

	// Match results back to input by word ID and sense
	for _, w := range input.Words {
		sensesMap := batchResult[w.WordID]
		if sensesMap == nil {
			continue
		}

		wID, err := bson.ObjectIDFromHex(w.WordID)
		if err != nil {
			continue
		}

		for _, m := range w.Meanings {
			candidates := sensesMap[m.SenseID]
			if len(candidates) > 0 {
				if len(candidates) > 5 {
					candidates = candidates[:5]
				}
				translations = append(translations, WordTranslation{
					WordID:         wID,
					SenseID:        m.SenseID,
					SourceLanguage: "en",
					TargetLanguage: input.TargetLanguage,
					Translations:   candidates,
					Provider:       "ai",
					Model:          "batch",
					Version:        1,
					CreatedAt:      now,
					UpdatedAt:      now,
				})
			}
		}
	}

	return translations, nil
}

func (s *service) buildBatchPrompt(input TranslateBatchInput) string {
	var wordsStr strings.Builder
	for i, w := range input.Words {
		wordsStr.WriteString(fmt.Sprintf("%d. Word ID: %s (Word: %s)\n   POS: %s\n   Meanings:\n", i+1, w.WordID, w.Word, w.PartOfSpeech))
		for _, m := range w.Meanings {
			wordsStr.WriteString(fmt.Sprintf("     [Sense %s] %s\n", m.SenseID, m.Definition))
		}
		wordsStr.WriteString("\n")
	}

	return fmt.Sprintf(`Target Language: %s

Please translate the following English words into %s based on their parts of speech and specific meanings (senses).
Instruction:
Generate the most useful target-language translations for EACH specific sense provided. Return unique candidates only. Return 1 to 5 concise translation candidates per sense.
Output ONLY a strict JSON object mapping the Word ID (as a string) to an object that maps the Sense ID (as a string) to an array of string translations. No markdown, no explanation.

Example output:
{
  "6a462b2d96e73b4df4752ff8": {
    "sense_uuid_1": ["nam châm", "đá nam châm"],
    "sense_uuid_2": ["sức hút", "người cuốn hút"]
  },
  "6a462b2d96e73b4df4752ff9": {
    "sense_uuid_3": ["tài khoản ngân hàng"]
  }
}

Words to translate:
%s`, input.TargetLanguage, input.TargetLanguage, wordsStr.String())
}

func extractJSONBatchCandidates(response string) (map[string]map[string][]string, error) {
	resp := strings.TrimSpace(response)

	// Remove <think>...</think> block if present
	if startThink := strings.Index(resp, "<think>"); startThink != -1 {
		if endThink := strings.Index(resp, "</think>"); endThink != -1 {
			resp = resp[:startThink] + resp[endThink+8:]
			resp = strings.TrimSpace(resp)
		}
	}

	if strings.HasPrefix(resp, "```json") {
		resp = strings.TrimPrefix(resp, "```json")
		if strings.HasSuffix(strings.TrimSpace(resp), "```") {
			resp = strings.TrimSuffix(strings.TrimSpace(resp), "```")
		}
	} else if strings.HasPrefix(resp, "```") {
		resp = strings.TrimPrefix(resp, "```")
		if strings.HasSuffix(strings.TrimSpace(resp), "```") {
			resp = strings.TrimSuffix(strings.TrimSpace(resp), "```")
		}
	}

	startIdx := strings.Index(resp, "{")
	endIdx := strings.LastIndex(resp, "}")
	if startIdx != -1 && endIdx != -1 && endIdx > startIdx {
		resp = resp[startIdx : endIdx+1]
	}

	var rawMap map[string]map[string][]string
	if err := json.Unmarshal([]byte(resp), &rawMap); err != nil {
		return nil, fmt.Errorf("invalid json format for batch: %v\nResp: %s", err, resp)
	}

	result := make(map[string]map[string][]string)
	for wordKey, sensesMap := range rawMap {
		normalizedKey := strings.ToLower(strings.TrimSpace(wordKey))
		senseResult := make(map[string][]string)

		for senseKey, rawCandidates := range sensesMap {
			var candidates []string
			seen := make(map[string]bool)
			for _, c := range rawCandidates {
				trimmed := strings.TrimSpace(c)
				if trimmed == "" {
					continue
				}
				lowerTrimmed := strings.ToLower(trimmed)
				if !seen[lowerTrimmed] {
					seen[lowerTrimmed] = true
					candidates = append(candidates, trimmed)
				}
			}
			senseResult[senseKey] = candidates
		}
		result[normalizedKey] = senseResult
	}

	return result, nil
}

func (s *service) GetTranslationsByWordIDs(ctx context.Context, wordIDs []bson.ObjectID, targetLang string) ([]WordTranslation, error) {
	return s.repo.GetTranslationsByWordIDs(ctx, wordIDs, targetLang)
}
