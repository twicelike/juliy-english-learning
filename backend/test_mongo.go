package main

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(options.Client().ApplyURI("mongodb://127.0.0.1:27017"))
	if err != nil {
		panic(err)
	}
	defer client.Disconnect(ctx)

	coll := client.Database("dictionary").Collection("words")

	count, err := coll.CountDocuments(ctx, bson.M{})
	fmt.Printf("Count: %d, err: %v\n", count, err)

	var doc bson.M
	err = coll.FindOne(ctx, bson.M{"word": "free"}).Decode(&doc)
	fmt.Printf("FindOne(free) err: %v\n", err)
	if err == nil {
		fmt.Printf("Doc: %v\n", doc["word"])
	}

	type WordDoc struct {
		ID   bson.ObjectID `bson:"_id"`
		Word string        `bson:"word"`
	}
	var typedDoc WordDoc
	err = coll.FindOne(ctx, bson.M{"word": "free"}).Decode(&typedDoc)
	fmt.Printf("Decode Typed err: %v\n", err)
}
