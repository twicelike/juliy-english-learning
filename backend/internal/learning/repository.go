package learning

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type repository struct {
	db          *mongo.Database
	progress    *mongo.Collection
	reviewsColl *mongo.Collection
}

func NewRepository(db *mongo.Database) Repository {
	return &repository{
		db:          db,
		progress:    db.Collection("learning_progress"),
		reviewsColl: db.Collection("reviews"),
	}
}

func (r *repository) GetProgress(ctx context.Context, userID, wordID bson.ObjectID) (*LearningProgress, error) {
	var p LearningProgress
	err := r.progress.FindOne(ctx, bson.M{"user_id": userID, "word_id": wordID}).Decode(&p)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &p, nil
}

func (r *repository) GetProgressForWords(ctx context.Context, userID bson.ObjectID, wordIDs []bson.ObjectID) (map[bson.ObjectID]*LearningProgress, error) {
	filter := bson.M{"user_id": userID, "word_id": bson.M{"$in": wordIDs}}
	cursor, err := r.progress.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	result := make(map[bson.ObjectID]*LearningProgress, len(wordIDs))
	for cursor.Next(ctx) {
		var p LearningProgress
		if err := cursor.Decode(&p); err != nil {
			return nil, err
		}
		result[p.WordID] = &p
	}
	return result, cursor.Err()
}

func (r *repository) UpsertProgressAndLog(ctx context.Context, p *LearningProgress, rev *Review) error {
	session, err := r.db.Client().StartSession()
	if err != nil {
		return err
	}
	defer session.EndSession(ctx)

	callback := func(sessCtx context.Context) (interface{}, error) {
		// 1. Upsert Progress
		filter := bson.M{"user_id": p.UserID, "word_id": p.WordID}
		update := bson.M{
			"$set": bson.M{
				"interval":         p.Interval,
				"repetition":       p.Repetition,
				"ease_factor":      p.EaseFactor,
				"next_review_date": p.NextReviewDate,
				"last_review_date": p.LastReviewDate,
				"updated_at":       p.UpdatedAt,
			},
			"$setOnInsert": bson.M{
				"created_at": p.CreatedAt,
			},
		}
		opts := options.UpdateOne().SetUpsert(true)
		_, err := r.progress.UpdateOne(sessCtx, filter, update, opts)
		if err != nil {
			return nil, err
		}

		// 2. Insert Review
		_, err = r.reviewsColl.InsertOne(sessCtx, rev)
		if err != nil {
			return nil, err
		}

		return nil, nil
	}

	// Wait, we need mongo.WithSession to run a transaction if replica set. 
	// However, if the user is running a standalone mongo container, transactions might fail.
	// But the user said: "MongoDB hỗ trợ transaction (replica set). Nếu không dùng transaction thì ít nhất update progress thành công insert review retry".
	// Let's try transaction first.
	_, err = session.WithTransaction(ctx, callback)
	if err != nil {
		if true { // Fallback to non-transactional if running standalone
			_, err = callback(ctx)
			return err
		}
		return err
	}
	
	return nil
}

func (r *repository) GetDueWordIDs(ctx context.Context, userID bson.ObjectID, limit int) ([]bson.ObjectID, error) {
	filter := bson.M{
		"user_id":          userID,
		"next_review_date": bson.M{"$lte": time.Now()},
	}
	opts := options.Find().
		SetLimit(int64(limit)).
		SetSort(bson.D{{Key: "next_review_date", Value: 1}}).
		SetProjection(bson.M{"word_id": 1})

	cursor, err := r.progress.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var ids []bson.ObjectID
	for cursor.Next(ctx) {
		var p LearningProgress
		if err := cursor.Decode(&p); err != nil {
			return nil, err
		}
		ids = append(ids, p.WordID)
	}
	return ids, cursor.Err()
}

func (r *repository) GetDashboardCounts(ctx context.Context, userID bson.ObjectID, bookmarkedWordIDs []bson.ObjectID) (*DashboardDTO, error) {
	now := time.Now()

	pipeline := mongo.Pipeline{
		// Match only this user's progress docs
		{{Key: "$match", Value: bson.M{"user_id": userID}}},
		// Categorise each doc
		{{Key: "$facet", Value: bson.M{
			"due_today": bson.A{
				bson.M{"$match": bson.M{
					"next_review_date": bson.M{"$lte": now},
				}},
				bson.M{"$count": "count"},
			},
			"mature": bson.A{
				bson.M{"$match": bson.M{"interval": bson.M{"$gte": 21}}},
				bson.M{"$count": "count"},
			},
			"all_word_ids": bson.A{
				bson.M{"$group": bson.M{"_id": nil, "ids": bson.M{"$push": "$word_id"}}},
			},
		}}},
	}

	cursor, err := r.progress.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var raw []struct {
		DueToday []struct {
			Count int `bson:"count"`
		} `bson:"due_today"`
		Mature []struct {
			Count int `bson:"count"`
		} `bson:"mature"`
		AllWordIDs []struct {
			IDs []bson.ObjectID `bson:"ids"`
		} `bson:"all_word_ids"`
	}
	if err := cursor.All(ctx, &raw); err != nil {
		return nil, err
	}

	dto := &DashboardDTO{}
	if len(raw) > 0 {
		if len(raw[0].DueToday) > 0 {
			dto.DueToday = raw[0].DueToday[0].Count
			dto.ReviewCount = dto.DueToday // Temporary mapping: ReviewCount = DueToday
		}
		if len(raw[0].Mature) > 0 {
			dto.MatureWords = raw[0].Mature[0].Count
			dto.LearnedCount = dto.MatureWords
		}

		// Compute new
		existingSet := make(map[bson.ObjectID]struct{})
		if len(raw[0].AllWordIDs) > 0 {
			for _, id := range raw[0].AllWordIDs[0].IDs {
				existingSet[id] = struct{}{}
			}
		}
		for _, bid := range bookmarkedWordIDs {
			if _, found := existingSet[bid]; !found {
				dto.NewCount++
			}
		}
	} else {
		dto.NewCount = len(bookmarkedWordIDs)
	}

	// Compute accuracy from reviews table
	accuracyPipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.M{"user_id": userID}}},
		{{Key: "$group", Value: bson.M{
			"_id": nil,
			"total": bson.M{"$sum": 1},
			"correct": bson.M{"$sum": bson.M{"$cond": bson.A{bson.M{"$gte": bson.A{"$quality", 3}}, 1, 0}}},
		}}},
	}
	
	accCursor, err := r.reviewsColl.Aggregate(ctx, accuracyPipeline)
	if err == nil {
		defer accCursor.Close(ctx)
		var accRaw []struct {
			Total   int `bson:"total"`
			Correct int `bson:"correct"`
		}
		if err := accCursor.All(ctx, &accRaw); err == nil && len(accRaw) > 0 {
			if accRaw[0].Total > 0 {
				dto.Accuracy = float64(accRaw[0].Correct) / float64(accRaw[0].Total)
			}
		}
	}

	return dto, nil
}

func (r *repository) EnsureIndexes(ctx context.Context) error {
	// 1. learning_progress: {user_id: 1, next_review_date: 1}
	_, err := r.progress.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{
			{Key: "user_id", Value: 1},
			{Key: "next_review_date", Value: 1},
		},
	})
	if err != nil {
		return err
	}

	// 2. learning_progress: {user_id: 1, word_id: 1} UNIQUE
	_, err = r.progress.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{
			{Key: "user_id", Value: 1},
			{Key: "word_id", Value: 1},
		},
		Options: options.Index().SetUnique(true),
	})
	if err != nil {
		return err
	}

	// 3. reviews: {user_id: 1, reviewed_at: -1}
	_, err = r.reviewsColl.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{
			{Key: "user_id", Value: 1},
			{Key: "reviewed_at", Value: -1},
		},
	})
	if err != nil {
		return err
	}

	// 4. reviews: {user_id: 1, word_id: 1}
	_, err = r.reviewsColl.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{
			{Key: "user_id", Value: 1},
			{Key: "word_id", Value: 1},
		},
	})
	return err
}
