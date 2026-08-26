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

	db := client.Database("english_dev")
	col := db.Collection("user_profiles")

	var profiles []bson.M
	cursor, err := col.Find(context.Background(), bson.M{})
	if err != nil {
		log.Fatal(err)
	}
	cursor.All(context.Background(), &profiles)

	for _, p := range profiles {
		skills, ok := p["skills"].(bson.A)
		if !ok {
			continue
		}

		seen := make(map[string]bool)
		var newSkills bson.A

		for _, s := range skills {
			if skillDoc, ok := s.(bson.M); ok {
				skillName := skillDoc["skill"].(string)
				if !seen[skillName] {
					seen[skillName] = true
					newSkills = append(newSkills, s)
				}
			}
		}

		if len(newSkills) != len(skills) {
			_, err := col.UpdateOne(context.Background(), bson.M{"_id": p["_id"]}, bson.M{"$set": bson.M{"skills": newSkills}})
			if err != nil {
				log.Fatal(err)
			}
			fmt.Printf("Deduplicated skills for user %v\n", p["user_id"])
		}
	}
}
