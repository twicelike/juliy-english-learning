package learning

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"strings"

	"github.com/english/backend/internal/ai"
	"github.com/english/backend/internal/dictionary"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type exerciseContentService struct {
	repo     ExerciseContentRepository
	dictRepo dictionary.Repository
	aiClient ai.AIClient
}

func NewExerciseContentService(repo ExerciseContentRepository, dictRepo dictionary.Repository, aiClient ai.AIClient) ExerciseContentService {
	return &exerciseContentService{
		repo:     repo,
		dictRepo: dictRepo,
		aiClient: aiClient,
	}
}

// GetExercises returns exercises for the given words. 
// If the DB pool is insufficient, it uses the AI client to generate and save more.
func (s *exerciseContentService) GetExercises(ctx context.Context, wordIDs []bson.ObjectID, exType ExerciseType, limitPerWord int) (map[bson.ObjectID][]ExerciseContent, error) {
	// 1. Fetch from DB pool
	pool, err := s.repo.GetRandomExercises(ctx, wordIDs, exType, limitPerWord)
	if err != nil {
		return nil, err
	}

	if pool == nil {
		pool = make(map[bson.ObjectID][]ExerciseContent)
	}

	// 2. Identify words that need more exercises
	var missingWordIDs []bson.ObjectID
	for _, wid := range wordIDs {
		if len(pool[wid]) < limitPerWord {
			missingWordIDs = append(missingWordIDs, wid)
		}
	}

	// 3. Generate missing exercises via AI
	if len(missingWordIDs) > 0 && s.aiClient != nil {
		// Note: We'll implement batch AI generation shortly
		err := s.generateAndSaveMissing(ctx, missingWordIDs, pool, exType, limitPerWord)
		if err != nil {
			// Log error, but we still return whatever we have in the pool so the user can continue learning
		}
	}

	return pool, nil
}

func (s *exerciseContentService) generateAndSaveMissing(ctx context.Context, missingWordIDs []bson.ObjectID, pool map[bson.ObjectID][]ExerciseContent, exType ExerciseType, limitPerWord int) error {
	maxRetries := 3

	// Resolve words from dictionary for the prompt
	wordDocs, err := s.dictRepo.GetManyByIDs(ctx, missingWordIDs)
	if err != nil {
		return err
	}
	wordMap := make(map[bson.ObjectID]*dictionary.WordDocument)
	for i := range wordDocs {
		wordMap[wordDocs[i].ID] = &wordDocs[i]
	}

	for retry := 0; retry < maxRetries; retry++ {
		var contexts []WordGenerationContext
		var currentBatchIDs []bson.ObjectID

		for _, wid := range missingWordIDs {
			needed := limitPerWord - len(pool[wid])
			if needed > 0 {
				doc, ok := wordMap[wid]
				if !ok {
					continue
				}

				// Build context
				def := ""
				if len(doc.Senses) > 0 && len(doc.Senses[0].Glosses) > 0 {
					def = doc.Senses[0].Glosses[0]
				}

				// Gather existing sentences to prevent duplicates
				var existing []string
				for _, ex := range pool[wid] {
					if ex.Sentence != "" {
						existing = append(existing, ex.Sentence)
					} else if ex.Question != "" {
						existing = append(existing, ex.Question)
					}
				}

				for i := 0; i < needed; i++ {
					contexts = append(contexts, WordGenerationContext{
						WordID:            wid,
						Word:              doc.Word,
						POS:               doc.POS,
						Definition:        def,
						ExistingSentences: existing,
					})
					currentBatchIDs = append(currentBatchIDs, wid)
				}
			}
		}

		if len(contexts) == 0 {
			break // all done
		}

		if exType == ExerciseFillBlank {
			prompt := buildBatchFillBlankPrompt(contexts)
			var aiResp AIBatchFillBlankResponse
			if err := s.aiClient.GenerateJSON(ctx, prompt, nil, &aiResp); err != nil {
				return err
			}

			for _, ex := range aiResp.Exercises {
				if err := validateFillBlankItem(ex); err != nil {
					continue
				}
				wid, err := bson.ObjectIDFromHex(ex.WordID)
				if err != nil {
					continue
				}

				hash := generateContentHash(ex.Question)
				content := &ExerciseContent{
					WordID:        wid,
					Type:          exType,
					Question:      ex.Question,
					Answer:        ex.Answer,
					ContentHash:   hash,
					PromptVersion: 1, // L4 Version
				}

				if err := s.repo.Save(ctx, content); err != nil {
					continue // Likely Duplicate Key Error
				}
				pool[wid] = append(pool[wid], *content)
			}

		} else if exType == ExerciseMultipleChoice {
			prompt := buildBatchMCQPrompt(contexts)
			var aiResp AIBatchMCQResponse
			if err := s.aiClient.GenerateJSON(ctx, prompt, nil, &aiResp); err != nil {
				return err
			}

			for _, ex := range aiResp.Exercises {
				if err := validateMCQItem(ex); err != nil {
					continue
				}
				wid, err := bson.ObjectIDFromHex(ex.WordID)
				if err != nil {
					continue
				}

				// Generate hash based on choices + word to ensure unique question
				hashData := strings.Join(ex.Choices, "|") + "|" + ex.Answer
				hash := generateContentHash(hashData)
				content := &ExerciseContent{
					WordID:        wid,
					Type:          exType,
					Choices:       ex.Choices,
					Answer:        ex.Answer,
					ContentHash:   hash,
					PromptVersion: 1, // L4 Version
				}

				if err := s.repo.Save(ctx, content); err != nil {
					continue // Likely Duplicate Key Error
				}
				pool[wid] = append(pool[wid], *content)
			}
		}
	}

	return nil
}

func generateContentHash(text string) string {
	text = strings.ToLower(strings.TrimSpace(text))
	hash := sha256.Sum256([]byte(text))
	return hex.EncodeToString(hash[:])
}
