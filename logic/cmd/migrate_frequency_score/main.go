// cmd/migrate_frequency_score/main.go
// Backfills the top-level frequency_score field on each word document by
// copying the value from senses[0].quality.frequency_score.
//
// Usage:
//
//	go run ./cmd/migrate_frequency_score
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func main() {
	_ = godotenv.Load(".env", "../.env", "../../.env")

	mongoURI := os.Getenv("MONGO_URI")
	if mongoURI == "" {
		mongoURI = "mongodb://127.0.0.1:27017"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	client, err := mongo.Connect(options.Client().ApplyURI(mongoURI))
	if err != nil {
		log.Fatalf("connect: %v", err)
	}
	defer client.Disconnect(ctx)

	coll := client.Database("dictionary").Collection("words")

	// Use aggregation pipeline update to copy the nested field to top-level.
	// IMPORTANT: In MongoDB aggregation, "$senses.0.quality.frequency_score" returns
	// an ARRAY (all elements), not the first element. Must use $arrayElemAt.
	pipeline := mongo.Pipeline{
		{{Key: "$set", Value: bson.D{
			{Key: "frequency_score", Value: bson.D{
				{Key: "$ifNull", Value: bson.A{
					bson.D{{Key: "$arrayElemAt", Value: bson.A{
						"$senses.quality.frequency_score",
						0,
					}}},
					0,
				}},
			}},
		}}},
	}

	result, err := coll.UpdateMany(ctx, bson.M{}, pipeline)
	if err != nil {
		log.Fatalf("updateMany: %v", err)
	}

	fmt.Printf("✅ Migration complete: matched=%d, modified=%d\n", result.MatchedCount, result.ModifiedCount)
}
