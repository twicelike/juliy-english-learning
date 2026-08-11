package learning

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func (r *repository) SuspendWord(ctx context.Context, userID bson.ObjectID, wordID bson.ObjectID) error {
	filter := bson.M{"user_id": userID, "word_id": wordID}
	update := bson.M{
		"$set": bson.M{
			"interval": -1,
			"updated_at": time.Now().UTC(),
		},
		"$setOnInsert": bson.M{
			"created_at": time.Now().UTC(),
		},
	}
	opts := options.UpdateOne().SetUpsert(true)
	_, err := r.progress.UpdateOne(ctx, filter, update, opts)
	return err
}
