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

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	clientOptions := options.Client().ApplyURI(cfg.MongoURI)
	mongoClient, err := mongo.Connect(clientOptions)
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer mongoClient.Disconnect(ctx)

	dictColl := mongoClient.Database("dictionary").Collection("words")

	// case-insensitive match
	cursor, err := dictColl.Find(ctx, bson.M{"word": bson.M{"$regex": "(?i)^praemunire$"}})
	if err != nil {
		log.Fatal(err)
	}
	var docs []bson.M
	cursor.All(ctx, &docs)
	for _, doc := range docs {
		fmt.Printf("ID: %v, Word: '%v', POS: '%v'\n", doc["_id"], doc["word"], doc["pos"])
	}
}
