package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/english/backend/internal/config"
	"github.com/english/backend/internal/dictionary"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Hour)
	defer cancel()

	clientOptions := options.Client().ApplyURI(cfg.MongoURI)
	mongoClient, err := mongo.Connect(clientOptions)
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer mongoClient.Disconnect(ctx)

	dictDB := mongoClient.Database("dictionary")
	coll := dictDB.Collection("words")

	log.Println("Starting SenseID Migration (Idempotent)...")

	// Statistics
	var totalWords int
	var updatedDocuments int
	var beforeCount int
	var missingIdCount int
	var assignedCount int
	var alreadyExistingCount int
	var invalidIdCount int
	var duplicateIdCount int

	// For duplicate detection
	globalSenseIDs := make(map[string]bson.ObjectID) // senseID -> wordID (first seen)

	cursor, err := coll.Find(ctx, bson.M{})
	if err != nil {
		log.Fatalf("Failed to fetch words: %v", err)
	}
	defer cursor.Close(ctx)

	// We will batch updates for performance if needed, but for safety and reporting
	// we can do it document by document since this is a one-time migration.
	// For 430k words, doing it one by one might take a few minutes, which is fine.

	for cursor.Next(ctx) {
		totalWords++
		var doc dictionary.WordDocument
		if err := cursor.Decode(&doc); err != nil {
			log.Printf("Failed to decode document: %v", err)
			continue
		}

		changed := false
		for i := range doc.Senses {
			beforeCount++
			sense := &doc.Senses[i]

			if sense.ID == "" {
				missingIdCount++
				// Generate new UUID
				newID := uuid.NewString()
				sense.ID = newID
				assignedCount++
				changed = true
			} else {
				alreadyExistingCount++
				// Basic validate UUID (just checking length for now, or parsing if needed)
				_, err := uuid.Parse(sense.ID)
				if err != nil {
					invalidIdCount++
					log.Printf("WordID %s: Invalid SenseID format '%s'", doc.ID.Hex(), sense.ID)
				}
			}

			// Duplicate Check
			if existingWordID, exists := globalSenseIDs[sense.ID]; exists {
				duplicateIdCount++
				log.Printf("CRITICAL: Duplicate SenseID '%s' found in Word %s and Word %s", sense.ID, doc.ID.Hex(), existingWordID.Hex())
			} else {
				globalSenseIDs[sense.ID] = doc.ID
			}
		}

		if changed {
			filter := bson.M{"_id": doc.ID}
			update := bson.M{"$set": bson.M{"senses": doc.Senses}}
			_, err := coll.UpdateOne(ctx, filter, update)
			if err != nil {
				log.Printf("Failed to update Word %s: %v", doc.ID.Hex(), err)
			} else {
				updatedDocuments++
			}
		}

		if totalWords%10000 == 0 {
			log.Printf("Processed %d words...", totalWords)
		}
	}

	if err := cursor.Err(); err != nil {
		log.Fatalf("Cursor error: %v", err)
	}

	fmt.Println("\n==================================================")
	fmt.Println("MIGRATION REPORT")
	fmt.Println("==================================================")
	fmt.Printf("Total Words Scanned:      %d\n", totalWords)
	fmt.Printf("Updated Documents:        %d\n", updatedDocuments)
	fmt.Println("--------------------------------------------------")
	fmt.Printf("Total Senses (Before):    %d\n", beforeCount)
	fmt.Printf("Already Existing IDs:     %d\n", alreadyExistingCount)
	fmt.Printf("Missing IDs Found:        %d\n", missingIdCount)
	fmt.Printf("New IDs Assigned:         %d\n", assignedCount)
	fmt.Println("--------------------------------------------------")
	fmt.Printf("Invalid IDs Detected:     %d\n", invalidIdCount)
	fmt.Printf("Duplicate IDs Detected:   %d\n", duplicateIdCount)
	fmt.Println("==================================================")

	if missingIdCount == assignedCount && duplicateIdCount == 0 {
		fmt.Println("SUCCESS: Invariants satisfied. All senses have unique valid UUIDs.")
	} else {
		fmt.Println("WARNING: Migration finished with anomalies. Please review logs.")
	}
}
