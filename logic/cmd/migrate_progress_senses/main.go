package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/english/backend/internal/config"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	ctx := context.Background()
	clientOptions := options.Client().ApplyURI(cfg.MongoURI)
	mongoClient, err := mongo.Connect(clientOptions)
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer mongoClient.Disconnect(ctx)

	db := mongoClient.Database("dictionary")
	learningDB := mongoClient.Database("learning")

	log.Println("Starting Phase 4 Data Migration...")

	// 1. Build WordID -> []SenseID map
	log.Println("Building SenseID lookup map from words collection...")
	wordsColl := db.Collection("words")
	cursor, err := wordsColl.Find(ctx, bson.M{}, options.Find().SetProjection(bson.M{"senses.id": 1}))
	if err != nil {
		log.Fatalf("Failed to fetch words: %v", err)
	}

	wordSenseMap := make(map[string][]string)
	for cursor.Next(ctx) {
		var doc struct {
			ID     bson.ObjectID `bson:"_id"`
			Senses []struct {
				ID string `bson:"id"`
			} `bson:"senses"`
		}
		if err := cursor.Decode(&doc); err != nil {
			log.Fatalf("Failed to decode word doc: %v", err)
		}
		senseIDs := make([]string, len(doc.Senses))
		for i, s := range doc.Senses {
			senseIDs[i] = s.ID
		}
		wordSenseMap[doc.ID.Hex()] = senseIDs
	}
	cursor.Close(ctx)
	log.Printf("Loaded %d words into memory.", len(wordSenseMap))

	// Helper to resolve sense ID
	resolveSenseID := func(wordID bson.ObjectID, senseIndex int) (string, error) {
		senses, ok := wordSenseMap[wordID.Hex()]
		if !ok {
			return "", fmt.Errorf("word_id not found in words collection")
		}
		if senseIndex < 0 || senseIndex >= len(senses) {
			return "", fmt.Errorf("sense_index out of range (index=%d, len=%d)", senseIndex, len(senses))
		}
		if senses[senseIndex] == "" {
			return "", fmt.Errorf("sense_id is empty at index %d", senseIndex)
		}
		return senses[senseIndex], nil
	}

	hasCriticalErrors := false

	// 2. Migrate learning_progress
	log.Println("Migrating learning_progress...")
	progressColl := learningDB.Collection("learning_progress")
	pCursor, err := progressColl.Find(ctx, bson.M{})
	if err != nil {
		log.Fatalf("Failed to query learning_progress: %v", err)
	}

	var pTotal, pMigrated, pSkipped, pFailed int
	var pBulkOps []mongo.WriteModel

	for pCursor.Next(ctx) {
		pTotal++
		var doc struct {
			ID         bson.ObjectID `bson:"_id"`
			WordID     bson.ObjectID `bson:"word_id"`
			SenseIndex *int          `bson:"sense_index"`
			SenseID    *string       `bson:"sense_id"`
		}
		if err := pCursor.Decode(&doc); err != nil {
			log.Printf("[MIGRATION ERROR] learning_progress _id=%s decode failed: %v", doc.ID.Hex(), err)
			pFailed++
			hasCriticalErrors = true
			continue
		}

		if doc.SenseID != nil && *doc.SenseID != "" {
			pSkipped++ // Already migrated
			continue
		}

		if doc.SenseIndex == nil {
			log.Printf("[MIGRATION ERROR] learning_progress _id=%s missing both sense_id and sense_index", doc.ID.Hex())
			pFailed++
			hasCriticalErrors = true
			continue
		}

		senseID, err := resolveSenseID(doc.WordID, *doc.SenseIndex)
		if err != nil {
			log.Printf("[MIGRATION ERROR] learning_progress word_id=%s sense_index=%d reason=%v", doc.WordID.Hex(), *doc.SenseIndex, err)
			pFailed++
			hasCriticalErrors = true
			continue
		}

		update := mongo.NewUpdateOneModel().
			SetFilter(bson.M{"_id": doc.ID}).
			SetUpdate(bson.M{"$set": bson.M{"sense_id": senseID}, "$unset": bson.M{"sense_index": ""}})
		pBulkOps = append(pBulkOps, update)
	}
	pCursor.Close(ctx)

	if len(pBulkOps) > 0 {
		if _, err := progressColl.BulkWrite(ctx, pBulkOps, options.BulkWrite().SetOrdered(false)); err != nil {
			log.Fatalf("Failed to bulk update learning_progress: %v", err)
		}
		pMigrated = len(pBulkOps)
	}

	// 3. Migrate word_sets
	log.Println("Migrating word_sets...")
	wordSetsColl := learningDB.Collection("word_sets")
	wCursor, err := wordSetsColl.Find(ctx, bson.M{})
	if err != nil {
		log.Fatalf("Failed to query word_sets: %v", err)
	}

	var wTotal, wMigrated, wSkipped, wFailed int
	var wBulkOps []mongo.WriteModel

	for wCursor.Next(ctx) {
		wTotal++
		var doc struct {
			ID    bson.ObjectID `bson:"_id"`
			Items []struct {
				WordID     bson.ObjectID `bson:"word_id"`
				SenseIndex *int          `bson:"sense_index"`
				SenseID    *string       `bson:"sense_id"`
				AddedAt    interface{}   `bson:"added_at"`
			} `bson:"items"`
		}
		if err := wCursor.Decode(&doc); err != nil {
			log.Printf("[MIGRATION ERROR] word_sets _id=%s decode failed: %v", doc.ID.Hex(), err)
			wFailed++
			hasCriticalErrors = true
			continue
		}

		needsUpdate := false
		valid := true
		newItems := make([]bson.M, 0, len(doc.Items))

		for _, item := range doc.Items {
			if item.SenseID != nil && *item.SenseID != "" {
				// already has ID
				newItems = append(newItems, bson.M{
					"word_id":  item.WordID,
					"sense_id": *item.SenseID,
					"added_at": item.AddedAt,
				})
				continue
			}

			if item.SenseIndex == nil {
				log.Printf("[MIGRATION ERROR] word_sets _id=%s item word_id=%s missing sense_id and sense_index", doc.ID.Hex(), item.WordID.Hex())
				valid = false
				break
			}

			senseID, err := resolveSenseID(item.WordID, *item.SenseIndex)
			if err != nil {
				log.Printf("[MIGRATION ERROR] word_sets _id=%s item word_id=%s sense_index=%d reason=%v", doc.ID.Hex(), item.WordID.Hex(), *item.SenseIndex, err)
				valid = false
				break
			}

			newItems = append(newItems, bson.M{
				"word_id":  item.WordID,
				"sense_id": senseID,
				"added_at": item.AddedAt,
			})
			needsUpdate = true
		}

		if !valid {
			wFailed++
			hasCriticalErrors = true
			continue
		}

		if needsUpdate {
			update := mongo.NewUpdateOneModel().
				SetFilter(bson.M{"_id": doc.ID}).
				SetUpdate(bson.M{"$set": bson.M{"items": newItems}})
			wBulkOps = append(wBulkOps, update)
		} else {
			wSkipped++
		}
	}
	wCursor.Close(ctx)

	if len(wBulkOps) > 0 {
		if _, err := wordSetsColl.BulkWrite(ctx, wBulkOps, options.BulkWrite().SetOrdered(false)); err != nil {
			log.Fatalf("Failed to bulk update word_sets: %v", err)
		}
		wMigrated = len(wBulkOps)
	}

	// 4. Migrate reviews (enrich with sense_id, keep sense_index)
	log.Println("Migrating reviews...")
	reviewsColl := learningDB.Collection("reviews")
	rCursor, err := reviewsColl.Find(ctx, bson.M{})
	if err != nil {
		log.Fatalf("Failed to query reviews: %v", err)
	}

	var rTotal, rMigrated, rSkipped, rFailed int
	var rBulkOps []mongo.WriteModel

	for rCursor.Next(ctx) {
		rTotal++
		var doc struct {
			ID         bson.ObjectID `bson:"_id"`
			WordID     bson.ObjectID `bson:"word_id"`
			SenseIndex *int          `bson:"sense_index"`
			SenseID    *string       `bson:"sense_id"`
		}
		if err := rCursor.Decode(&doc); err != nil {
			log.Printf("[MIGRATION ERROR] reviews _id=%s decode failed: %v", doc.ID.Hex(), err)
			rFailed++
			hasCriticalErrors = true
			continue
		}

		if doc.SenseID != nil && *doc.SenseID != "" {
			rSkipped++
			continue
		}

		if doc.SenseIndex == nil {
			log.Printf("[MIGRATION ERROR] reviews _id=%s missing sense_id and sense_index", doc.ID.Hex())
			rFailed++
			hasCriticalErrors = true
			continue
		}

		senseID, err := resolveSenseID(doc.WordID, *doc.SenseIndex)
		if err != nil {
			log.Printf("[MIGRATION ERROR] reviews _id=%s word_id=%s sense_index=%d reason=%v", doc.ID.Hex(), doc.WordID.Hex(), *doc.SenseIndex, err)
			rFailed++
			hasCriticalErrors = true
			continue
		}

		update := mongo.NewUpdateOneModel().
			SetFilter(bson.M{"_id": doc.ID}).
			SetUpdate(bson.M{"$set": bson.M{"sense_id": senseID}}) // NO unset here
		rBulkOps = append(rBulkOps, update)
	}
	rCursor.Close(ctx)

	if len(rBulkOps) > 0 {
		if _, err := reviewsColl.BulkWrite(ctx, rBulkOps, options.BulkWrite().SetOrdered(false)); err != nil {
			log.Fatalf("Failed to bulk update reviews: %v", err)
		}
		rMigrated = len(rBulkOps)
	}

	// 5. learning_sessions: delete 'active' and 'preparing'
	log.Println("Deleting active/preparing learning_sessions...")
	sessionsColl := learningDB.Collection("learning_sessions")
	res, err := sessionsColl.DeleteMany(ctx, bson.M{"status": bson.M{"$in": []string{"active", "preparing"}}})
	if err != nil {
		log.Fatalf("Failed to delete sessions: %v", err)
	}
	activeSessionsDeleted := res.DeletedCount

	// Report
	fmt.Println("\n==================================================")
	fmt.Println("MIGRATION REPORT")
	fmt.Println("==================================================")
	fmt.Println("learning_progress:")
	fmt.Printf("  records_total:    %d\n", pTotal)
	fmt.Printf("  records_migrated: %d\n", pMigrated)
	fmt.Printf("  records_skipped:  %d\n", pSkipped)
	fmt.Printf("  records_failed:   %d\n", pFailed)
	fmt.Println("--------------------------------------------------")
	fmt.Println("word_sets:")
	fmt.Printf("  items_total:      %d\n", wTotal)
	fmt.Printf("  items_migrated:   %d\n", wMigrated)
	fmt.Printf("  items_skipped:    %d\n", wSkipped)
	fmt.Printf("  items_failed:     %d\n", wFailed)
	fmt.Println("--------------------------------------------------")
	fmt.Println("reviews:")
	fmt.Printf("  records_total:    %d\n", rTotal)
	fmt.Printf("  records_migrated: %d\n", rMigrated)
	fmt.Printf("  records_skipped:  %d\n", rSkipped)
	fmt.Printf("  records_failed:   %d\n", rFailed)
	fmt.Println("--------------------------------------------------")
	fmt.Println("learning_sessions:")
	fmt.Printf("  active_sessions_deleted: %d\n", activeSessionsDeleted)
	fmt.Println("==================================================")

	if hasCriticalErrors {
		fmt.Println("FAILED: Migration finished with unresolved critical errors. Fix data and retry.")
		os.Exit(1)
	}

	fmt.Println("SUCCESS: Migration completed idempotently and safely.")
}
