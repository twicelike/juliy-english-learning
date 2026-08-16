package speaking

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"time"

	"github.com/english/backend/internal/ai"
	"github.com/english/backend/internal/storage"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type service struct {
	repo      Repository
	aiService ai.AITutorService
	storage   storage.StorageService // Let's use an interface, but we need to define it or import it. We created it in internal/storage.
	processor ProcessingService
}

// NewService creates a new speaking service.
func NewService(repo Repository, aiService ai.AITutorService, storage storage.StorageService, processor ProcessingService) Service {
	return &service{
		repo:      repo,
		aiService: aiService,
		storage:   storage,
		processor: processor,
	}
}

func (s *service) GetRandomTask(ctx context.Context, topic, difficulty, taskType, focus string) (*SpeakingTask, error) {
	// 1. Try to find an existing task in DB
	task, err := s.repo.GetRandomTask(ctx, topic, difficulty, taskType, focus)
	if err != nil {
		return nil, err
	}
	if task != nil {
		return task, nil
	}

	// 2. Fallback to AI Generation
	if difficulty == "" {
		difficulty = "medium" // default difficulty
	}
	
	resp, err := s.aiService.GenerateSpeakingTask(ctx, topic, taskType, difficulty)
	if err != nil {
		return nil, err
	}

	// 3. Map to Domain Model
	newTask := &SpeakingTask{
		ID:                  bson.NewObjectID(),
		Topic:               topic,
		Type:                taskType,
		Difficulty:          difficulty,
		Prompt:              resp.Prompt,
		Scenario:            resp.Scenario,
		Instructions:        resp.Instructions,
		Hints:               resp.Hints,
		SuggestedVocabulary: resp.SuggestedVocabulary,
		DurationSeconds:     resp.DurationSeconds,
		CreatedAt:           time.Now(),
		UpdatedAt:           time.Now(),
	}

	// 4. Save to DB to reuse later
	if err := s.repo.CreateTask(ctx, newTask); err != nil {
		// Log error but we can still return the task to the user
		// In production, we'd log this properly
	}

	return newTask, nil
}

func (s *service) StartSession(ctx context.Context, userID string, taskID string) (*SpeakingSession, error) {
	uid, err := bson.ObjectIDFromHex(userID)
	if err != nil {
		return nil, err
	}
	tid, err := bson.ObjectIDFromHex(taskID)
	if err != nil {
		return nil, err
	}

	// 1. Fetch the Task
	task, err := s.repo.GetTaskByID(ctx, tid)
	if err != nil {
		return nil, err
	}
	if task == nil {
		return nil, fmt.Errorf("task not found")
	}

	// 2. Build the Snapshot
	snapshot := SpeakingTaskSnapshot{
		Topic:               task.Topic,
		TaskType:            task.Type,
		Difficulty:          task.Difficulty,
		Prompt:              task.Prompt,
		Scenario:            task.Scenario,
		Instructions:        task.Instructions,
		Hints:               task.Hints,
		SuggestedVocabulary: task.SuggestedVocabulary,
		DurationSeconds:     task.DurationSeconds,
	}

	session := &SpeakingSession{
		ID:           bson.NewObjectID(),
		UserID:       uid,
		TaskID:       tid,
		TaskSnapshot: snapshot,
		Status:       "in_progress",
		StartedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	err = s.repo.CreateSession(ctx, session)
	if err != nil {
		return nil, err
	}

	return session, nil
}

func (s *service) GetSessionDetails(ctx context.Context, userID string, sessionID string) (*SpeakingSession, []SpeakingAttempt, error) {
	uid, err := bson.ObjectIDFromHex(userID)
	if err != nil {
		return nil, nil, err
	}
	sid, err := bson.ObjectIDFromHex(sessionID)
	if err != nil {
		return nil, nil, err
	}

	session, err := s.repo.GetSessionByID(ctx, uid, sid)
	if err != nil {
		return nil, nil, err
	}

	attempts, err := s.repo.GetAttemptsBySession(ctx, uid, sid)
	if err != nil {
		return nil, nil, err
	}

	return session, attempts, nil
}

func (s *service) GetHistory(ctx context.Context, userID string, limit int) ([]SpeakingHistoryItem, error) {
	uid, err := bson.ObjectIDFromHex(userID)
	if err != nil {
		return nil, err
	}
	return s.repo.GetHistory(ctx, uid, limit)
}

func (s *service) SubmitAttempt(ctx context.Context, userID string, sessionID string, mimeType string, durationMs int, audioBytes []byte) (*SpeakingAttempt, error) {
	uid, err := bson.ObjectIDFromHex(userID)
	if err != nil {
		return nil, err
	}
	sid, err := bson.ObjectIDFromHex(sessionID)
	if err != nil {
		return nil, err
	}

	session, err := s.repo.GetSessionByID(ctx, uid, sid)
	if err != nil || session == nil {
		return nil, fmt.Errorf("session not found")
	}

	// 1. Generate Attempt ID
	attemptID := bson.NewObjectID()

	// 2. Save audio to storage
	filename := fmt.Sprintf("%s.webm", attemptID.Hex()) // Assuming webm for now, or could check mimeType
	storedFile, err := s.storage.Save(ctx, "speaking", filename, mimeType, bytes.NewReader(audioBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to save audio: %w", err)
	}

	// 3. Create Attempt in DB with "processing" status
	now := time.Now()
	attempt := &SpeakingAttempt{
		ID:                  attemptID,
		SessionID:           sid,
		UserID:              uid,
		TaskSnapshot:        session.TaskSnapshot,
		AudioKey:            storedFile.Key,
		AudioMimeType:       mimeType,
		AudioDuration:       durationMs,
		Status:              AttemptProcessing,
		CreatedAt:           now,
		ProcessingStartedAt: &now,
	}

	if err := s.repo.CreateAttempt(ctx, attempt); err != nil {
		return nil, err
	}

	// 4. Trigger background processing
	// Fire and forget using a context that won't be cancelled when HTTP request ends.
	go s.processor.ProcessAttempt(context.Background(), attempt.ID.Hex())

	return attempt, nil
}

func (s *service) GetAttemptByID(ctx context.Context, userID string, attemptID string) (*SpeakingAttempt, error) {
	uid, err := bson.ObjectIDFromHex(userID)
	if err != nil {
		return nil, err
	}
	aid, err := bson.ObjectIDFromHex(attemptID)
	if err != nil {
		return nil, err
	}

	attempt, err := s.repo.GetAttemptByID(ctx, uid, aid)
	if err != nil {
		return nil, err
	}
	if attempt == nil {
		return nil, fmt.Errorf("attempt not found")
	}

	return attempt, nil
}

func (s *service) GetAttemptAudio(ctx context.Context, userID string, attemptID string) (io.ReadCloser, int64, string, error) {
	attempt, err := s.GetAttemptByID(ctx, userID, attemptID)
	if err != nil {
		return nil, 0, "", err
	}

	reader, storedFile, err := s.storage.Open(ctx, attempt.AudioKey)
	if err != nil {
		return nil, 0, "", fmt.Errorf("failed to open audio file: %w", err)
	}

	return reader, storedFile.Size, attempt.AudioMimeType, nil
}
