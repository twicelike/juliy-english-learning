package main

import (
	"context"
	"log"

	"github.com/english/backend/internal/config"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	client, err := mongo.Connect(options.Client().ApplyURI(cfg.MongoURI))
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer client.Disconnect(context.Background())

	db := client.Database(cfg.MongoDB)

	log.Println("Dropping legacy exercise_contents collection...")
	err = db.Collection("exercise_contents").Drop(context.Background())
	if err != nil {
		log.Printf("Error dropping collection (might not exist): %v", err)
	} else {
		log.Println("Successfully dropped exercise_contents collection.")
	}
}
