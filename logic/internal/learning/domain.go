package learning

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// -----------------------------------------------------------------------------
// Domain Models (DB Structs)
// -----------------------------------------------------------------------------

// LearningProgress is the single source of truth about a user's learning state
// for a given word. It stores a reference (word_id) to the words collection —
// no word details are duplicated here.
type LearningProgress struct {
	ID             bson.ObjectID `bson:"_id,omitempty"`
	UserID         bson.ObjectID `bson:"user_id"`
	WordID         bson.ObjectID `bson:"word_id"`
	SenseID        string        `bson:"sense_id"`
	Interval       int           `bson:"interval"`    // Interval in days
	Repetition     int           `bson:"repetition"`  // Consecutive correct reviews
	EaseFactor     float64       `bson:"ease_factor"` // SM-2 Ease Factor (default 2.5)
	NextReviewDate time.Time     `bson:"next_review_date"`
	LastReviewDate time.Time     `bson:"last_review_date"` // New field
	CreatedAt      time.Time     `bson:"created_at"`
	UpdatedAt      time.Time     `bson:"updated_at"`
}

// LearningProgressKey uniquely identifies a progress record.
type LearningProgressKey struct {
	WordID  bson.ObjectID
	SenseID string
}

type SessionStatus string

const (
	SessionPreparing SessionStatus = "preparing"
	SessionActive    SessionStatus = "active"
	SessionCompleted SessionStatus = "completed"
	SessionExpired   SessionStatus = "expired"
)

type ReviewStrategy string

const (
	ReviewStrategyContinue ReviewStrategy = "continue_learning"
	ReviewStrategyFresh    ReviewStrategy = "fresh"
	ReviewStrategySmart    ReviewStrategy = "smart"
)

// SessionItemRef pairs a WordID with the specific SenseID the user is learning.
type SessionItemRef struct {
	WordID  bson.ObjectID `bson:"word_id"`
	SenseID string        `bson:"sense_id"`
}

// LearningSession tracks a session of learning (e.g., a batch of 10 new words).
type LearningSession struct {
	ID           bson.ObjectID    `bson:"_id,omitempty"`
	UserID       bson.ObjectID    `bson:"user_id"`
	Mode         string           `bson:"mode"`
	Source       string           `bson:"source"`
	WordSetID    *bson.ObjectID   `bson:"word_set_id,omitempty"`
	Strategy     *ReviewStrategy  `bson:"strategy,omitempty"`
	WordCount    int              `bson:"word_count,omitempty"`
	Items        []SessionItemRef `bson:"items"`         // Changed from WordIDs to track sense
	CurrentIndex int              `bson:"current_index"` // Deprecated/Not heavily used in Phase L4
	Status       SessionStatus    `bson:"status"`        // "preparing", "active", "completed", "expired"
	CreatedAt    time.Time        `bson:"created_at"`
	UpdatedAt    time.Time        `bson:"updated_at"`
}

// -----------------------------------------------------------------------------
// Exercise Content Pool (Phase L4)
// -----------------------------------------------------------------------------

type ExerciseType string

const (
	ExerciseFlashcard      ExerciseType = "flashcard"
	ExerciseMultipleChoice ExerciseType = "multiple_choice"
	ExerciseFillBlank      ExerciseType = "fill_blank"
)

type ExerciseContent struct {
	ID      bson.ObjectID `bson:"_id,omitempty"`
	WordID  bson.ObjectID `bson:"word_id"`
	SenseID string        `bson:"sense_id"`
	Type    ExerciseType  `bson:"type"`

	// Flexible structure for different exercise types
	Question string   `bson:"question,omitempty"` // For Fill Blank
	Sentence string   `bson:"sentence,omitempty"` // Example sentence or full text
	Choices  []string `bson:"choices,omitempty"`  // For MCQ
	Answer   string   `bson:"answer"`             // The correct answer

	ContentHash   string `bson:"content_hash"`   // sha256(normalize(question/sentence))
	PromptVersion int    `bson:"prompt_version"` // Allows migrating AI prompts

	CreatedAt time.Time `bson:"created_at"`
	UpdatedAt time.Time `bson:"updated_at"`
}

// Quality represents the SM-2 self-assessment quality (0-5)
type Quality int

const (
	QualityBlackout Quality = 0 // Not used in UI yet
	QualityAgain    Quality = 1 // Wrong/Forgot
	QualityHard     Quality = 3 // Difficult to recall
	QualityGood     Quality = 4 // Recalled with some hesitation
	QualityEasy     Quality = 5 // Recalled perfectly
)

// Review represents an append-only log of a single review event.
type Review struct {
	ID             bson.ObjectID `bson:"_id,omitempty"`
	UserID         bson.ObjectID `bson:"user_id"`
	WordID         bson.ObjectID `bson:"word_id"`
	SenseID        string        `bson:"sense_id"`
	SenseIndex     int           `bson:"sense_index,omitempty"` // Legacy field
	Quality        Quality       `bson:"quality"`
	ResponseTimeMs int           `bson:"response_time_ms,omitempty"`
	OldInterval    int           `bson:"old_interval"`
	NewInterval    int           `bson:"new_interval"`
	OldEF          float64       `bson:"old_ef"`
	NewEF          float64       `bson:"new_ef"`
	OldRepetition  int           `bson:"old_repetition"`
	NewRepetition  int           `bson:"new_repetition"`
	ReviewedAt     time.Time     `bson:"reviewed_at"`
}

// Scheduler defines the abstraction for spaced repetition algorithms.
type Scheduler interface {
	// CalculateNext returns a modified LearningProgress based on the quality response.
	// It does NOT mutate the input progress.
	CalculateNext(p *LearningProgress, q Quality) *LearningProgress
}

// DetermineState determines the word progress state for viewers
func DetermineState(p *LearningProgress) string {
	if p == nil {
		return "new"
	}

	if !p.NextReviewDate.After(time.Now().UTC()) {
		return "due"
	}

	if p.Interval >= 21 {
		return "mastered"
	}

	return "learning"
}

// -----------------------------------------------------------------------------
// DTOs (Data Transfer Objects)
// -----------------------------------------------------------------------------

// DashboardDTO is returned by GET /learning/dashboard.
// Generated by a single aggregation — never multiple queries.
type DashboardDTO struct {
	NewCount     int     `json:"new"`
	PendingCount int     `json:"pending"`
	ReviewCount  int     `json:"review"`
	LearnedCount int     `json:"learned"`
	DueToday     int     `json:"due_today"`
	MatureWords  int     `json:"mature_words"`
	Accuracy     float64 `json:"accuracy"` // e.g., 0.85 for 85%
}

// SessionMode enumerates valid session types.
// New modes (listening, typing, ai) can be added without changing the API shape.
type SessionMode string

const (
	ModeFlashcard SessionMode = "flashcard"
	ModeQuiz      SessionMode = "quiz"
)

// ChoiceDTO is one option in a multiple-choice question.
type ChoiceDTO struct {
	Definition string `json:"definition"`
	IsCorrect  bool   `json:"is_correct"`
}

// SessionItemDTO carries the word data for one card in a session.
// For flashcard mode, Choices is nil. For quiz mode, Choices is populated.
type SessionItemDTO struct {
	WordID      string   `json:"word_id"`
	SenseID     string   `json:"sense_id"`
	Word        string   `json:"word"`
	POS         string   `json:"pos"`
	IPA         string   `json:"ipa"`
	Definitions []string `json:"definitions"` // first 3 glosses
	Example     string   `json:"example"`

	TranslationCandidates []string           `json:"translation_candidates,omitempty"`
	Annotation            *UserAnnotationDTO `json:"annotation,omitempty"`

	Choices           []ChoiceDTO `json:"choices"`             // MCQ
	FillBlankQuestion string      `json:"fill_blank_question"` // FB
	FillBlankAnswer   string      `json:"fill_blank_answer"`   // FB
}

// UserAnnotationDTO is the subset of wordset.UserAnnotation exposed to learning sessions.
type UserAnnotationDTO struct {
	Translation string `json:"translation"`
	Note        string `json:"note,omitempty"`
}

// SessionDTO is the response body for GET /learning/session and POST /learning/session.
type SessionDTO struct {
	ID           string           `json:"id"`
	Mode         SessionMode      `json:"mode"`
	Status       SessionStatus    `json:"status"`
	CurrentIndex int              `json:"current_index"`
	Items        []SessionItemDTO `json:"items"`
}

// WordResult represents the user's performance on a single word across the session phases.
type WordResult struct {
	WordID                string `json:"word_id" binding:"required"`
	SenseID               string `json:"sense_id"`
	FlashcardQuality      int    `json:"flashcard_quality" binding:"required,min=1,max=5"`
	MultipleChoiceCorrect *bool  `json:"multiple_choice_correct,omitempty"`
	FillBlankCorrect      *bool  `json:"fill_blank_correct,omitempty"`
}

// SubmitSessionRequest is the request body for POST /learning/sessions/:id/submit.
type SubmitSessionRequest struct {
	Results []WordResult `json:"results" binding:"required,dive"`
}

// LearningAssessment calculates the final quality based on evidence.
type LearningAssessment struct {
	BaseQuality  int
	MCPenalty    int
	FBPenalty    int
	FinalQuality int
}

// -----------------------------------------------------------------------------
// Repository Interface
// -----------------------------------------------------------------------------

type Repository interface {
	GetProgress(ctx context.Context, userID, wordID bson.ObjectID) (*LearningProgress, error)
	GetProgressForWords(ctx context.Context, userID bson.ObjectID, items []SessionItemRef) (map[LearningProgressKey]*LearningProgress, error)
	// UpsertProgressAndLog uses a MongoDB transaction to update progress and insert a review atomically.
	UpsertProgressAndLog(ctx context.Context, p *LearningProgress, r *Review) error
	GetDueWordIDs(ctx context.Context, userID bson.ObjectID, limit int) ([]SessionItemRef, error)
	GetDashboardCounts(ctx context.Context, userID bson.ObjectID, bookmarkedWordIDs []bson.ObjectID) (*DashboardDTO, error)
	EnsureIndexes(ctx context.Context) error
	DebugAdvanceTime(ctx context.Context, userID bson.ObjectID, days int) error
	GetActiveSession(ctx context.Context, userID bson.ObjectID, mode string) (*LearningSession, error)
	GetSessionByID(ctx context.Context, sessionID bson.ObjectID) (*LearningSession, error)
	CreateSession(ctx context.Context, session *LearningSession) error
	SubmitSessionWithTransaction(ctx context.Context, sessionID bson.ObjectID, progresses []*LearningProgress, reviews []*Review, totalXP int) error
	UpdateSessionStatus(ctx context.Context, sessionID bson.ObjectID, status SessionStatus) error
	GetPendingWordIDs(ctx context.Context, userID bson.ObjectID, limit int) ([]SessionItemRef, error)
	SuspendWord(ctx context.Context, userID bson.ObjectID, wordID bson.ObjectID) error
	GetWordIDsForWordSet(ctx context.Context, userID bson.ObjectID, setID bson.ObjectID) ([]SessionItemRef, error)
	GetActiveWordSetSession(ctx context.Context, userID bson.ObjectID, setID bson.ObjectID) (*LearningSession, error)
}

type ExerciseContentRepository interface {
	Save(ctx context.Context, content *ExerciseContent) error
	GetRandomExercises(ctx context.Context, refs []SessionItemRef, exType ExerciseType, limitPerWord int) (map[string][]ExerciseContent, error)
	EnsureIndexes(ctx context.Context) error
}

// SessionEnricher batch loads missing domain data (like translations/annotations) into SessionItemDTOs
type SessionEnricher interface {
	EnrichItems(ctx context.Context, userID bson.ObjectID, targetLang string, items []SessionItemDTO) ([]SessionItemDTO, error)
}

// -----------------------------------------------------------------------------
// Service Interface
// -----------------------------------------------------------------------------

type ExerciseContentService interface {
	// GetExercises returns the requested number of exercises per word, generating them via AI if the DB pool is insufficient.
	GetExercises(ctx context.Context, refs []SessionItemRef, exType ExerciseType, limitPerWord int) (map[string][]ExerciseContent, error)
}

type Service interface {
	GetDashboard(ctx context.Context, userID string) (*DashboardDTO, error)
	GetSession(ctx context.Context, userID string, mode SessionMode, limit int) (*SessionDTO, error)
	CreateNewWordsSession(ctx context.Context, userID string) (*SessionDTO, error)
	CreateReviewSession(ctx context.Context, userID string) (*SessionDTO, error)
	GetSessionByID(ctx context.Context, sessionID string) (*SessionDTO, error)
	SubmitSessionResults(ctx context.Context, userID string, sessionID string, req *SubmitSessionRequest) error
	SuspendWord(ctx context.Context, userID string, wordID string) error
	DebugAdvanceTime(ctx context.Context, userID string, days int) error
	CreateWordSetReviewSession(ctx context.Context, userID string, setID string, strategy ReviewStrategy, wordCount int) (*SessionDTO, error)
}
