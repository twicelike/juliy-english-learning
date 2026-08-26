package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/english/backend/internal/config"
	"github.com/english/backend/internal/dictionary"
	"github.com/english/backend/internal/translation"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// This script enqueues all words in the dictionary for translation to a target language.
func main() {
	cfg, _ := config.Load()

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Hour)
	defer cancel()

	clientOptions := options.Client().ApplyURI(cfg.MongoURI)
	mongoClient, err := mongo.Connect(clientOptions)
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer mongoClient.Disconnect(ctx)

	dictDB := mongoClient.Database("dictionary")
	transDB := mongoClient.Database(cfg.MongoDB)

	dictRepo := dictionary.NewRepository(dictDB)
	_ = dictRepo
	transRepo := translation.NewRepository(transDB)

	// Admin CLI configuration
	targetLang := "vi"
	batchSize := 1000

	log.Printf("Starting bulk enqueue for language: %s", targetLang)

	// Because dictionary might have ~430k words, we use a simple cursor or pagination.
	// We'll query raw collection to avoid memory overload.
	coll := dictDB.Collection("words")
	cursor, err := coll.Find(ctx, bson.M{})
	if err != nil {
		log.Fatalf("Failed to fetch words: %v", err)
	}
	defer cursor.Close(ctx)

	var batch []string
	totalEnqueued := 0

	// We create a dummy service just to use EnqueueJobs
	// EnqueueJobs will cleanly handle idempotent inserts.
	// Since we don't need AI for enqueuing, we can pass nil to the client.
	transSvc := translation.NewService(nil, transRepo)

	for cursor.Next(ctx) {
		var doc struct {
			ID bson.ObjectID `bson:"_id"`
		}
		if err := cursor.Decode(&doc); err != nil {
			continue
		}

		batch = append(batch, doc.ID.Hex())

		if len(batch) >= batchSize {
			err := transSvc.EnqueueJobs(ctx, batch, targetLang)
			if err != nil {
				log.Printf("Failed to enqueue batch: %v", err)
			} else {
				totalEnqueued += len(batch)
				fmt.Printf("\rEnqueued %d words...", totalEnqueued)
			}
			batch = batch[:0] // reset
		}
	}

	// Enqueue remaining
	if len(batch) > 0 {
		err := transSvc.EnqueueJobs(ctx, batch, targetLang)
		if err != nil {
			log.Printf("Failed to enqueue final batch: %v", err)
		} else {
			totalEnqueued += len(batch)
		}
	}

	fmt.Printf("\nDone! Total words processed/enqueued: %d\n", totalEnqueued)
	log.Printf("The background worker will pick these up automatically.")
}
