package match

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type repository struct {
	coll *mongo.Collection
}

func NewRepository(db *mongo.Database) Repository {
	return &repository{
		coll: db.Collection("user_swipes"),
	}
}

func (r *repository) EnsureIndexes(ctx context.Context) error {
	indexes := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "user_id", Value: 1}, {Key: "word_id", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: bson.D{{Key: "word_id", Value: 1}, {Key: "action", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "user_id", Value: 1}, {Key: "created_at", Value: -1}},
		},
	}
	_, err := r.coll.Indexes().CreateMany(ctx, indexes)
	return err
}

func (r *repository) RecordSwipe(ctx context.Context, swipe *UserSwipe) error {
	swipe.ID = bson.NewObjectID()
	swipe.CreatedAt = time.Now()

	// Use upsert to handle the unique constraint gracefully
	filter := bson.M{"user_id": swipe.UserID, "word_id": swipe.WordID}
	update := bson.M{
		"$set": bson.M{
			"action":     swipe.Action,
			"created_at": swipe.CreatedAt,
		},
	}
	opts := options.UpdateOne().SetUpsert(true)
	_, err := r.coll.UpdateOne(ctx, filter, update, opts)
	return err
}

func (r *repository) GetSwipedWordIDs(ctx context.Context, userID bson.ObjectID) ([]bson.ObjectID, error) {
	cursor, err := r.coll.Find(ctx, bson.M{"user_id": userID}, options.Find().SetProjection(bson.M{"word_id": 1}))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var ids []bson.ObjectID
	for cursor.Next(ctx) {
		var doc struct {
			WordID bson.ObjectID `bson:"word_id"`
		}
		if err := cursor.Decode(&doc); err != nil {
			return nil, err
		}
		ids = append(ids, doc.WordID)
	}
	return ids, cursor.Err()
}

func (r *repository) GetCommunitySignal(ctx context.Context, wordID bson.ObjectID) (likes int, total int, err error) {
	// Aggregate to count likes and total
	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.M{"word_id": wordID}}},
		{{Key: "$group", Value: bson.M{
			"_id":   nil,
			"total": bson.M{"$sum": 1},
			"likes": bson.M{"$sum": bson.M{"$cond": bson.A{bson.M{"$eq": bson.A{"$action", SwipeKeep}}, 1, 0}}},
		}}},
	}

	cursor, err := r.coll.Aggregate(ctx, pipeline)
	if err != nil {
		return 0, 0, err
	}
	defer cursor.Close(ctx)

	if cursor.Next(ctx) {
		var result struct {
			Total int `bson:"total"`
			Likes int `bson:"likes"`
		}
		if err := cursor.Decode(&result); err != nil {
			return 0, 0, err
		}
		return result.Likes, result.Total, nil
	}
	return 0, 0, nil
}
