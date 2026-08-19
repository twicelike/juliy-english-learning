package evaluation

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// EvidenceSource indicates where this evidence originated from
type EvidenceSource string

const (
	SourceLearningSession EvidenceSource = "learning_session"
	SourceWritingLab      EvidenceSource = "writing_lab"
	SourceSpeakingLab     EvidenceSource = "speaking_lab"
)

const (
	WeightReview   float64 = 1.0
	WeightWriting  float64 = 3.0
	WeightSpeaking float64 = 3.0
)

// Evidence represents a raw observation of user performance
type Evidence struct {
	ID         bson.ObjectID  `bson:"_id,omitempty"`
	UserID     bson.ObjectID  `bson:"user_id"`
	EventID    string         `bson:"event_id"` // Idempotency key to prevent duplicates
	Skill      string         `bson:"skill"`
	Score      float64        `bson:"score"`  // Normalized performance score (0-100)
	Weight     float64        `bson:"weight"` // Importance of this evidence
	Source     EvidenceSource `bson:"source"`
	Metadata   bson.M         `bson:"metadata,omitempty"` // Context (e.g., exercise type, errors, word_id, etc.)
	OccurredAt time.Time      `bson:"occurred_at"`        // When the interaction actually happened
	CreatedAt  time.Time      `bson:"created_at"`         // When the evidence was persisted
}

// Repository defines data access for Evidence
type Repository interface {
	InsertEvidence(ctx context.Context, ev *Evidence) error
	InsertMany(ctx context.Context, evs []*Evidence) error
	GetEvidencesBySkill(ctx context.Context, userID string, skill string) ([]*Evidence, error)
	EnsureIndexes(ctx context.Context) error
}

// Service defines the business logic for Evaluation
type Service interface {
	RecordEvidence(ctx context.Context, ev *Evidence) error
	RecordEvidences(ctx context.Context, evs []*Evidence) error
	CalculateSkillAssessment(ctx context.Context, userID string, skill string) (SkillAssessmentResult, error)
}
