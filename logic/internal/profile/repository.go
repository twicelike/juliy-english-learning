package profile

import (
	"context"
	"fmt"

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

func (r *repository) GetByUserID(ctx context.Context, userID bson.ObjectID) (*UserProfile, error) {
	var profile UserProfile
	err := r.db.Collection("user_profiles").FindOne(ctx, bson.M{"user_id": userID}).Decode(&profile)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil // Return nil, nil to indicate not found (cleaner than returning error)
		}
		return nil, fmt.Errorf("profile.repository.GetByUserID: %w", err)
	}
	return &profile, nil
}

func (r *repository) Upsert(ctx context.Context, profile *UserProfile) error {
	filter := bson.M{"user_id": profile.UserID}
	update := bson.M{"$set": profile}

	opts := options.UpdateOne().SetUpsert(true)
	_, err := r.db.Collection("user_profiles").UpdateOne(ctx, filter, update, opts)
	if err != nil {
		return fmt.Errorf("profile.repository.Upsert: %w", err)
	}

	return nil
}

func (r *repository) EnsureIndexes(ctx context.Context) error {
	collection := r.db.Collection("user_profiles")

	indexes := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "user_id", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
	}

	_, err := collection.Indexes().CreateMany(ctx, indexes)
	if err != nil {
		return fmt.Errorf("failed to create profile indexes: %w", err)
	}

	return nil
}
