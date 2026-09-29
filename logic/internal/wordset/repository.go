package wordset

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

func (r *repository) CreateWordSet(ctx context.Context, set *WordSet) error {
	_, err := r.db.Collection("word_sets").InsertOne(ctx, set)
	if err != nil {
		return fmt.Errorf("wordset.CreateWordSet: %w", err)
	}
	return nil
}

func (r *repository) GetWordSetsByUserID(ctx context.Context, userID bson.ObjectID) ([]WordSet, error) {
	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}})
	cursor, err := r.db.Collection("word_sets").Find(ctx, bson.M{"user_id": userID}, opts)
	if err != nil {
		return nil, fmt.Errorf("wordset.GetWordSetsByUserID: %w", err)
	}
	defer cursor.Close(ctx)

	var sets []WordSet
	if err := cursor.All(ctx, &sets); err != nil {
		return nil, fmt.Errorf("wordset.GetWordSetsByUserID decode: %w", err)
	}
	return sets, nil
}

func (r *repository) GetWordSetByID(ctx context.Context, id bson.ObjectID) (*WordSet, error) {
	var set WordSet
	err := r.db.Collection("word_sets").FindOne(ctx, bson.M{"_id": id}).Decode(&set)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, fmt.Errorf("wordset.GetWordSetByID: %w", err)
	}
	return &set, nil
}

func (r *repository) UpdateWordSet(ctx context.Context, id bson.ObjectID, update bson.M) error {
	_, err := r.db.Collection("word_sets").UpdateOne(ctx, bson.M{"_id": id}, bson.M{"$set": update})
	if err != nil {
		return fmt.Errorf("wordset.UpdateWordSet: %w", err)
	}
	return nil
}

func (r *repository) ExploreWordSets(ctx context.Context, req ExploreWordSetsRequest) ([]WordSet, int, error) {
	filter := bson.M{"visibility": WordSetVisibilityPublic}

	if req.Query != "" {
		filter["$or"] = []bson.M{
			{"name": bson.M{"$regex": req.Query, "$options": "i"}},
			{"description": bson.M{"$regex": req.Query, "$options": "i"}},
		}
	}
	if req.Topic != "" {
		filter["topics"] = req.Topic
	}
	if req.Difficulty != "" {
		if req.Difficulty == string(DifficultyBeginner) {
			filter["difficulty"] = bson.M{"$in": []interface{}{req.Difficulty, "", nil}}
		} else {
			filter["difficulty"] = req.Difficulty
		}
	}

	totalCount, err := r.db.Collection("word_sets").CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("wordset.ExploreWordSets count: %w", err)
	}

	sortDoc := bson.D{{Key: "clone_count", Value: -1}, {Key: "_id", Value: -1}}
	if req.Sort == "newest" {
		sortDoc = bson.D{{Key: "created_at", Value: -1}, {Key: "_id", Value: -1}}
	}

	page := req.Page
	if page < 1 {
		page = 1
	}
	limit := req.Limit
	if limit < 1 || limit > 100 {
		limit = 20
	}
	skip := int64((page - 1) * limit)

	opts := options.Find().SetSort(sortDoc).SetSkip(skip).SetLimit(int64(limit))

	cursor, err := r.db.Collection("word_sets").Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, fmt.Errorf("wordset.ExploreWordSets find: %w", err)
	}
	defer cursor.Close(ctx)

	var sets []WordSet
	if err := cursor.All(ctx, &sets); err != nil {
		return nil, 0, fmt.Errorf("wordset.ExploreWordSets decode: %w", err)
	}

	if sets == nil {
		sets = []WordSet{}
	}

	return sets, int(totalCount), nil
}

func (r *repository) DeleteWordSetWithItems(ctx context.Context, id bson.ObjectID) error {
	// Simple two-step deletion. In a replica set, this could be a multi-document transaction.
	_, err := r.db.Collection("word_set_items").DeleteMany(ctx, bson.M{"set_id": id})
	if err != nil {
		return fmt.Errorf("wordset.DeleteWordSetWithItems items: %w", err)
	}
	_, err = r.db.Collection("word_sets").DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return fmt.Errorf("wordset.DeleteWordSetWithItems set: %w", err)
	}
	return nil
}

func (r *repository) CreateWordSetItem(ctx context.Context, item *WordSetItem) error {
	_, err := r.db.Collection("word_set_items").InsertOne(ctx, item)
	if err != nil {
		return fmt.Errorf("wordset.CreateWordSetItem: %w", err)
	}
	// Maintain count
	_, _ = r.db.Collection("word_sets").UpdateOne(ctx, bson.M{"_id": item.SetID}, bson.M{"$inc": bson.M{"word_count": 1}})
	return nil
}

func (r *repository) GetWordSetItems(ctx context.Context, setID bson.ObjectID) ([]WordSetItem, error) {
	opts := options.Find().SetSort(bson.D{{Key: "position", Value: 1}})
	cursor, err := r.db.Collection("word_set_items").Find(ctx, bson.M{"set_id": setID}, opts)
	if err != nil {
		return nil, fmt.Errorf("wordset.GetWordSetItems: %w", err)
	}
	defer cursor.Close(ctx)

	var items []WordSetItem
	if err := cursor.All(ctx, &items); err != nil {
		return nil, fmt.Errorf("wordset.GetWordSetItems decode: %w", err)
	}
	return items, nil
}

func (r *repository) RemoveWordSetItem(ctx context.Context, setID, wordID bson.ObjectID) error {
	res, err := r.db.Collection("word_set_items").DeleteMany(ctx, bson.M{"set_id": setID, "word_id": wordID})
	if err != nil {
		return fmt.Errorf("wordset.RemoveWordSetItem: %w", err)
	}
	if res.DeletedCount > 0 {
		_, _ = r.db.Collection("word_sets").UpdateOne(ctx, bson.M{"_id": setID}, bson.M{"$inc": bson.M{"word_count": -res.DeletedCount}})
	}
	return nil
}

// BatchRemoveWordSetItems deletes multiple word items in a single operation and decrements word_count atomically.
func (r *repository) BatchRemoveWordSetItems(ctx context.Context, setID bson.ObjectID, wordIDs []bson.ObjectID) (int64, error) {
	if len(wordIDs) == 0 {
		return 0, nil
	}
	filter := bson.M{
		"set_id":  setID,
		"word_id": bson.M{"$in": wordIDs},
	}
	res, err := r.db.Collection("word_set_items").DeleteMany(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("wordset.BatchRemoveWordSetItems: %w", err)
	}
	if res.DeletedCount > 0 {
		// Clamp: never let word_count go below 0
		_, _ = r.db.Collection("word_sets").UpdateOne(
			ctx,
			bson.M{"_id": setID},
			bson.M{"$inc": bson.M{"word_count": -res.DeletedCount}},
		)
	}
	return res.DeletedCount, nil
}

func (r *repository) ReorderWordSetItems(ctx context.Context, setID bson.ObjectID, wordIDs []bson.ObjectID) error {
	var models []mongo.WriteModel
	for i, wID := range wordIDs {
		filter := bson.M{"set_id": setID, "word_id": wID}
		update := bson.M{"$set": bson.M{"position": i}}
		models = append(models, mongo.NewUpdateOneModel().SetFilter(filter).SetUpdate(update))
	}

	if len(models) == 0 {
		return nil
	}

	_, err := r.db.Collection("word_set_items").BulkWrite(ctx, models)
	if err != nil {
		return fmt.Errorf("wordset.ReorderWordSetItems: %w", err)
	}
	return nil
}

func (r *repository) CreateWordSetItems(ctx context.Context, items []WordSetItem) (int, error) {
	if len(items) == 0 {
		return 0, nil
	}

	var docs []interface{}
	for _, it := range items {
		docs = append(docs, it)
	}

	opts := options.InsertMany().SetOrdered(false)
	res, err := r.db.Collection("word_set_items").InsertMany(ctx, docs, opts)

	insertedCount := 0
	if res != nil {
		insertedCount = len(res.InsertedIDs)
	}

	if err != nil {
		if bwe, ok := err.(mongo.BulkWriteException); ok {
			for _, we := range bwe.WriteErrors {
				if we.Code != 11000 { // 11000 is duplicate key error
					return insertedCount, fmt.Errorf("CreateWordSetItems bulk write error: %w", err)
				}
			}
			err = nil
		} else {
			return insertedCount, fmt.Errorf("CreateWordSetItems: %w", err)
		}
	}

	if insertedCount > 0 {
		_, _ = r.db.Collection("word_sets").UpdateOne(ctx, bson.M{"_id": items[0].SetID}, bson.M{"$inc": bson.M{"word_count": insertedCount}})
	}

	return insertedCount, err
}

func (r *repository) GetWordSetsContainingWord(ctx context.Context, userID bson.ObjectID, wordIDs []bson.ObjectID) ([]WordSetItem, error) {
	if len(wordIDs) == 0 {
		return []WordSetItem{}, nil
	}
	cursor, err := r.db.Collection("word_sets").Find(ctx, bson.M{"user_id": userID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var userSets []WordSet
	if err := cursor.All(ctx, &userSets); err != nil {
		return nil, err
	}

	if len(userSets) == 0 {
		return []WordSetItem{}, nil
	}

	var setIDs []bson.ObjectID
	for _, s := range userSets {
		setIDs = append(setIDs, s.ID)
	}

	filter := bson.M{
		"set_id":  bson.M{"$in": setIDs},
		"word_id": bson.M{"$in": wordIDs},
	}

	itemCursor, err := r.db.Collection("word_set_items").Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer itemCursor.Close(ctx)

	var items []WordSetItem
	if err := itemCursor.All(ctx, &items); err != nil {
		return nil, err
	}

	return items, nil
}

func (r *repository) VerifyWordsExist(ctx context.Context, wordIDs []bson.ObjectID) (int, error) {
	if len(wordIDs) == 0 {
		return 0, nil
	}
	filter := bson.M{"_id": bson.M{"$in": wordIDs}}
	count, err := r.db.Collection("words").CountDocuments(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("wordset.VerifyWordsExist: %w", err)
	}
	return int(count), nil
}

func (r *repository) EnsureIndexes(ctx context.Context) error {
	// word_sets indexes
	wsIndex := []mongo.IndexModel{
		{
			Keys: bson.D{{Key: "user_id", Value: 1}, {Key: "created_at", Value: -1}},
		},
	}
	if _, err := r.db.Collection("word_sets").Indexes().CreateMany(ctx, wsIndex); err != nil {
		return fmt.Errorf("wordset.EnsureIndexes (word_sets): %w", err)
	}

	// word_set_items indexes
	wsiIndex := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "set_id", Value: 1}, {Key: "word_id", Value: 1}, {Key: "sense_id", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: bson.D{{Key: "set_id", Value: 1}, {Key: "position", Value: 1}},
		},
	}
	if _, err := r.db.Collection("word_set_items").Indexes().CreateMany(ctx, wsiIndex); err != nil {
		return fmt.Errorf("wordset.EnsureIndexes (word_set_items): %w", err)
	}

	return nil
}

func (r *repository) UpdateWordSetItemAnnotation(ctx context.Context, setID, wordID bson.ObjectID, annotation *UserAnnotation) error {
	filter := bson.M{
		"set_id":  setID,
		"word_id": wordID,
	}
	var update bson.M
	if annotation == nil {
		update = bson.M{"$unset": bson.M{"annotation": ""}}
	} else {
		update = bson.M{"$set": bson.M{"annotation": annotation}}
	}

	_, err := r.db.Collection("word_set_items").UpdateOne(ctx, filter, update)
	return err
}
