package translation

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type Repository interface {
	EnsureIndexes(ctx context.Context) error

	// Jobs
	EnqueueJob(ctx context.Context, wordID bson.ObjectID, targetLang string) error
	ClaimJobs(ctx context.Context, limit int) ([]TranslationJob, error)
	UpdateJobStatus(ctx context.Context, id bson.ObjectID, status JobStatus, errStr string, nextAttemptAt time.Time) error
	GetJobStats(ctx context.Context) (map[string]int64, error)
	RetryFailedJobs(ctx context.Context) (int64, error)
	GetFailedJobs(ctx context.Context, limit int) ([]TranslationJob, error)

	// Translations
	SaveManyTranslations(ctx context.Context, translations []WordTranslation) error
	GetTranslationsByWordIDs(ctx context.Context, wordIDs []bson.ObjectID, targetLang string) ([]WordTranslation, error)
}
