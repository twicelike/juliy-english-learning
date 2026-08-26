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
func (s *exerciseContentService) GetExercises(ctx context.Context, refs []SessionItemRef, exType ExerciseType, limitPerWord int) (map[string][]ExerciseContent, error) {
	// 1. Fetch from DB pool
	pool, err := s.repo.GetRandomExercises(ctx, refs, exType, limitPerWord)
	if err != nil {
		return nil, err
	}

	if pool == nil {
		pool = make(map[string][]ExerciseContent)
	}

	// 2. Identify words that need more exercises
	var missingRefs []SessionItemRef
	for _, ref := range refs {
		key := ref.WordID.Hex() + ":" + ref.SenseID
		if len(pool[key]) < limitPerWord {
			missingRefs = append(missingRefs, ref)
		}
	}

	// 3. Generate missing exercises via AI
	if len(missingRefs) > 0 && s.aiClient != nil {
		err := s.generateAndSaveMissing(ctx, missingRefs, pool, exType, limitPerWord)
		if err != nil {
			// Log error, but we still return whatever we have in the pool so the user can continue learning
		}
	}

	return pool, nil
}

func (s *exerciseContentService) generateAndSaveMissing(ctx context.Context, missingRefs []SessionItemRef, pool map[string][]ExerciseContent, exType ExerciseType, limitPerWord int) error {
	maxRetries := 3

	// Resolve words from dictionary for the prompt
	var missingWordIDs []bson.ObjectID
	for _, ref := range missingRefs {
		missingWordIDs = append(missingWordIDs, ref.WordID)
	}

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
		var currentBatchRefs []SessionItemRef

		for _, ref := range missingRefs {
			key := ref.WordID.Hex() + ":" + ref.SenseID
			needed := limitPerWord - len(pool[key])
			if needed > 0 {
				doc, ok := wordMap[ref.WordID]
				if !ok {
					continue
				}

				// Find exact sense
				var targetSense *dictionary.SenseDoc
				for i := range doc.Senses {
					if doc.Senses[i].ID == ref.SenseID {
						targetSense = &doc.Senses[i]
						break
					}
				}

				if targetSense == nil {
					// DO NOT fallback to Senses[0].
					// Skip generating exercise for this invalid SenseID.
					continue
				}

				// Build context
				def := ""
				if len(targetSense.Glosses) > 0 {
					def = targetSense.Glosses[0]
				}

				// Gather existing sentences to prevent duplicates
				var existing []string
				for _, ex := range pool[key] {
					if ex.Sentence != "" {
						existing = append(existing, ex.Sentence)
					} else if ex.Question != "" {
						existing = append(existing, ex.Question)
					}
				}

				for i := 0; i < needed; i++ {
					reqID := bson.NewObjectID().Hex() // unique ID to match AI response
					contexts = append(contexts, WordGenerationContext{
						ReqID:             reqID,
						WordID:            ref.WordID,
						SenseID:           ref.SenseID,
						Word:              doc.Word,
						POS:               doc.POS,
						Definition:        def,
						ExistingSentences: existing,
					})
					currentBatchRefs = append(currentBatchRefs, ref)
				}
			}
		}

		if len(contexts) == 0 {
			break // all done
		}

		// Create map of req_id to contexts for quick lookup
		reqMap := make(map[string]WordGenerationContext)
		for _, c := range contexts {
			reqMap[c.ReqID] = c
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
				ctxData, ok := reqMap[ex.ReqID]
				if !ok {
					continue
				}

				hash := generateContentHash(ex.Question)
				content := &ExerciseContent{
					WordID:        ctxData.WordID,
					SenseID:       ctxData.SenseID,
					Type:          exType,
					Question:      ex.Question,
					Answer:        ex.Answer,
					ContentHash:   hash,
					PromptVersion: 2, // Sense-aware Version
				}

				if err := s.repo.Save(ctx, content); err != nil {
					continue // Likely Duplicate Key Error
				}
				key := ctxData.WordID.Hex() + ":" + ctxData.SenseID
				pool[key] = append(pool[key], *content)
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
				ctxData, ok := reqMap[ex.ReqID]
				if !ok {
					continue
				}

				// Generate hash based on choices + word to ensure unique question
				hashData := strings.Join(ex.Choices, "|") + "|" + ex.Answer
				hash := generateContentHash(hashData)
				content := &ExerciseContent{
					WordID:        ctxData.WordID,
					SenseID:       ctxData.SenseID,
					Type:          exType,
					Choices:       ex.Choices,
					Answer:        ex.Answer,
					ContentHash:   hash,
					PromptVersion: 2, // Sense-aware Version
				}

				if err := s.repo.Save(ctx, content); err != nil {
					continue // Likely Duplicate Key Error
				}
				key := ctxData.WordID.Hex() + ":" + ctxData.SenseID
				pool[key] = append(pool[key], *content)
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
