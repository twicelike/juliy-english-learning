package main

import (
	"context"
	"fmt"
	"log"

	"github.com/english/backend/internal/dictionary"
	"github.com/english/backend/internal/learning"
	"github.com/english/backend/internal/user"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func main() {
	client, err := mongo.Connect(options.Client().ApplyURI("mongodb://127.0.0.1:27017"))
	if err != nil {
		log.Fatal(err)
	}

	dbDict := client.Database("dictionary")

	dictRepo := dictionary.NewRepository(dbDict)
	userRepo := user.NewRepository(dbDict)
	learnRepo := learning.NewRepository(dbDict)

	service := learning.NewService(learnRepo, dictRepo, userRepo, learning.NewSM2Scheduler())

	// Need a valid user ID. Let's find one.
	var u bson.M
	err = dbDict.Collection("users").FindOne(context.Background(), bson.M{}).Decode(&u)
	if err != nil {
		log.Fatal("No users found")
	}

	id := u["_id"].(bson.ObjectID)
	fmt.Printf("User: %s\n", id.Hex())

	session, err := service.GetSession(context.Background(), id.Hex(), learning.ModeQuiz, 10)
	if err != nil {
		log.Fatalf("Error GetSession: %v", err)
	}

	fmt.Printf("Session items: %d\n", len(session.Items))
	for _, item := range session.Items {
		fmt.Printf("- %s (choices: %d)\n", item.Word, len(item.Choices))
	}
}
