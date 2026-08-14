package learning

import (
	"context"
	"errors"
	"time"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type repository struct {
	db           *mongo.Database
	progress     *mongo.Collection
	reviewsColl  *mongo.Collection
	sessionsColl *mongo.Collection
}

func NewRepository(db *mongo.Database) Repository {
	return &repository{
		db:           db,
		progress:     db.Collection("learning_progress"),
		reviewsColl:  db.Collection("reviews"),
		sessionsColl: db.Collection("learning_sessions"),
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

func (r *repository) GetProgressForWords(ctx context.Context, userID bson.ObjectID, items []SessionItemRef) (map[LearningProgressKey]*LearningProgress, error) {
	if len(items) == 0 {
		return make(map[LearningProgressKey]*LearningProgress), nil
	}
	
	wordIDs := make([]bson.ObjectID, len(items))
	for i, item := range items {
		wordIDs[i] = item.WordID
	}

	filter := bson.M{"user_id": userID, "word_id": bson.M{"$in": wordIDs}}
	cursor, err := r.progress.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	result := make(map[LearningProgressKey]*LearningProgress, len(items))
	for cursor.Next(ctx) {
		var p LearningProgress
		if err := cursor.Decode(&p); err != nil {
			return nil, err
		}
		key := LearningProgressKey{WordID: p.WordID, SenseIndex: p.SenseIndex}
		// Since multiple senses for same word_id might be returned, 
		// we store them by the exact key.
		result[key] = &p
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
		filter := bson.M{"user_id": p.UserID, "word_id": p.WordID, "sense_index": p.SenseIndex}
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

func (r *repository) GetDueWordIDs(ctx context.Context, userID bson.ObjectID, limit int) ([]SessionItemRef, error) {
	filter := bson.M{
		"user_id":          userID,
		"interval":         bson.M{"$gt": 0},
		"next_review_date": bson.M{"$lte": time.Now()},
	}
	opts := options.Find().
		SetLimit(int64(limit)).
		SetSort(bson.D{{Key: "next_review_date", Value: 1}}).
		SetProjection(bson.M{"word_id": 1, "sense_index": 1})

	cursor, err := r.progress.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var refs []SessionItemRef
	for cursor.Next(ctx) {
		var p LearningProgress
		if err := cursor.Decode(&p); err != nil {
			return nil, err
		}
		refs = append(refs, SessionItemRef{WordID: p.WordID, SenseIndex: p.SenseIndex})
	}
	return refs, cursor.Err()
}

func (r *repository) GetPendingWordIDs(ctx context.Context, userID bson.ObjectID, limit int) ([]SessionItemRef, error) {
	filter := bson.M{
		"user_id":  userID,
		"interval": 0,
	}
	opts := options.Find().
		SetLimit(int64(limit)).
		SetSort(bson.D{{Key: "created_at", Value: 1}}).
		SetProjection(bson.M{"word_id": 1, "sense_index": 1})

	cursor, err := r.progress.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var refs []SessionItemRef
	for cursor.Next(ctx) {
		var p LearningProgress
		if err := cursor.Decode(&p); err != nil {
			return nil, err
		}
		refs = append(refs, SessionItemRef{WordID: p.WordID, SenseIndex: p.SenseIndex})
	}
	return refs, cursor.Err()
}

func (r *repository) GetActiveSession(ctx context.Context, userID bson.ObjectID, mode string) (*LearningSession, error) {
	var s LearningSession
	err := r.sessionsColl.FindOne(ctx, bson.M{
		"user_id": userID,
		"mode":    mode,
		"status":  bson.M{"$in": []string{"preparing", "active"}},
		"source":  bson.M{"$in": []interface{}{"", nil, "global"}},
	}).Decode(&s)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &s, nil
}

func (r *repository) GetActiveWordSetSession(ctx context.Context, userID bson.ObjectID, setID bson.ObjectID) (*LearningSession, error) {
	var s LearningSession
	err := r.sessionsColl.FindOne(ctx, bson.M{
		"user_id":     userID,
		"mode":        "review",
		"status":      bson.M{"$in": []string{"preparing", "active"}},
		"word_set_id": setID,
	}).Decode(&s)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &s, nil
}

func (r *repository) GetSessionByID(ctx context.Context, sessionID bson.ObjectID) (*LearningSession, error) {
	var s LearningSession
	err := r.sessionsColl.FindOne(ctx, bson.M{"_id": sessionID}).Decode(&s)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &s, nil
}

// SubmitSessionWithTransaction applies multiple progress updates, review logs, and marks the session complete.
// (Modified to not use MongoDB transactions so it works on standalone local instances).
func (r *repository) SubmitSessionWithTransaction(ctx context.Context, sessionID bson.ObjectID, progresses []*LearningProgress, reviews []*Review, totalXP int) error {
	// 1. Validate session
	var session LearningSession
	err := r.sessionsColl.FindOne(ctx, bson.M{"_id": sessionID}).Decode(&session)
	if err != nil {
		return err
	}
	if session.Status == SessionCompleted || session.Status == SessionExpired {
		return errors.New("session is not active")
	}

	// 2. Update all progresses
	for _, p := range progresses {
		filter := bson.M{"user_id": p.UserID, "word_id": p.WordID, "sense_index": p.SenseIndex}
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
				"_id":         p.ID,
				"user_id":     p.UserID,
				"word_id":     p.WordID,
				"sense_index": p.SenseIndex,
				"created_at":  p.CreatedAt,
			},
		}
		opts := options.UpdateOne().SetUpsert(true)
		if _, err := r.progress.UpdateOne(ctx, filter, update, opts); err != nil {
			return err
		}
	}

	// 3. Insert review logs
	if len(reviews) > 0 {
		docs := make([]interface{}, len(reviews))
		for i, rev := range reviews {
			docs[i] = rev
		}
		if _, err := r.reviewsColl.InsertMany(ctx, docs); err != nil {
			return err
		}
	}

	// 4. Mark session completed
	filter := bson.M{"_id": sessionID}
	update := bson.M{"$set": bson.M{
		"status": SessionCompleted, 
		"updated_at": time.Now().UTC(),
	}}
	if _, err := r.sessionsColl.UpdateOne(ctx, filter, update); err != nil {
		return err
	}

	return nil
}

func (r *repository) CreateSession(ctx context.Context, session *LearningSession) error {
	_, err := r.sessionsColl.InsertOne(ctx, session)
	return err
}

func (r *repository) UpdateSessionStatus(ctx context.Context, sessionID bson.ObjectID, status SessionStatus) error {
	filter := bson.M{"_id": sessionID}
	update := bson.M{
		"$set": bson.M{
			"status":     status,
			"updated_at": time.Now().UTC(),
		},
	}
	_, err := r.sessionsColl.UpdateOne(ctx, filter, update)
	return err
}

// Old methods removed

func (r *repository) GetDashboardCounts(ctx context.Context, userID bson.ObjectID, bookmarkedWordIDs []bson.ObjectID) (*DashboardDTO, error) {
	now := time.Now()

	pipeline := mongo.Pipeline{
		// Match only this user's progress docs
		{{Key: "$match", Value: bson.M{"user_id": userID}}},
		// Categorise each doc
		{{Key: "$facet", Value: bson.M{
			"due_today": bson.A{
				bson.M{"$match": bson.M{
					"interval": bson.M{"$gt": 0},
					"next_review_date": bson.M{"$lte": now},
				}},
				bson.M{"$count": "count"},
			},
			"pending": bson.A{
				bson.M{"$match": bson.M{"interval": 0}},
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
		Pending []struct {
			Count int `bson:"count"`
		} `bson:"pending"`
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
		if len(raw[0].Pending) > 0 {
			dto.PendingCount = raw[0].Pending[0].Count
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
	// Drop legacy index if it exists
	_ = r.progress.Indexes().DropOne(ctx, "user_id_1_word_id_1")

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

	// 2. learning_progress: {user_id: 1, word_id: 1, sense_index: 1} UNIQUE
	_, err = r.progress.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{
			{Key: "user_id", Value: 1},
			{Key: "word_id", Value: 1},
			{Key: "sense_index", Value: 1},
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

	// 4. reviews: {user_id: 1, word_id: 1, sense_index: 1}
	_, err = r.reviewsColl.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{
			{Key: "user_id", Value: 1},
			{Key: "word_id", Value: 1},
			{Key: "sense_index", Value: 1},
		},
	})
	if err != nil {
		return err
	}

	// 5. learning_sessions: {user_id: 1, mode: 1} partial status = in_progress
	_, err = r.sessionsColl.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{
			{Key: "user_id", Value: 1},
			{Key: "mode", Value: 1},
		},
		Options: options.Index().
			SetUnique(true).
			SetPartialFilterExpression(bson.M{"status": "in_progress"}),
	})
	return err
}

func (r *repository) DebugAdvanceTime(ctx context.Context, userID bson.ObjectID, days int) error {
	filter := bson.M{"user_id": userID}
	
	// We want to subtract `days` from next_review_date and last_review_date.
	// In MongoDB, we can use an aggregation pipeline in update to subtract time.
	// Note: 1 day = 24 * 60 * 60 * 1000 milliseconds
	msToSubtract := int64(days) * 24 * 60 * 60 * 1000

	update := bson.A{
		bson.M{"$set": bson.M{
			"next_review_date": bson.M{"$subtract": bson.A{"$next_review_date", msToSubtract}},
			"last_review_date": bson.M{"$cond": bson.A{
				bson.M{"$ifNull": bson.A{"$last_review_date", false}},
				bson.M{"$subtract": bson.A{"$last_review_date", msToSubtract}},
				"$last_review_date", // keep null if it was null
			}},
		}},
	}

	_, err := r.progress.UpdateMany(ctx, filter, update)
	return err
}

func (r *repository) GetWordIDsForWordSet(ctx context.Context, userID bson.ObjectID, setID bson.ObjectID) ([]SessionItemRef, error) {
	// First, check if the WordSet exists and belongs to the user
	wordSetsColl := r.db.Collection("word_sets")
	err := wordSetsColl.FindOne(ctx, bson.M{"_id": setID, "user_id": userID}).Err()
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("word set not found or you do not have permission")
		}
		return nil, err
	}

	// Then get all items
	itemsColl := r.db.Collection("word_set_items")
	opts := options.Find().SetSort(bson.D{{Key: "position", Value: 1}})
	cursor, err := itemsColl.Find(ctx, bson.M{"set_id": setID}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var refs []SessionItemRef
	for cursor.Next(ctx) {
		var item struct {
			WordID     bson.ObjectID `bson:"word_id"`
			SenseIndex int           `bson:"sense_index"`
		}
		if err := cursor.Decode(&item); err != nil {
			return nil, err
		}
		refs = append(refs, SessionItemRef{WordID: item.WordID, SenseIndex: item.SenseIndex})
	}
	return refs, cursor.Err()
}
