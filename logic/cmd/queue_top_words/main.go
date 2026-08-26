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
	transRepo := translation.NewRepository(transDB)
	transSvc := translation.NewService(nil, transRepo)

	// Clear existing pending queue
	_, err = transDB.Collection("translation_jobs").DeleteMany(ctx, bson.M{"status": "pending"})
	if err != nil {
		log.Fatalf("Failed to clear queue: %v", err)
	}
	fmt.Println("Cleared existing pending jobs.")

	// Get Top 100 words
	topWords, err := dictRepo.GetTopWordsNotIn(ctx, nil, 100)
	if err != nil {
		log.Fatalf("Failed to get top words: %v", err)
	}

	var batch []string
	for _, w := range topWords {
		batch = append(batch, w.ID.Hex())
	}

	err = transSvc.EnqueueJobs(ctx, batch, "vi")
	if err != nil {
		log.Fatalf("Failed to enqueue: %v", err)
	}

	fmt.Printf("Enqueued TOP %d words! Translation worker will process them in a few seconds.\n", len(batch))
}
