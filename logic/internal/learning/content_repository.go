package learning

import (
	"context"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type exerciseContentRepository struct {
	coll *mongo.Collection
}

func NewExerciseContentRepository(db *mongo.Database) ExerciseContentRepository {
	return &exerciseContentRepository{
		coll: db.Collection("exercise_contents"),
	}
}

func (r *exerciseContentRepository) EnsureIndexes(ctx context.Context) error {
	// Unique index: (word_id, type, content_hash)
	_, err := r.coll.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{
			{Key: "word_id", Value: 1},
			{Key: "sense_id", Value: 1},
			{Key: "type", Value: 1},
			{Key: "content_hash", Value: 1},
		},
		Options: options.Index().SetUnique(true),
	})
	return err
}

func (r *exerciseContentRepository) Save(ctx context.Context, content *ExerciseContent) error {
	if content.ID == bson.NilObjectID {
		content.ID = bson.NewObjectID()
	}
	_, err := r.coll.InsertOne(ctx, content)
	return err
}

func (r *exerciseContentRepository) GetRandomExercises(ctx context.Context, refs []SessionItemRef, exType ExerciseType, limitPerWord int) (map[string][]ExerciseContent, error) {
	if len(refs) == 0 {
		return nil, nil
	}

	var orConditions []bson.M
	for _, ref := range refs {
		orConditions = append(orConditions, bson.M{
			"word_id":  ref.WordID,
			"sense_id": ref.SenseID,
		})
	}

	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.M{
			"$or":  orConditions,
			"type": exType,
		}}},
		// Group by word_id and sense_id
		{{Key: "$group", Value: bson.M{
			"_id": bson.M{
				"word_id":  "$word_id",
				"sense_id": "$sense_id",
			},
			"exercises": bson.M{"$push": "$$ROOT"},
		}}},
		// Slice the array to get at most limitPerWord random items.
		// Note: $sample is not easily usable per-group in standard aggregation without $lookup,
		// but since we only keep a small pool per word (e.g., 5), we can just fetch them and slice.
		// A more advanced approach would be to shuffle, but slicing the first N is okay for MVP since we can shuffle in Go.
		{{Key: "$project", Value: bson.M{
			"exercises": bson.M{"$slice": []interface{}{"$exercises", limitPerWord}},
		}}},
	}

	cursor, err := r.coll.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	type groupID struct {
		WordID  bson.ObjectID `bson:"word_id"`
		SenseID string        `bson:"sense_id"`
	}

	type groupResult struct {
		ID        groupID           `bson:"_id"`
		Exercises []ExerciseContent `bson:"exercises"`
	}

	var results []groupResult
	if err := cursor.All(ctx, &results); err != nil {
		return nil, err
	}

	out := make(map[string][]ExerciseContent)
	for _, res := range results {
		key := res.ID.WordID.Hex() + ":" + res.ID.SenseID
		out[key] = res.Exercises
	}

	return out, nil
}
