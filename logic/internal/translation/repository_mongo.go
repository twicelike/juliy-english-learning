package translation

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

const (
	collectionName = "word_translations"
)

type mongoRepository struct {
	db       *mongo.Database
	coll     *mongo.Collection
	jobsColl *mongo.Collection
}

func NewRepository(db *mongo.Database) Repository {
	return &mongoRepository{
		db:       db,
		coll:     db.Collection(collectionName),
		jobsColl: db.Collection("translation_jobs"),
	}
}

func (r *mongoRepository) EnsureIndexes(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	indexModel := mongo.IndexModel{
		Keys: bson.D{
			{Key: "word_id", Value: 1},
			{Key: "target_language", Value: 1},
		},
		Options: options.Index().
			SetUnique(true).
			SetName("word_id_1_target_language_1"),
	}

	_, err := r.coll.Indexes().CreateOne(ctx, indexModel)
	if err != nil {
		return fmt.Errorf("translation.EnsureIndexes (%s): %w", collectionName, err)
	}

	jobIndexModel := mongo.IndexModel{
		Keys: bson.D{
			{Key: "word_id", Value: 1},
			{Key: "target_language", Value: 1},
		},
		Options: options.Index().
			SetUnique(true).
			SetName("word_id_1_target_language_1"),
	}

	_, err = r.jobsColl.Indexes().CreateOne(ctx, jobIndexModel)
	if err != nil {
		return fmt.Errorf("translation.EnsureIndexes (translation_jobs): %w", err)
	}

	// Drop old $or-era index if it exists
	_ = r.jobsColl.Indexes().DropOne(ctx, "status_1_next_attempt_at_1_created_at_1")

	// ClaimJobs index: claimable_at range scan + status filter, sorted by created_at.
	// This replaces the old $or query so a single index scan covers the entire operation.
	claimJobIndexModel := mongo.IndexModel{
		Keys: bson.D{
			{Key: "claimable_at", Value: 1},
			{Key: "status", Value: 1},
			{Key: "created_at", Value: 1},
		},
		Options: options.Index().
			SetName("claimable_at_1_status_1_created_at_1"),
	}

	_, err = r.jobsColl.Indexes().CreateOne(ctx, claimJobIndexModel)
	if err != nil {
		return fmt.Errorf("translation.EnsureIndexes (translation_jobs claim_job_index): %w", err)
	}

	return nil
}

func (r *mongoRepository) EnqueueJob(ctx context.Context, wordID bson.ObjectID, targetLang string) error {
	now := time.Now()

	filter := bson.M{
		"word_id":         wordID,
		"target_language": targetLang,
	}

	update := bson.M{
		"$setOnInsert": bson.M{
			"status":   JobStatusPending,
			"attempts": 0,
			// claimable_at = zero-time means immediately claimable.
			"claimable_at":    time.Time{},
			"next_attempt_at": now,
			"created_at":      now,
			"updated_at":      now,
		},
	}

	opts := options.UpdateOne().SetUpsert(true)
	_, err := r.jobsColl.UpdateOne(ctx, filter, update, opts)
	return err
}

func (r *mongoRepository) ClaimJobs(ctx context.Context, limit int) ([]TranslationJob, error) {
	now := time.Now()
	var claimed []TranslationJob

	// claimable_at <= now covers both:
	//   - pending jobs (claimable_at = zero-time, always <= now)
	//   - failed jobs ready for retry (claimable_at = next_attempt_at)
	filter := bson.M{
		"claimable_at": bson.M{"$lte": now},
		"status":       bson.M{"$in": []JobStatus{JobStatusPending, JobStatusFailed}},
	}

	update := bson.M{
		"$set": bson.M{
			"status":     JobStatusProcessing,
			"locked_at":  now,
			"updated_at": now,
		},
		"$inc": bson.M{"attempts": 1},
	}

	// FindOneAndUpdate is ATOMIC. This prevents race conditions between multiple workers.
	opts := options.FindOneAndUpdate().
		SetSort(bson.D{{Key: "claimable_at", Value: 1}}).
		SetReturnDocument(options.After)

	for i := 0; i < limit; i++ {
		var job TranslationJob
		err := r.jobsColl.FindOneAndUpdate(ctx, filter, update, opts).Decode(&job)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				break // No more jobs to claim
			}
			return claimed, err
		}
		claimed = append(claimed, job)
	}

	return claimed, nil
}

func (r *mongoRepository) UpdateJobStatus(ctx context.Context, id bson.ObjectID, status JobStatus, errStr string, nextAttemptAt time.Time) error {
	// Keep claimable_at in sync so the ClaimJobs index stays effective.
	// Failed jobs: claimable_at = nextAttemptAt (defer the retry).
	// Completed/other: claimable_at is irrelevant but we set a far-future sentinel
	// so they are never accidentally picked up again.
	claimableAt := nextAttemptAt
	if status != JobStatusFailed {
		claimableAt = time.Date(9999, 1, 1, 0, 0, 0, 0, time.UTC)
	}

	update := bson.M{
		"$set": bson.M{
			"status":          status,
			"error":           errStr,
			"next_attempt_at": nextAttemptAt,
			"claimable_at":    claimableAt,
			"updated_at":      time.Now(),
		},
		"$unset": bson.M{
			"locked_at": "",
		},
	}

	_, err := r.jobsColl.UpdateByID(ctx, id, update)
	return err
}

func (r *mongoRepository) GetJobStats(ctx context.Context) (map[string]int64, error) {
	pipeline := mongo.Pipeline{
		{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: "$status"},
			{Key: "count", Value: bson.D{{Key: "$sum", Value: 1}}},
		}}},
	}

	cursor, err := r.jobsColl.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []struct {
		ID    string `bson:"_id"`
		Count int64  `bson:"count"`
	}

	if err := cursor.All(ctx, &results); err != nil {
		return nil, err
	}

	stats := make(map[string]int64)
	for _, res := range results {
		stats[res.ID] = res.Count
	}
	return stats, nil
}

func (r *mongoRepository) SaveManyTranslations(ctx context.Context, translations []WordTranslation) error {
	if len(translations) == 0 {
		return nil
	}

	var models []mongo.WriteModel
	for _, tr := range translations {
		filter := bson.M{
			"word_id":         tr.WordID,
			"sense_id":        tr.SenseID,
			"target_language": tr.TargetLanguage,
		}

		update := bson.M{
			"$set": bson.M{
				"translations": tr.Translations,
				"provider":     tr.Provider,
				"model":        tr.Model,
				"version":      tr.Version,
				"updated_at":   tr.UpdatedAt,
			},
			"$setOnInsert": bson.M{
				"source_language": tr.SourceLanguage,
				"created_at":      tr.CreatedAt,
			},
		}

		model := mongo.NewUpdateOneModel().SetFilter(filter).SetUpdate(update).SetUpsert(true)
		models = append(models, model)
	}

	_, err := r.coll.BulkWrite(ctx, models)
	return err
}

func (r *mongoRepository) GetTranslationsByWordIDs(ctx context.Context, wordIDs []bson.ObjectID, targetLang string) ([]WordTranslation, error) {
	if len(wordIDs) == 0 {
		return nil, nil
	}

	filter := bson.M{
		"word_id":         bson.M{"$in": wordIDs},
		"target_language": targetLang,
	}

	cursor, err := r.coll.Find(ctx, filter)
	if err != nil {
		return nil, err
	}

	var results []WordTranslation
	if err = cursor.All(ctx, &results); err != nil {
		return nil, err
	}

	return results, nil
}

func (r *mongoRepository) RetryFailedJobs(ctx context.Context) (int64, error) {
	filter := bson.M{"status": JobStatusFailed}
	
	now := time.Now()
	update := bson.M{
		"$set": bson.M{
			"status":          JobStatusPending,
			"attempts":        0,
			"error":           "",
			"claimable_at":    time.Time{},
			"next_attempt_at": now,
			"updated_at":      now,
		},
		"$unset": bson.M{
			"locked_at": "",
		},
	}

	result, err := r.jobsColl.UpdateMany(ctx, filter, update)
	if err != nil {
		return 0, err
	}
	return result.ModifiedCount, nil
}
