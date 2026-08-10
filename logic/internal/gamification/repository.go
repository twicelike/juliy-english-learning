package gamification

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

func (r *repository) GetDailyProgress(ctx context.Context, userID bson.ObjectID, date string) (*DailyProgress, error) {
	var dp DailyProgress
	err := r.db.Collection("daily_progress").FindOne(ctx, bson.M{"user_id": userID, "date": date}).Decode(&dp)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, fmt.Errorf("gamification.GetDailyProgress: %w", err)
	}
	return &dp, nil
}

func (r *repository) UpsertDailyProgress(ctx context.Context, dp *DailyProgress) error {
	filter := bson.M{"user_id": dp.UserID, "date": dp.Date}
	update := bson.M{"$set": dp}

	opts := options.UpdateOne().SetUpsert(true)
	_, err := r.db.Collection("daily_progress").UpdateOne(ctx, filter, update, opts)
	if err != nil {
		return fmt.Errorf("gamification.UpsertDailyProgress: %w", err)
	}
	return nil
}

func (r *repository) GetProgressHistory(ctx context.Context, userID bson.ObjectID, startDate, endDate string) ([]DailyProgress, error) {
	filter := bson.M{
		"user_id": userID,
		"date": bson.M{
			"$gte": startDate,
			"$lte": endDate,
		},
	}
	opts := options.Find().SetSort(bson.D{{Key: "date", Value: 1}})

	cursor, err := r.db.Collection("daily_progress").Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("gamification.GetProgressHistory: %w", err)
	}
	defer cursor.Close(ctx)

	var results []DailyProgress
	if err := cursor.All(ctx, &results); err != nil {
		return nil, fmt.Errorf("gamification.GetProgressHistory decode: %w", err)
	}

	return results, nil
}

func (r *repository) GetUserStats(ctx context.Context, userID bson.ObjectID) (*UserStats, error) {
	var stats UserStats
	err := r.db.Collection("user_stats").FindOne(ctx, bson.M{"user_id": userID}).Decode(&stats)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, fmt.Errorf("gamification.GetUserStats: %w", err)
	}
	return &stats, nil
}

func (r *repository) UpsertUserStats(ctx context.Context, stats *UserStats) error {
	filter := bson.M{"user_id": stats.UserID}
	update := bson.M{"$set": stats}

	opts := options.UpdateOne().SetUpsert(true)
	_, err := r.db.Collection("user_stats").UpdateOne(ctx, filter, update, opts)
	if err != nil {
		return fmt.Errorf("gamification.UpsertUserStats: %w", err)
	}
	return nil
}

func (r *repository) HasAchievement(ctx context.Context, userID bson.ObjectID, achievementID string) (bool, error) {
	count, err := r.db.Collection("user_achievements").CountDocuments(ctx, bson.M{
		"user_id":        userID,
		"achievement_id": achievementID,
	})
	if err != nil {
		return false, fmt.Errorf("gamification.HasAchievement: %w", err)
	}
	return count > 0, nil
}

func (r *repository) UnlockAchievement(ctx context.Context, ua *UserAchievement) error {
	_, err := r.db.Collection("user_achievements").InsertOne(ctx, ua)
	if err != nil {
		return fmt.Errorf("gamification.UnlockAchievement: %w", err)
	}
	return nil
}

func (r *repository) GetUserAchievements(ctx context.Context, userID bson.ObjectID) ([]UserAchievement, error) {
	cursor, err := r.db.Collection("user_achievements").Find(ctx, bson.M{"user_id": userID})
	if err != nil {
		return nil, fmt.Errorf("gamification.GetUserAchievements: %w", err)
	}
	defer cursor.Close(ctx)

	var results []UserAchievement
	if err := cursor.All(ctx, &results); err != nil {
		return nil, fmt.Errorf("gamification.GetUserAchievements decode: %w", err)
	}
	return results, nil
}

func (r *repository) EnsureIndexes(ctx context.Context) error {
	// daily_progress indexes
	dpIndex := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "user_id", Value: 1}, {Key: "date", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
	}
	if _, err := r.db.Collection("daily_progress").Indexes().CreateMany(ctx, dpIndex); err != nil {
		return err
	}

	// user_stats indexes
	usIndex := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "user_id", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
	}
	if _, err := r.db.Collection("user_stats").Indexes().CreateMany(ctx, usIndex); err != nil {
		return err
	}

	// user_achievements indexes
	uaIndex := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "user_id", Value: 1}, {Key: "achievement_id", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
	}
	if _, err := r.db.Collection("user_achievements").Indexes().CreateMany(ctx, uaIndex); err != nil {
		return err
	}

	return nil
}
