package speaking

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"time"

	"github.com/english/backend/internal/ai"
	"github.com/english/backend/internal/dictionary"
	"github.com/english/backend/internal/evaluation"
	"github.com/english/backend/internal/events"
	"github.com/english/backend/internal/gamification"
	"github.com/english/backend/internal/storage"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type processingService struct {
	repo      Repository
	aiService ai.AITutorService
	storage   storage.StorageService
	dictRepo  dictionary.Repository
	gamiSvc   gamification.Service
	evalSvc   evaluation.Service
	publisher events.Publisher
}

func NewProcessingService(repo Repository, aiService ai.AITutorService, storage storage.StorageService, dictRepo dictionary.Repository, gamiSvc gamification.Service, evalSvc evaluation.Service, publisher events.Publisher) ProcessingService {
	return &processingService{
		repo:      repo,
		aiService: aiService,
		storage:   storage,
		dictRepo:  dictRepo,
		gamiSvc:   gamiSvc,
		evalSvc:   evalSvc,
		publisher: publisher,
	}
}

func (s *processingService) ProcessAttempt(ctx context.Context, attemptIDStr string) error {
	attemptID, err := bson.ObjectIDFromHex(attemptIDStr)
	if err != nil {
		slog.Error("ProcessAttempt invalid attempt ID", "error", err, "attemptID", attemptIDStr)
		return err
	}

	// 1. Fetch attempt (we don't have GetAttemptByID without userID, but we can add one to repo or we can just fetch it.
	// Wait, the repo has `GetAttemptByID(ctx context.Context, userID, attemptID bson.ObjectID)`. We might need to bypass userID or find it.
	// We can update Repository to add GetAttempt(ctx, attemptID). Let's assume we can fetch it, or we need to add it.)
	// For now, let's assume we will add `GetAttempt(ctx context.Context, attemptID bson.ObjectID) (*SpeakingAttempt, error)` to Repository.

	attempt, err := s.repo.GetAttempt(ctx, attemptID)
	if err != nil {
		slog.Error("ProcessAttempt failed to fetch attempt", "error", err, "attemptID", attemptIDStr)
		return err
	}
	if attempt == nil {
		slog.Error("ProcessAttempt attempt not found", "attemptID", attemptIDStr)
		return fmt.Errorf("attempt not found")
	}

	if attempt.Status != AttemptProcessing {
		slog.Info("ProcessAttempt skipped, attempt not in processing state", "attemptID", attemptIDStr, "status", attempt.Status)
		return nil
	}

	// Helper to mark attempt as failed
	markFailed := func(code, msg string) {
		now := time.Now()
		attempt.Status = AttemptFailed
		attempt.ErrorCode = code
		attempt.ErrorMessage = msg
		attempt.CompletedAt = &now
		if err := s.repo.UpdateAttempt(ctx, attempt); err != nil {
			slog.Error("ProcessAttempt failed to update attempt as failed", "error", err, "attemptID", attemptIDStr)
		}
	}

	// 2. Fetch audio from storage
	audioStream, _, err := s.storage.Open(ctx, attempt.AudioKey)
	if err != nil {
		markFailed("STORAGE_ERROR", "Failed to retrieve audio file")
		return err
	}
	defer audioStream.Close()

	audioBytes, err := io.ReadAll(audioStream)
	if err != nil {
		markFailed("STORAGE_READ_ERROR", "Failed to read audio file")
		return err
	}

	// 3. Call AI Evaluation
	resp, err := s.aiService.EvaluateSpeaking(ctx, audioBytes, attempt.AudioMimeType, attempt.TaskSnapshot)
	if err != nil {
		markFailed("AI_PROCESSING_FAILED", "Unable to process your recording.")
		return err
	}

	// 4. Update attempt with success
	now := time.Now()
	attempt.Status = AttemptCompleted
	attempt.Transcript = resp.Transcript

	newUsedWords := make([]NewUsedWord, 0, len(resp.NewUsedVocabulary))
	for _, vocab := range resp.NewUsedVocabulary {
		wordID := ""
		// Look up the word in dictionary to find its ID
		docs, err := s.dictRepo.GetWords(ctx, vocab.Word)
		if err == nil && len(docs) > 0 {
			wordID = docs[0].ID.Hex()
		}
		newUsedWords = append(newUsedWords, NewUsedWord{
			Word:   vocab.Word,
			WordID: wordID,
			Reason: vocab.Reason,
		})
	}

	attempt.Evaluation = &SpeakingEvaluation{
		Overall:           resp.Score,
		TaskCompletion:    resp.TaskCompletion.Score,
		Fluency:           resp.Fluency.Score,
		Grammar:           resp.Grammar.Score,
		Vocabulary:        resp.Vocabulary.Score,
		Pronunciation:     resp.Pronunciation.Score,
		Strengths:         resp.Strengths,
		Improvements:      resp.Improvements,
		NewUsedVocabulary: newUsedWords,
	}
	attempt.CompletedAt = &now

	if err := s.repo.UpdateAttempt(ctx, attempt); err != nil {
		slog.Error("ProcessAttempt failed to update attempt on success", "error", err, "attemptID", attemptIDStr)
		return err
	}

	// Track in Gamification
	if s.gamiSvc != nil {
		if err := s.gamiSvc.TrackSpeakingSession(ctx, attempt.UserID.Hex(), attempt.Evaluation.Overall); err != nil {
			slog.Error("ProcessAttempt failed to track gamification", "error", err, "attemptID", attemptIDStr)
			// Do not return error here so that speaking evaluation remains successful
		}
	}

	// Publish Evidence to Evaluation Engine
	if s.evalSvc != nil {
		ev := &evaluation.Evidence{
			UserID:   attempt.UserID,
			EventID:  attempt.ID.Hex(),
			Skill:    "speaking",
			Score:    float64(attempt.Evaluation.Overall),
			Weight:   evaluation.WeightSpeaking,
			Source:   evaluation.SourceSpeakingLab,
			Metadata: bson.M{"session_id": attempt.SessionID.Hex()},
		}

		// Run in background to avoid blocking
		go func(e *evaluation.Evidence, uid string) {
			bgCtx := context.Background()
			if err := s.evalSvc.RecordEvidence(bgCtx, e); err != nil {
				slog.Error("ProcessAttempt failed to record evaluation evidence", "error", err, "attemptID", attemptIDStr)
			}

			if s.publisher != nil {
				_ = s.publisher.Publish(bgCtx, events.Event{
					Type:   events.EventTypeProfileRefreshRequested,
					UserID: uid,
					Reason: "speaking_session_completed",
				})
			}
		}(ev, attempt.UserID.Hex())
	}

	slog.Info("ProcessAttempt completed successfully", "attemptID", attemptIDStr, "score", resp.Score)
	return nil
}
