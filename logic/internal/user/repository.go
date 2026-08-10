package user

import (
	"context"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type Repository interface {
	AddBookmark(ctx context.Context, bookmark *SavedWord) error
	RemoveBookmark(ctx context.Context, userID bson.ObjectID, word string) error
	ListBookmarks(ctx context.Context, userID bson.ObjectID) ([]SavedWord, error)
	EnsureIndexes(ctx context.Context) error
}

type repository struct {
	coll *mongo.Collection
}

func NewRepository(db *mongo.Database) Repository {
	return &repository{
		coll: db.Collection("saved_words"),
	}
}

func (r *repository) AddBookmark(ctx context.Context, bookmark *SavedWord) error {
	res, err := r.coll.InsertOne(ctx, bookmark)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return nil // Already bookmarked, idempotent
		}
		return err
	}
	if id, ok := res.InsertedID.(bson.ObjectID); ok {
		bookmark.ID = id
	}
	return nil
}

func (r *repository) RemoveBookmark(ctx context.Context, userID bson.ObjectID, word string) error {
	filter := bson.M{
		"user_id": userID,
		"word":    word,
	}
	_, err := r.coll.DeleteOne(ctx, filter)
	return err
}

func (r *repository) ListBookmarks(ctx context.Context, userID bson.ObjectID) ([]SavedWord, error) {
	filter := bson.M{"user_id": userID}
	opts := options.Find().SetSort(bson.D{{Key: "added_at", Value: -1}})

	cursor, err := r.coll.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var bookmarks []SavedWord
	if err := cursor.All(ctx, &bookmarks); err != nil {
		return nil, err
	}

	return bookmarks, nil
}

func (r *repository) EnsureIndexes(ctx context.Context) error {
	// Index: user_id
	_, err := r.coll.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "user_id", Value: 1}},
	})
	if err != nil {
		return err
	}

	// Index: compound user_id + word (UNIQUE)
	_, err = r.coll.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{
			{Key: "user_id", Value: 1},
			{Key: "word", Value: 1},
		},
		Options: options.Index().SetUnique(true),
	})
	return err
}
