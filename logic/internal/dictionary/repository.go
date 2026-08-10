package dictionary

import (
	"context"
	"crypto/md5"

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

func (r *repository) SearchPrefix(ctx context.Context, prefix string, limit int) ([]WordDocument, error) {
	if limit > 20 {
		limit = 20
	}
	if limit <= 0 {
		limit = 10
	}

	// Autocomplete via regex prefix search (anchored to start)
	filter := bson.M{
		"word": bson.M{"$regex": "^" + prefix, "$options": "i"},
	}

	opts := options.Find().SetLimit(int64(limit)).SetSort(bson.D{{Key: "word", Value: 1}})
	
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

func (r *repository) GetRandomWord(ctx context.Context) (*WordDocument, error) {
	pipeline := mongo.Pipeline{
		{{Key: "$sample", Value: bson.D{{Key: "size", Value: 1}}}},
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
	return &docs[0], nil
}

func (r *repository) GetWordOfTheDay(ctx context.Context, dateStr string) (*WordDocument, error) {
	// Deterministic approach: hash the date string, create a pseudo-random ObjectID,
	// and find the first document with _id >= this ObjectID.
	hash := md5.Sum([]byte(dateStr))
	// An ObjectID is 12 bytes. We can just use the first 12 bytes of the MD5 hash.
	var oid [12]byte
	copy(oid[:], hash[:12])

	objectID := bson.ObjectID(oid)

	// First try to find a document with _id >= objectID
	filter := bson.M{"_id": bson.M{"$gte": objectID}}
	opts := options.FindOne().SetSort(bson.D{{Key: "_id", Value: 1}})

	var doc WordDocument
	err := r.coll.FindOne(ctx, filter, opts).Decode(&doc)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			// If we generated an ID larger than the max ID in the collection, wrap around.
			optsWrap := options.FindOne().SetSort(bson.D{{Key: "_id", Value: 1}})
			err = r.coll.FindOne(ctx, bson.M{}, optsWrap).Decode(&doc)
			if err != nil {
				return nil, err
			}
			return &doc, nil
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
		{{Key: "$sample", Value: bson.D{{Key: "size", Value: 200}}}}, // Sample 200 random words first for O(1) performance
		{{Key: "$match", Value: filter}},
		{{Key: "$limit", Value: n}},
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

