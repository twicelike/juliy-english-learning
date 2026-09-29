package dictionary

import (
	"context"
	"crypto/md5"
	"fmt"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type repository struct {
	coll *mongo.Collection
}

// NewRepository creates a new dictionary repository.
func NewRepository(db *mongo.Database) Repository {
	return &repository{
		coll: db.Collection("words"),
	}
}

// EnsureIndexes creates indexes needed for efficient queries.
func (r *repository) EnsureIndexes(ctx context.Context) error {
	// Compound index: covers GetTopWordsNotIn (sort by frequency_score, filter by pos)
	// frequency_score must be first to avoid in-memory sorting over $in array.
	_, err := r.coll.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{
			{Key: "frequency_score", Value: -1},
			{Key: "pos", Value: 1},
		},
		Options: options.Index().SetName("frequency_score_pos"),
	})
	if err != nil {
		return fmt.Errorf("dictionary.EnsureIndexes (pos_frequency_score): %w", err)
	}
	return nil
}

func (r *repository) SearchPrefix(ctx context.Context, prefix string, limit int) ([]WordDocument, error) {
	if limit > 50 {
		limit = 50
	}
	if limit <= 0 {
		limit = 10
	}

	// Autocomplete via regex prefix search (anchored to start)
	regex := bson.M{"$regex": "^" + prefix, "$options": "i"}
	filter := bson.M{
		"$or": []bson.M{
			{"word": regex},
			{"forms.form": regex},
		},
	}

	opts := options.Find().
		SetLimit(int64(limit)).
		SetSort(bson.D{{Key: "word", Value: 1}}).
		SetCollation(&options.Collation{Locale: "en", Strength: 2})

	// We only need the 'word' field for search suggestions, but the schema has more.
	// The plan mentions returning WordDTOs, so we'll fetch the whole doc or a projection.
	// For now, let's fetch the full document to map to DTO in service.
	cursor, err := r.coll.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var words []WordDocument
	if err := cursor.All(ctx, &words); err != nil {
		return nil, err
	}

	return words, nil
}

func (r *repository) GetWord(ctx context.Context, word string) (*WordDocument, error) {
	filter := bson.M{"word": word}
	var doc WordDocument
	err := r.coll.FindOne(ctx, filter).Decode(&doc)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil // Not found
		}
		return nil, err
	}
	return &doc, nil
}

func (r *repository) GetWords(ctx context.Context, word string) ([]WordDocument, error) {
	filter := bson.M{"word": word}
	cursor, err := r.coll.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var docs []WordDocument
	if err := cursor.All(ctx, &docs); err != nil {
		return nil, err
	}
	return docs, nil
}

func (r *repository) GetTopWordsNotIn(ctx context.Context, excludeIDs []bson.ObjectID, limit int) ([]WordDocument, error) {
	validPOS := []string{"noun", "verb", "adjective", "adverb", "phrasal verb"}
	
	filter := bson.M{
		"pos": bson.M{"$in": validPOS},
	}
	
	if len(excludeIDs) > 0 {
		filter["_id"] = bson.M{"$nin": excludeIDs}
	}
	// Sort by top-level frequency_score (denormalized from senses[0].quality.frequency_score)
	// This field is indexed via {pos, frequency_score} compound index.
	opts := options.Find().
		SetLimit(int64(limit)).
		SetSort(bson.D{{Key: "frequency_score", Value: -1}})

	cursor, err := r.coll.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var docs []WordDocument
	if err := cursor.All(ctx, &docs); err != nil {
		return nil, err
	}
	return docs, nil
}

func (r *repository) GetRandomWord(ctx context.Context) (*WordDocument, error) {
	// MongoDB's $sample is O(1) ONLY if it's the first stage.
	// If it follows a $match, it forces a collection or index scan of all matching documents.
	// To avoid this slow query, we sample 10 random words first, then filter out 'name' in-memory.
	pipeline := mongo.Pipeline{
		{{Key: "$sample", Value: bson.D{{Key: "size", Value: 10}}}},
	}
	cursor, err := r.coll.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var docs []WordDocument
	if err := cursor.All(ctx, &docs); err != nil {
		return nil, err
	}
	
	if len(docs) == 0 {
		return nil, nil
	}

	for _, doc := range docs {
		if doc.POS != "name" {
			return &doc, nil
		}
	}
	
	// Fallback to the first one if all happen to be names (extremely rare)
	return &docs[0], nil
}

func (r *repository) GetWordOfTheDay(ctx context.Context, dateStr string) (*WordDocument, error) {
	// 1. Hash the date string to get a deterministic integer
	hash := md5.Sum([]byte(dateStr))
	var hashInt uint64
	for i := 0; i < 8; i++ {
		hashInt = (hashInt << 8) | uint64(hash[i])
	}

	// 2. Modulo by 5000 to pick a deterministic offset from the top 5000 most frequent words.
	// This avoids a massive O(N) CountDocuments and O(N) Skip, dropping latency from >1000ms to ~2ms.
	// It also ensures the "Word of the Day" is actually a useful, common word.
	offset := int64(hashInt % 5000)

	filter := bson.M{"pos": bson.M{"$in": []string{"noun", "verb", "adjective", "adverb"}}}
	
	// 3. Find with skip using the frequency_score index
	opts := options.FindOne().
		SetSort(bson.D{{Key: "frequency_score", Value: -1}}).
		SetSkip(offset)
		
	var doc WordDocument
	err := r.coll.FindOne(ctx, filter, opts).Decode(&doc)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}

	return &doc, nil
}

func (r *repository) GetByID(ctx context.Context, id bson.ObjectID) (*WordDocument, error) {
	var doc WordDocument
	err := r.coll.FindOne(ctx, bson.M{"_id": id}).Decode(&doc)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &doc, nil
}

func (r *repository) GetManyByIDs(ctx context.Context, ids []bson.ObjectID) ([]WordDocument, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	filter := bson.M{"_id": bson.M{"$in": ids}}
	cursor, err := r.coll.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	var docs []WordDocument
	if err := cursor.All(ctx, &docs); err != nil {
		return nil, err
	}
	return docs, nil
}

func (r *repository) GetManyByWords(ctx context.Context, words []string) ([]WordDocument, error) {
	if len(words) == 0 {
		return nil, nil
	}
	filter := bson.M{"word": bson.M{"$in": words}}
	cursor, err := r.coll.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	var docs []WordDocument
	if err := cursor.All(ctx, &docs); err != nil {
		return nil, err
	}
	return docs, nil
}

// GetDistractors returns n random words with the same POS, excluding the given _id.
// Results are projected to only include the first sense to minimise data transfer.
func (r *repository) GetDistractors(ctx context.Context, pos string, excludeID bson.ObjectID, n int) ([]WordDocument, error) {
	filter := bson.D{
		{Key: "pos", Value: pos},
		{Key: "_id", Value: bson.M{"$ne": excludeID}},
		{Key: "senses.0.glosses.0", Value: bson.M{"$exists": true}},
	}
	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: filter}},
		{{Key: "$sample", Value: bson.D{{Key: "size", Value: n}}}},
		{{Key: "$project", Value: bson.D{
			{Key: "word", Value: 1},
			{Key: "pos", Value: 1},
			{Key: "senses", Value: bson.D{{Key: "$slice", Value: bson.A{"$senses", 1}}}},
		}}},
	}
	cursor, err := r.coll.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	var docs []WordDocument
	if err := cursor.All(ctx, &docs); err != nil {
		return nil, err
	}
	return docs, nil
}

// GetDistractorsBulk returns n random words with the same POS, excluding the given excludeIDs.
func (r *repository) GetDistractorsBulk(ctx context.Context, pos string, excludeIDs []bson.ObjectID, n int) ([]WordDocument, error) {
	if len(excludeIDs) == 0 {
		excludeIDs = []bson.ObjectID{bson.NewObjectID()}
	}

	filter := bson.D{
		{Key: "pos", Value: pos},
		{Key: "_id", Value: bson.M{"$nin": excludeIDs}},
		{Key: "senses.0.glosses.0", Value: bson.M{"$exists": true}},
	}
	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: filter}},
		{{Key: "$sample", Value: bson.D{{Key: "size", Value: n}}}},
		{{Key: "$project", Value: bson.D{
			{Key: "word", Value: 1},
			{Key: "pos", Value: 1},
			{Key: "senses", Value: bson.D{{Key: "$slice", Value: bson.A{"$senses", 1}}}},
		}}},
	}
	cursor, err := r.coll.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	var docs []WordDocument
	if err := cursor.All(ctx, &docs); err != nil {
		return nil, err
	}
	return docs, nil
}
