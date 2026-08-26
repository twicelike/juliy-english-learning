package evaluation

import (
	"context"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type repository struct {
	db *mongo.Database
}

func NewRepository(db *mongo.Database) Repository {
	return &repository{db: db}
}

func (r *repository) getCollection() *mongo.Collection {
	return r.db.Collection("evidences")
}

func (r *repository) InsertEvidence(ctx context.Context, ev *Evidence) error {
	if ev.ID.IsZero() {
		ev.ID = bson.NewObjectID()
	}
	_, err := r.getCollection().InsertOne(ctx, ev)
	return err
}

func (r *repository) InsertMany(ctx context.Context, evs []*Evidence) error {
	if len(evs) == 0 {
		return nil
	}
	docs := make([]interface{}, len(evs))
	for i, ev := range evs {
		if ev.ID.IsZero() {
			ev.ID = bson.NewObjectID()
		}
		docs[i] = ev
	}
	// We use ordered: false to allow partial success if there are duplicates (idempotency key clash)
	opts := options.InsertMany().SetOrdered(false)
	_, err := r.getCollection().InsertMany(ctx, docs, opts)
	if err != nil {
		// Ignore duplicate key errors on event_id, as they just mean the evidence was already processed.
		if mongo.IsDuplicateKeyError(err) {
			return nil
		}
		return err
	}
	return nil
}

func (r *repository) GetEvidencesBySkill(ctx context.Context, userIDStr string, skill string) ([]*Evidence, error) {
	userID, err := bson.ObjectIDFromHex(userIDStr)
	if err != nil {
		return nil, err
	}

	filter := bson.M{
		"user_id": userID,
		"skill":   skill,
	}

	opts := options.Find().SetSort(bson.D{{Key: "occurred_at", Value: -1}})
	cursor, err := r.getCollection().Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var evidences []*Evidence
	if err := cursor.All(ctx, &evidences); err != nil {
		return nil, err
	}

	return evidences, nil
}

func (r *repository) EnsureIndexes(ctx context.Context) error {
	coll := r.getCollection()

	indexes := []mongo.IndexModel{
		// Unique index to prevent duplicate evidence for the same event
		{
			Keys: bson.D{
				{Key: "user_id", Value: 1},
				{Key: "event_id", Value: 1},
			},
			Options: options.Index().SetUnique(true),
		},
		// Index to quickly get recent evidence for a user
		{
			Keys: bson.D{
				{Key: "user_id", Value: 1},
				{Key: "occurred_at", Value: -1},
			},
		},
		// Index to query recent evidence by skill
		{
			Keys: bson.D{
				{Key: "user_id", Value: 1},
				{Key: "skill", Value: 1},
				{Key: "occurred_at", Value: -1},
			},
		},
	}

	_, err := coll.Indexes().CreateMany(ctx, indexes)
	return err
}
