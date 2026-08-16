package ai

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type repository struct {
	coll *mongo.Collection
}

// NewRepository creates a new AI repository.
func NewRepository(db *mongo.Database) ExplanationRepo {
	return &repository{
		coll: db.Collection("ai_explanations"),
	}
}

func (r *repository) GetSenseExplanation(ctx context.Context, wordID string, senseIndex int, level, nativeLang string, promptVersion int) (*SenseExplanation, error) {
	filter := bson.M{
		"word_id":         wordID,
		"sense_index":     senseIndex,
		"userLevel":       level,
		"nativeLanguage":  nativeLang,
		"prompt_version":  promptVersion,
	}
	var doc SenseExplanation
	err := r.coll.FindOne(ctx, filter).Decode(&doc)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &doc, nil
}

func (r *repository) GetCachedSenseIndices(ctx context.Context, wordID string, level, nativeLang string, promptVersion int) (map[int]bool, error) {
	filter := bson.M{
		"word_id":         wordID,
		"userLevel":       level,
		"nativeLanguage":  nativeLang,
		"prompt_version":  promptVersion,
	}

	cursor, err := r.coll.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	indices := make(map[int]bool)
	for cursor.Next(ctx) {
		var doc SenseExplanation
		if err := cursor.Decode(&doc); err != nil {
			return nil, err
		}
		indices[doc.SenseIndex] = true
	}
	return indices, nil
}

func (r *repository) SaveBatchSenseExplanations(ctx context.Context, explanations []SenseExplanation) error {
	if len(explanations) == 0 {
		return nil
	}

	var models []mongo.WriteModel
	for _, exp := range explanations {
		now := time.Now()
		exp.UpdatedAt = now
		if exp.CreatedAt.IsZero() {
			exp.CreatedAt = now
		}

		filter := bson.M{
			"word_id":         exp.WordID,
			"sense_index":     exp.SenseIndex,
			"userLevel":       exp.UserLevel,
			"nativeLanguage":  exp.NativeLanguage,
			"prompt_version":  exp.PromptVersion,
		}

		update := bson.M{"$set": exp}
		model := mongo.NewUpdateOneModel().
			SetFilter(filter).
			SetUpdate(update).
			SetUpsert(true)

		models = append(models, model)
	}

	_, err := r.coll.BulkWrite(ctx, models)
	return err
}
