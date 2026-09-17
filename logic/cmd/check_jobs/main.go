package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/english/backend/internal/config"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	clientOptions := options.Client().ApplyURI(cfg.MongoURI)
	mongoClient, err := mongo.Connect(clientOptions)
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer mongoClient.Disconnect(ctx)

	dictColl := mongoClient.Database("dictionary").Collection("words")

	validPOS := bson.A{"noun", "verb", "adjective", "adverb", "phrasal verb"}

	// Thresholds to check
	thresholds := []int{1000, 3000, 5000, 10000}

	for _, n := range thresholds {
		cursor, err := dictColl.Find(ctx,
			bson.M{"pos": bson.M{"$in": validPOS}},
			options.Find().
				SetSort(bson.D{{Key: "frequency_score", Value: -1}}).
				SetLimit(int64(n)).
				SetProjection(bson.M{"_id": 1, "word": 1, "senses.quality": 1}),
		)
		if err != nil {
			log.Fatal(err)
		}

		var docs []struct {
			ID     bson.ObjectID `bson:"_id"`
			Word   string        `bson:"word"`
			Senses []struct {
				Quality *struct{} `bson:"quality,omitempty"`
			} `bson:"senses"`
		}

		if err := cursor.All(ctx, &docs); err != nil {
			log.Fatal(err)
		}
		cursor.Close(ctx)

		noQuality := 0
		for _, doc := range docs {
			hasQuality := false
			for _, s := range doc.Senses {
				if s.Quality != nil {
					hasQuality = true
					break
				}
			}
			if !hasQuality {
				noQuality++
			}
		}

		pct := float64(noQuality) / float64(len(docs)) * 100
		fmt.Printf("Top %-6d words: %d / %d chưa được enrich (%.1f%%)\n", n, noQuality, len(docs), pct)
	}
}
