package main

import (
	"context"
	"fmt"
	"log"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func main() {
	client, err := mongo.Connect(options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(context.Background())

	coll := client.Database("dictionary").Collection("words")
	
	excludeID, _ := bson.ObjectIDFromHex("6a462b4196e73b4df477109c") // boldenone undecylenate
	pos := "noun"

	filter := bson.D{
		{Key: "pos", Value: pos},
		{Key: "_id", Value: bson.M{"$ne": excludeID}},
		{Key: "senses.0.glosses.0", Value: bson.M{"$exists": true}},
	}
	pipeline := mongo.Pipeline{
		{{Key: "$sample", Value: bson.D{{Key: "size", Value: 200}}}},
		{{Key: "$match", Value: filter}},
		{{Key: "$limit", Value: 3}},
		{{Key: "$project", Value: bson.D{
			{Key: "word", Value: 1},
			{Key: "pos", Value: 1},
			{Key: "senses", Value: bson.D{{Key: "$slice", Value: bson.A{"$senses", 1}}}},
		}}},
	}

	cursor, err := coll.Aggregate(context.Background(), pipeline)
	if err != nil {
		log.Fatal("Aggregate error:", err)
	}
	defer cursor.Close(context.Background())

	var docs []bson.M
	if err := cursor.All(context.Background(), &docs); err != nil {
		log.Fatal("Cursor error:", err)
	}

	fmt.Printf("Found %d distractors\n", len(docs))
}
