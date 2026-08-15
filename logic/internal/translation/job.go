package translation

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type JobStatus string

const (
	JobStatusPending    JobStatus = "pending"
	JobStatusProcessing JobStatus = "processing"
	JobStatusCompleted  JobStatus = "completed"
	JobStatusFailed     JobStatus = "failed"
)

type TranslationJob struct {
	ID             bson.ObjectID `bson:"_id,omitempty"`
	WordID         bson.ObjectID `bson:"word_id"`
	TargetLanguage string        `bson:"target_language"`
	Status         JobStatus     `bson:"status"`
	Attempts       int           `bson:"attempts"`
	// ClaimableAt is the earliest time at which this job may be claimed by a worker.
	// Pending jobs set it to zero-time (immediately claimable).
	// Failed jobs set it to next_attempt_at, deferring the retry.
	ClaimableAt    time.Time     `bson:"claimable_at"`
	NextAttemptAt  time.Time     `bson:"next_attempt_at"`
	LockedAt       *time.Time    `bson:"locked_at,omitempty"`
	Error          string        `bson:"error,omitempty"`
	CreatedAt      time.Time     `bson:"created_at"`
	UpdatedAt      time.Time     `bson:"updated_at"`
}
