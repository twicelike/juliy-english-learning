package writing

import (
	"context"
	"time"

	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type Repository struct {
	db             *mongo.Database
	missionsCol    *mongo.Collection
	progressCol    *mongo.Collection
}

func NewRepository(db *mongo.Database) *Repository {
	return &Repository{
		db:             db,
		missionsCol:    db.Collection("writing_missions"),
		progressCol:    db.Collection("user_writing_progress"),
	}
}

// EnsureIndexes creates necessary indexes for the collections
func (r *Repository) EnsureIndexes(ctx context.Context) error {
	// writing_missions indexes
	missionIndexes := []mongo.IndexModel{
		{
			Keys: bson.D{{Key: "hash", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: bson.D{
				{Key: "topic", Value: 1},
				{Key: "scenario", Value: 1},
				{Key: "difficulty", Value: 1},
			},
		},
	}

	_, err := r.missionsCol.Indexes().CreateMany(ctx, missionIndexes)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create indexes for writing_missions")
		return err
	}

	// user_writing_progress indexes
	progressIndexes := []mongo.IndexModel{
		{
			Keys: bson.D{
				{Key: "user_id", Value: 1},
				{Key: "mission_id", Value: 1},
			},
			Options: options.Index().SetUnique(true),
		},
	}

	_, err = r.progressCol.Indexes().CreateMany(ctx, progressIndexes)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create indexes for user_writing_progress")
		return err
	}

	return nil
}

// GetMissionCountByGroup counts missions for a specific topic, scenario, and difficulty
func (r *Repository) GetMissionCountByGroup(ctx context.Context, topic, scenario, difficulty string) (int64, error) {
	filter := bson.M{
		"topic":      topic,
		"scenario":   scenario,
		"difficulty": difficulty,
	}
	return r.missionsCol.CountDocuments(ctx, filter)
}

// InsertMissions bulk inserts missions, ignoring duplicate keys (hash)
func (r *Repository) InsertMissions(ctx context.Context, missions []WritingMission) (int, error) {
	if len(missions) == 0 {
		return 0, nil
	}

	var docs []interface{}
	for _, m := range missions {
		if m.ID.IsZero() {
			m.ID = bson.NewObjectID()
		}
		docs = append(docs, m)
	}

	// Ordered: false means it continues inserting even if a duplicate key error occurs
	opts := options.InsertMany().SetOrdered(false)
	res, err := r.missionsCol.InsertMany(ctx, docs, opts)
	
	if err != nil {
		// If it's a BulkWriteException, it might just be duplicate keys which is fine
		if mongo.IsDuplicateKeyError(err) {
			if res != nil {
				return len(res.InsertedIDs), nil
			}
			// In v2, sometimes IsDuplicateKeyError covers BulkWriteException partially,
			// We can log a warning but still return the number of inserted IDs.
			log.Debug().Err(err).Msg("Duplicate keys ignored during bulk insert")
			return 0, nil // worst case return 0 if we can't extract inserted count easily, but usually res is populated
		}
		return 0, err
	}
	
	return len(res.InsertedIDs), nil
}

// GetNewMission gets a random mission the user has never done.
func (r *Repository) GetNewMission(ctx context.Context, userID bson.ObjectID, topic, difficulty string) (*WritingMission, error) {
	matchStage := bson.M{}
	if topic != "" {
		matchStage["topic"] = topic
	}
	if difficulty != "" {
		matchStage["difficulty"] = difficulty
	}

	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: matchStage}},
		{{Key: "$lookup", Value: bson.M{
			"from": "user_writing_progress",
			"let":  bson.M{"missionId": "$_id"},
			"pipeline": mongo.Pipeline{
				{{Key: "$match", Value: bson.M{
					"$expr": bson.M{
						"$and": []bson.M{
							{"$eq": []interface{}{"$mission_id", "$$missionId"}},
							{"$eq": []interface{}{"$user_id", userID}},
						},
					},
				}}},
			},
			"as": "progress",
		}}},
		{{Key: "$match", Value: bson.M{"progress": bson.M{"$size": 0}}}},
		{{Key: "$sample", Value: bson.M{"size": 1}}},
	}

	cursor, err := r.missionsCol.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []WritingMission
	if err = cursor.All(ctx, &results); err != nil {
		return nil, err
	}

	if len(results) == 0 {
		return nil, nil // not found
	}
	return &results[0], nil
}

// GetMissionByID gets a writing mission by its ID.
func (r *Repository) GetMissionByID(ctx context.Context, id bson.ObjectID) (*WritingMission, error) {
	var mission WritingMission
	err := r.missionsCol.FindOne(ctx, bson.M{"_id": id}).Decode(&mission)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &mission, nil
}

// GetReviewMission gets a random mission the user has completed, prioritizing low scores and older reviews.
func (r *Repository) GetReviewMission(ctx context.Context, userID bson.ObjectID, topic, difficulty string) (*WritingMission, error) {
	matchProgress := bson.M{"user_id": userID}
	
	missionMatch := bson.M{}
	if topic != "" {
		missionMatch["mission.topic"] = topic
	}
	if difficulty != "" {
		missionMatch["mission.difficulty"] = difficulty
	}

	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: matchProgress}},
		{{Key: "$lookup", Value: bson.M{
			"from":         "writing_missions",
			"localField":   "mission_id",
			"foreignField": "_id",
			"as":           "mission",
		}}},
		{{Key: "$unwind", Value: "$mission"}},
	}
	
	// Add mission filters if any
	if len(missionMatch) > 0 {
		pipeline = append(pipeline, bson.D{{Key: "$match", Value: missionMatch}})
	}

	// Add fields for priority calculation
	// priority = (100 - last_score)*0.7 + (days_since_review)*0.3
	// We do this approximately with a $addFields step
	pipeline = append(pipeline, bson.D{{Key: "$addFields", Value: bson.M{
		"days_since": bson.M{
			"$divide": []interface{}{
				bson.M{"$subtract": []interface{}{time.Now(), "$last_completed_at"}},
				1000 * 60 * 60 * 24,
			},
		},
	}}})

	pipeline = append(pipeline, bson.D{{Key: "$addFields", Value: bson.M{
		"priority": bson.M{
			"$add": []interface{}{
				bson.M{"$multiply": []interface{}{bson.M{"$subtract": []interface{}{100, "$last_score"}}, 0.7}},
				bson.M{"$multiply": []interface{}{"$days_since", 0.3}},
			},
		},
	}}})

	pipeline = append(pipeline, 
		bson.D{{Key: "$sort", Value: bson.M{"priority": -1}}},
		bson.D{{Key: "$limit", Value: 10}}, // Top 10 worst/oldest
		bson.D{{Key: "$sample", Value: bson.M{"size": 1}}}, // Random 1 from Top 10
	)

	cursor, err := r.progressCol.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	// Since we unwound mission, we need to extract it
	type Result struct {
		Mission WritingMission `bson:"mission"`
	}
	
	var results []Result
	if err = cursor.All(ctx, &results); err != nil {
		return nil, err
	}

	if len(results) == 0 {
		return nil, nil // not found
	}
	return &results[0].Mission, nil
}

// GetProgress gets the user's progress for a specific mission.
func (r *Repository) GetProgress(ctx context.Context, userID, missionID bson.ObjectID) (*UserWritingProgress, error) {
	var progress UserWritingProgress
	err := r.progressCol.FindOne(ctx, bson.M{
		"user_id":    userID,
		"mission_id": missionID,
	}).Decode(&progress)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &progress, nil
}

// UpsertProgress updates or inserts a user's writing progress.
func (r *Repository) UpsertProgress(ctx context.Context, progress *UserWritingProgress) error {
	filter := bson.M{
		"user_id":    progress.UserID,
		"mission_id": progress.MissionID,
	}
	
	update := bson.M{
		"$inc": bson.M{"attempts": 1},
		"$set": bson.M{
			"last_score":        progress.LastScore,
			"average_score":     progress.AverageScore,
			"completed":         progress.Completed,
			"mastery_level":     progress.MasteryLevel,
			"next_review_at":    progress.NextReviewAt,
			"last_completed_at": progress.LastCompletedAt,
			"updated_at":        time.Now(),
		},
		"$max": bson.M{
			"best_score": progress.BestScore,
		},
		"$setOnInsert": bson.M{
			"first_completed_at": progress.FirstCompletedAt,
			"created_at":         time.Now(),
		},
	}

	opts := options.UpdateOne().SetUpsert(true)
	_, err := r.progressCol.UpdateOne(ctx, filter, update, opts)
	return err
}

// GetDashboardStats computes aggregated stats for a user's writing progress.
func (r *Repository) GetDashboardStats(ctx context.Context, userID bson.ObjectID) (*DashboardStatsDTO, error) {
	matchStage := bson.M{"user_id": userID, "completed": true}
	
	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: matchStage}},
		{{Key: "$group", Value: bson.M{
			"_id": nil,
			"total_completed": bson.M{"$sum": 1},
			"average_score":   bson.M{"$avg": "$average_score"},
			"mastered": bson.M{"$sum": bson.M{
				"$cond": []interface{}{bson.M{"$eq": []interface{}{"$mastery_level", "Mastered"}}, 1, 0},
			}},
			"familiar": bson.M{"$sum": bson.M{
				"$cond": []interface{}{bson.M{"$eq": []interface{}{"$mastery_level", "Familiar"}}, 1, 0},
			}},
			"beginner": bson.M{"$sum": bson.M{
				"$cond": []interface{}{bson.M{"$eq": []interface{}{"$mastery_level", "Beginner"}}, 1, 0},
			}},
		}}},
	}

	cursor, err := r.progressCol.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	type aggResult struct {
		TotalCompleted int     `bson:"total_completed"`
		AverageScore   float64 `bson:"average_score"`
		Mastered       int     `bson:"mastered"`
		Familiar       int     `bson:"familiar"`
		Beginner       int     `bson:"beginner"`
	}

	var results []aggResult
	if err = cursor.All(ctx, &results); err != nil {
		return nil, err
	}

	stats := &DashboardStatsDTO{}
	if len(results) > 0 {
		res := results[0]
		stats.TotalCompleted = res.TotalCompleted
		stats.AverageScore = res.AverageScore
		stats.MasteryMastered = res.Mastered
		stats.MasteryFamiliar = res.Familiar
		stats.MasteryBeginner = res.Beginner
	}

	return stats, nil
}
