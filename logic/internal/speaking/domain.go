package speaking

import (
	"context"
	"io"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// SpeakingTask represents a speaking practice task.
type SpeakingTask struct {
	ID                  bson.ObjectID `bson:"_id,omitempty" json:"id"`
	Topic               string        `bson:"topic" json:"topic"`
	Type                string        `bson:"type" json:"type"`             // "guided", "scenario", "free_talk"
	Difficulty          string        `bson:"difficulty" json:"difficulty"` // "easy", "medium", "hard"
	Title               string        `bson:"title" json:"title"`
	Prompt              string        `bson:"prompt" json:"prompt"`
	Scenario            string        `bson:"scenario,omitempty" json:"scenario,omitempty"`
	Instructions        []string      `bson:"instructions,omitempty" json:"instructions,omitempty"`
	Hints               []string      `bson:"hints,omitempty" json:"hints,omitempty"`
	SuggestedVocabulary []string      `bson:"suggested_vocabulary,omitempty" json:"suggested_vocabulary,omitempty"`
	Focus               string        `bson:"focus,omitempty" json:"focus,omitempty"`
	DurationSeconds     int           `bson:"duration_seconds" json:"duration_seconds"`
	CreatedAt           time.Time     `bson:"created_at" json:"created_at"`
	UpdatedAt           time.Time     `bson:"updated_at" json:"updated_at"`
}

// SpeakingTaskSnapshot stores a frozen copy of the task at the time the session started.
type SpeakingTaskSnapshot struct {
	Topic               string   `bson:"topic" json:"topic"`
	TaskType            string   `bson:"task_type" json:"task_type"`
	Difficulty          string   `bson:"difficulty" json:"difficulty"`
	Prompt              string   `bson:"prompt" json:"prompt"`
	Scenario            string   `bson:"scenario,omitempty" json:"scenario,omitempty"`
	Instructions        []string `bson:"instructions" json:"instructions"`
	Hints               []string `bson:"hints,omitempty" json:"hints,omitempty"`
	SuggestedVocabulary []string `bson:"suggested_vocabulary,omitempty" json:"suggested_vocabulary,omitempty"`
	Focus               string   `bson:"focus,omitempty" json:"focus,omitempty"`
	DurationSeconds     int      `bson:"duration_seconds" json:"duration_seconds"`
}

// SpeakingSession represents a user's practice session.
type SpeakingSession struct {
	ID           bson.ObjectID        `bson:"_id,omitempty" json:"id"`
	UserID       bson.ObjectID        `bson:"user_id" json:"user_id"`
	TaskID       bson.ObjectID        `bson:"task_id" json:"task_id"`
	TaskSnapshot SpeakingTaskSnapshot `bson:"task_snapshot" json:"task_snapshot"`
	Status       string               `bson:"status" json:"status"` // "in_progress", "completed"
	BestScore    int                  `bson:"best_score" json:"best_score"`
	AttemptCount int                  `bson:"attempt_count" json:"attempt_count"`
	StartedAt    time.Time            `bson:"started_at" json:"started_at"`
	CompletedAt  *time.Time           `bson:"completed_at,omitempty" json:"completed_at,omitempty"`
	UpdatedAt    time.Time            `bson:"updated_at" json:"updated_at"`
}

// NewUsedWord represents a vocabulary word the user used naturally.
type NewUsedWord struct {
	Word   string `bson:"word" json:"word"`
	WordID string `bson:"word_id,omitempty" json:"word_id,omitempty"`
	Reason string `bson:"reason" json:"reason"`
}

// SpeakingEvaluation contains the AI's assessment of the user's speaking.
type SpeakingEvaluation struct {
	Overall           int           `bson:"overall" json:"overall"`
	TaskCompletion    int           `bson:"task_completion" json:"task_completion"`
	Fluency           int           `bson:"fluency" json:"fluency"`
	Grammar           int           `bson:"grammar" json:"grammar"`
	Vocabulary        int           `bson:"vocabulary" json:"vocabulary"`
	Pronunciation     int           `bson:"pronunciation" json:"pronunciation"`
	Strengths         []string      `bson:"strengths" json:"strengths"`
	Improvements      []string      `bson:"improvements" json:"improvements"`
	NewUsedVocabulary []NewUsedWord `bson:"new_used_vocabulary,omitempty" json:"new_used_vocabulary,omitempty"`
}

// AttemptStatus represents the current processing state of a speaking attempt.
type AttemptStatus string

const (
	AttemptProcessing AttemptStatus = "processing"
	AttemptCompleted  AttemptStatus = "completed"
	AttemptFailed     AttemptStatus = "failed"
)

// SpeakingAttempt represents a single recording attempt within a session.
type SpeakingAttempt struct {
	ID                  bson.ObjectID        `bson:"_id,omitempty" json:"id"`
	SessionID           bson.ObjectID        `bson:"session_id" json:"session_id"`
	UserID              bson.ObjectID        `bson:"user_id" json:"user_id"`
	TaskSnapshot        SpeakingTaskSnapshot `bson:"task_snapshot" json:"task_snapshot"`
	AudioKey            string               `bson:"audio_key" json:"audio_key"`
	AudioMimeType       string               `bson:"audio_mime_type" json:"audio_mime_type"`
	AudioDuration       int                  `bson:"audio_duration_ms" json:"audio_duration_ms"`
	Status              AttemptStatus        `bson:"status" json:"status"`
	Transcript          string               `bson:"transcript,omitempty" json:"transcript,omitempty"`
	Evaluation          *SpeakingEvaluation  `bson:"evaluation,omitempty" json:"evaluation,omitempty"`
	ErrorCode           string               `bson:"error_code,omitempty" json:"error_code,omitempty"`
	ErrorMessage        string               `bson:"error_message,omitempty" json:"error_message,omitempty"`
	CreatedAt           time.Time            `bson:"created_at" json:"created_at"`
	ProcessingStartedAt *time.Time           `bson:"processing_started_at,omitempty" json:"processing_started_at,omitempty"`
	CompletedAt         *time.Time           `bson:"completed_at,omitempty" json:"completed_at,omitempty"`
}

// SpeakingHistoryItem represents a session and its best score for history display.
type SpeakingHistoryItem struct {
	Session       SpeakingSession `bson:"session" json:"session"`
	BestScore     *int            `bson:"best_score" json:"best_score"`
	AttemptsCount int             `bson:"attempts_count" json:"attempts_count"`
}

// -----------------------------------------------------------------------------
// Interfaces
// -----------------------------------------------------------------------------

// Repository defines the data access methods for the Speaking domain.
type Repository interface {
	// Task
	GetTaskByID(ctx context.Context, id bson.ObjectID) (*SpeakingTask, error)
	GetRandomTask(ctx context.Context, topic, difficulty, taskType, focus string) (*SpeakingTask, error)
	CreateTask(ctx context.Context, task *SpeakingTask) error
	UpdateTask(ctx context.Context, task *SpeakingTask) error

	// Session
	CreateSession(ctx context.Context, session *SpeakingSession) error
	UpdateSession(ctx context.Context, session *SpeakingSession) error
	GetSessionByID(ctx context.Context, userID, sessionID bson.ObjectID) (*SpeakingSession, error)
	GetHistory(ctx context.Context, userID bson.ObjectID, limit int) ([]SpeakingHistoryItem, error)

	// Attempt
	CreateAttempt(ctx context.Context, attempt *SpeakingAttempt) error
	UpdateAttempt(ctx context.Context, attempt *SpeakingAttempt) error
	GetAttempt(ctx context.Context, attemptID bson.ObjectID) (*SpeakingAttempt, error)
	GetAttemptByID(ctx context.Context, userID, attemptID bson.ObjectID) (*SpeakingAttempt, error)
	GetAttemptsBySession(ctx context.Context, userID, sessionID bson.ObjectID) ([]SpeakingAttempt, error)
	GetRecentCompletedAttempts(ctx context.Context, userID bson.ObjectID, limit int) ([]SpeakingAttempt, error)
}

// Service defines the business logic methods for the Speaking domain.
type Service interface {
	GetRandomTask(ctx context.Context, topic, difficulty, taskType, focus string) (*SpeakingTask, error)
	StartSession(ctx context.Context, userID string, taskID string) (*SpeakingSession, error)
	GetSessionDetails(ctx context.Context, userID string, sessionID string) (*SpeakingSession, []SpeakingAttempt, error)
	GetHistory(ctx context.Context, userID string, limit int) ([]SpeakingHistoryItem, error)
	SubmitAttempt(ctx context.Context, userID string, sessionID string, mimeType string, durationMs int, audioBytes []byte) (*SpeakingAttempt, error)
	GetAttemptByID(ctx context.Context, userID string, attemptID string) (*SpeakingAttempt, error)
	GetAttemptAudio(ctx context.Context, userID string, attemptID string) (io.ReadCloser, int64, string, error)
}

// ProcessingService defines the background job operations for processing speaking attempts.
type ProcessingService interface {
	ProcessAttempt(ctx context.Context, attemptID string) error
}

type RecommendationReason struct {
	Skill           string `json:"skill"`
	Focus           string `json:"focus"`
	BasedOnSessions int    `json:"based_on_sessions"`
}

type RecommendedTaskResponse struct {
	Task   *SpeakingTask         `json:"task"`
	Reason *RecommendationReason `json:"reason,omitempty"`
}

// RecommendationService defines logic for adaptive speaking.
type RecommendationService interface {
	GetRecommendedTask(ctx context.Context, userID string) (*RecommendedTaskResponse, error)
}
