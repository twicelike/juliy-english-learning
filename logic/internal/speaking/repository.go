package speaking

import (
	"context"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type repository struct {
	tasksColl    *mongo.Collection
	sessionsColl *mongo.Collection
	attemptsColl *mongo.Collection
}

// NewRepository creates a new speaking repository.
func NewRepository(db *mongo.Database) Repository {
	r := &repository{
		tasksColl:    db.Collection("speaking_tasks"),
		sessionsColl: db.Collection("speaking_sessions"),
		attemptsColl: db.Collection("speaking_attempts"),
	}
	r.EnsureIndexes(context.Background())
	return r
}

func (r *repository) EnsureIndexes(ctx context.Context) error {
	_, err := r.sessionsColl.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "user_id", Value: 1}, {Key: "started_at", Value: -1}},
	})
	if err != nil {
		return err
	}

	_, err = r.attemptsColl.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "session_id", Value: 1}},
	})
	return err
}

// -----------------------------------------------------------------------------
// Tasks
// -----------------------------------------------------------------------------

func (r *repository) GetTaskByID(ctx context.Context, id bson.ObjectID) (*SpeakingTask, error) {
	var task SpeakingTask
	err := r.tasksColl.FindOne(ctx, bson.M{"_id": id}).Decode(&task)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &task, nil
}

func (r *repository) GetRandomTask(ctx context.Context, topic, difficulty, taskType, focus string) (*SpeakingTask, error) {
	filter := bson.M{}
	if topic != "" {
		filter["topic"] = topic
	}
	if difficulty != "" {
		filter["difficulty"] = difficulty
	}
	if taskType != "" {
		filter["type"] = taskType
	}
	if focus != "" {
		filter["focus"] = focus
	}

	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: filter}},
		{{Key: "$sample", Value: bson.D{{Key: "size", Value: 1}}}},
	}

	cursor, err := r.tasksColl.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var tasks []SpeakingTask
	if err := cursor.All(ctx, &tasks); err != nil {
		return nil, err
	}
	if len(tasks) == 0 {
		return nil, nil
	}
	return &tasks[0], nil
}

func (r *repository) CreateTask(ctx context.Context, task *SpeakingTask) error {
	if task.ID.IsZero() {
		task.ID = bson.NewObjectID()
	}
	_, err := r.tasksColl.InsertOne(ctx, task)
	return err
}

func (r *repository) UpdateTask(ctx context.Context, task *SpeakingTask) error {
	filter := bson.M{"_id": task.ID}
	update := bson.M{"$set": task}
	opts := options.UpdateOne().SetUpsert(true)
	_, err := r.tasksColl.UpdateOne(ctx, filter, update, opts)
	return err
}

// -----------------------------------------------------------------------------
// Sessions
// -----------------------------------------------------------------------------

func (r *repository) CreateSession(ctx context.Context, session *SpeakingSession) error {
	if session.ID.IsZero() {
		session.ID = bson.NewObjectID()
	}
	_, err := r.sessionsColl.InsertOne(ctx, session)
	return err
}

func (r *repository) UpdateSession(ctx context.Context, session *SpeakingSession) error {
	filter := bson.M{"_id": session.ID, "user_id": session.UserID} // Ownership check
	update := bson.M{"$set": session}
	
	result, err := r.sessionsColl.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return mongo.ErrNoDocuments
	}
	return nil
}

func (r *repository) GetSessionByID(ctx context.Context, userID, sessionID bson.ObjectID) (*SpeakingSession, error) {
	filter := bson.M{"_id": sessionID, "user_id": userID} // Ownership check
	var session SpeakingSession
	err := r.sessionsColl.FindOne(ctx, filter).Decode(&session)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &session, nil
}

func (r *repository) GetHistory(ctx context.Context, userID bson.ObjectID, limit int) ([]SpeakingHistoryItem, error) {
	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.D{{Key: "user_id", Value: userID}}}},
		{{Key: "$sort", Value: bson.D{{Key: "started_at", Value: -1}}}},
		{{Key: "$limit", Value: limit}},
		{{Key: "$lookup", Value: bson.D{
			{Key: "from", Value: "speaking_attempts"},
			{Key: "let", Value: bson.D{{Key: "session_id", Value: "$_id"}}},
			{Key: "pipeline", Value: mongo.Pipeline{
				{{Key: "$match", Value: bson.D{
					{Key: "$expr", Value: bson.D{
						{Key: "$eq", Value: bson.A{"$session_id", "$$session_id"}},
					}},
				}}},
				{{Key: "$project", Value: bson.D{
					{Key: "overall_score", Value: "$evaluation.overall"},
				}}},
			}},
			{Key: "as", Value: "attempts"},
		}}},
		{{Key: "$addFields", Value: bson.D{
			{Key: "best_score", Value: bson.D{{Key: "$max", Value: "$attempts.overall_score"}}},
			{Key: "attempts_count", Value: bson.D{{Key: "$size", Value: "$attempts"}}},
		}}},
		{{Key: "$project", Value: bson.D{
			{Key: "session", Value: "$$ROOT"},
			{Key: "best_score", Value: 1},
			{Key: "attempts_count", Value: 1},
		}}},
	}

	cursor, err := r.sessionsColl.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var items []SpeakingHistoryItem
	if err := cursor.All(ctx, &items); err != nil {
		return nil, err
	}

	return items, nil
}

// -----------------------------------------------------------------------------
// Attempts
// -----------------------------------------------------------------------------

func (r *repository) CreateAttempt(ctx context.Context, attempt *SpeakingAttempt) error {
	if attempt.ID.IsZero() {
		attempt.ID = bson.NewObjectID()
	}
	_, err := r.attemptsColl.InsertOne(ctx, attempt)
	return err
}

func (r *repository) UpdateAttempt(ctx context.Context, attempt *SpeakingAttempt) error {
	filter := bson.M{"_id": attempt.ID, "user_id": attempt.UserID} // Ownership check
	update := bson.M{"$set": attempt}

	result, err := r.attemptsColl.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return mongo.ErrNoDocuments
	}
	return nil
}

func (r *repository) GetAttempt(ctx context.Context, attemptID bson.ObjectID) (*SpeakingAttempt, error) {
	var attempt SpeakingAttempt
	err := r.attemptsColl.FindOne(ctx, bson.M{"_id": attemptID}).Decode(&attempt)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &attempt, nil
}

func (r *repository) GetAttemptByID(ctx context.Context, userID, attemptID bson.ObjectID) (*SpeakingAttempt, error) {
	filter := bson.M{"_id": attemptID, "user_id": userID} // Ownership check
	var attempt SpeakingAttempt
	err := r.attemptsColl.FindOne(ctx, filter).Decode(&attempt)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &attempt, nil
}

func (r *repository) GetAttemptsBySession(ctx context.Context, userID, sessionID bson.ObjectID) ([]SpeakingAttempt, error) {
	filter := bson.M{"session_id": sessionID, "user_id": userID} // Ownership check
	
	// Sort by created_at ascending so older attempts appear first
	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: 1}})
	
	cursor, err := r.attemptsColl.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var attempts []SpeakingAttempt
	if err := cursor.All(ctx, &attempts); err != nil {
		return nil, err
	}
	return attempts, nil
}

func (r *repository) GetRecentCompletedAttempts(ctx context.Context, userID bson.ObjectID, limit int) ([]SpeakingAttempt, error) {
	// Status should be completed and evaluation should exist
	filter := bson.M{
		"user_id": userID,
		"status":  AttemptCompleted,
		"evaluation": bson.M{"$exists": true},
	}
	
	opts := options.Find().
		SetSort(bson.D{{Key: "created_at", Value: -1}}).
		SetLimit(int64(limit))
		
	cursor, err := r.attemptsColl.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var attempts []SpeakingAttempt
	if err := cursor.All(ctx, &attempts); err != nil {
		return nil, err
	}
	return attempts, nil
}
