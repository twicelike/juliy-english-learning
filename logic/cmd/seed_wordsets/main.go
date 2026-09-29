package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/english/backend/internal/config"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// DTOs for decoding
type WordDoc struct {
	ID             bson.ObjectID `bson:"_id"`
	Word           string        `bson:"word"`
	FrequencyScore float64       `bson:"frequency_score"`
	Senses         []struct {
		ID string `bson:"id"`
	} `bson:"senses"`
}

type UserDoc struct {
	ID    bson.ObjectID `bson:"_id"`
	Email string        `bson:"email"`
}

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	clientOptions := options.Client().ApplyURI(cfg.MongoURI)
	mongoClient, err := mongo.Connect(clientOptions)
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer mongoClient.Disconnect(ctx)

	appDB := mongoClient.Database(cfg.MongoDB)
	dictDB := mongoClient.Database("dictionary")

	// 1. Lấy ID của Admin (hoặc user đầu tiên trong DB để gán quyền sở hữu Set)
	var user UserDoc
	err = appDB.Collection("users").FindOne(ctx, bson.M{}).Decode(&user)
	if err != nil {
		log.Fatalf("No user found to own the word sets: %v", err)
	}
	fmt.Printf("Using User [%s] (%s) as the Creator\n", user.Email, user.ID.Hex())

	// 2. Đọc file oxford.json và seed theo level
	seedOxfordFromFile(ctx, appDB, dictDB, user.ID, "./cmd/seed_wordsets/oxford.json", 50)

	fmt.Println("Seeding completed!")
}

func seedTopFrequency(ctx context.Context, appDB, dictDB *mongo.Database, userID bson.ObjectID, totalWords int, wordsPerSet int) {
	fmt.Printf("Fetching Top %d words by frequency...\n", totalWords)
	
	// Query top words
	opts := options.Find().
		SetSort(bson.D{{Key: "frequency_score", Value: -1}}).
		SetLimit(int64(totalWords))
		
	cursor, err := dictDB.Collection("words").Find(ctx, bson.M{
		"senses": bson.M{"$exists": true, "$ne": []interface{}{}},
	}, opts)
	if err != nil {
		log.Fatalf("Failed to fetch words: %v", err)
	}
	defer cursor.Close(ctx)

	var allWords []WordDoc
	if err := cursor.All(ctx, &allWords); err != nil {
		log.Fatalf("Failed to decode words: %v", err)
	}

	fmt.Printf("Found %d words.\n", len(allWords))

	// Chia thành các bộ nhỏ (Ví dụ: 300 từ / 1 set)
	totalSets := (len(allWords) + wordsPerSet - 1) / wordsPerSet

	for i := 0; i < totalSets; i++ {
		start := i * wordsPerSet
		end := start + wordsPerSet
		if end > len(allWords) {
			end = len(allWords)
		}
		
		batch := allWords[start:end]
		setName := fmt.Sprintf("Top %d Common Words - Part %d", totalWords, i+1)
		
		createWordSet(ctx, appDB, userID, setName, "Bộ từ vựng thông dụng nhất được sắp xếp theo mức độ sử dụng.", batch)
	}
}

func seedFromList(ctx context.Context, appDB, dictDB *mongo.Database, userID bson.ObjectID, setName, description string, wordList []string) {
	if len(wordList) == 0 {
		return
	}

	fmt.Printf("Creating Set '%s' from list of %d words...\n", setName, len(wordList))

	// Convert list to lowercase for exact match
	for i, w := range wordList {
		wordList[i] = strings.ToLower(strings.TrimSpace(w))
	}

	cursor, err := dictDB.Collection("words").Find(ctx, bson.M{"word": bson.M{"$in": wordList}})
	if err != nil {
		log.Printf("Failed to fetch word list: %v", err)
		return
	}
	defer cursor.Close(ctx)

	var words []WordDoc
	if err := cursor.All(ctx, &words); err != nil {
		log.Printf("Failed to decode word list: %v", err)
		return
	}

	createWordSet(ctx, appDB, userID, setName, description, words)
}

func createWordSet(ctx context.Context, appDB *mongo.Database, userID bson.ObjectID, name, description string, words []WordDoc) {
	if len(words) == 0 {
		return
	}

	now := time.Now()
	setID := bson.NewObjectID()

	// 1. Tạo WordSet Document
	wordSetDoc := bson.M{
		"_id":                 setID,
		"user_id":             userID,
		"name":                name,
		"description":         description,
		"visibility":          "public",
		"topics":              []string{"General", "Frequency"},
		"difficulty":          "beginner",
		"source_language":     "en",
		"annotation_language": "vi",
		"word_count":          len(words),
		"clone_count":         0,
		"created_at":          now,
		"updated_at":          now,
	}

	_, err := appDB.Collection("word_sets").InsertOne(ctx, wordSetDoc)
	if err != nil {
		log.Printf("Failed to insert word set '%s': %v", name, err)
		return
	}

	// 2. Tạo WordSet Items
	var items []interface{}
	for idx, w := range words {
		// Chỉ lấy sense đầu tiên
		senseID := ""
		if len(w.Senses) > 0 {
			senseID = w.Senses[0].ID
		}

		items = append(items, bson.M{
			"set_id":     setID,
			"word_id":    w.ID,
			"sense_id":   senseID,
			"position":   idx,
			"created_at": now,
		})
	}

	if len(items) > 0 {
		_, err = appDB.Collection("word_set_items").InsertMany(ctx, items)
		if err != nil {
			log.Printf("Failed to insert word set items for '%s': %v", name, err)
		}
	}

	fmt.Printf("✅ Created public set: '%s' with %d words.\n", name, len(words))
}

type OxfordWord struct {
	Word  string `json:"word"`
	Pos   string `json:"pos"` // Ví dụ: "v.", "n.", "adj."
	Level string `json:"level"`
}

func mapPOS(shortPOS string) string {
	shortPOS = strings.ToLower(strings.Trim(shortPOS, " ."))
	switch shortPOS {
	case "v":
		return "verb"
	case "n":
		return "noun"
	case "adj":
		return "adjective"
	case "adv":
		return "adverb"
	case "prep":
		return "preposition"
	case "conj":
		return "conjunction"
	case "pron":
		return "pronoun"
	case "det":
		return "determiner"
	}
	return shortPOS
}

func seedOxfordFromFile(ctx context.Context, appDB, dictDB *mongo.Database, userID bson.ObjectID, filepath string, maxWordsPerSet int) {
	fmt.Printf("Loading Oxford words from %s...\n", filepath)
	file, err := os.Open(filepath)
	if err != nil {
		log.Fatalf("Failed to open file: %v", err)
	}
	defer file.Close()

	var oxfordList []OxfordWord
	if err := json.NewDecoder(file).Decode(&oxfordList); err != nil {
		log.Fatalf("Failed to decode JSON: %v", err)
	}

	// Group words by level
	levelMap := make(map[string][]OxfordWord)
	for _, item := range oxfordList {
		lvl := strings.ToUpper(strings.TrimSpace(item.Level))
		if lvl == "" {
			continue
		}
		item.Word = strings.ToLower(strings.TrimSpace(item.Word))
		levelMap[lvl] = append(levelMap[lvl], item)
	}

	for level, list := range levelMap {
		fmt.Printf("\nProcessing %s Level (%d words)...\n", level, len(list))

		var orConditions []bson.M
		for _, w := range list {
			cond := bson.M{"word": w.Word}
			if w.Pos != "" {
				cond["pos"] = mapPOS(w.Pos)
			}
			orConditions = append(orConditions, cond)
		}

		opts := options.Find().SetSort(bson.D{{Key: "frequency_score", Value: -1}})
		cursor, err := dictDB.Collection("words").Find(ctx, bson.M{"$or": orConditions}, opts)
		if err != nil {
			log.Printf("Failed to fetch word list for %s: %v", level, err)
			continue
		}

		var rawWords []WordDoc
		if err := cursor.All(ctx, &rawWords); err != nil {
			log.Printf("Failed to decode word list for %s: %v", level, err)
			cursor.Close(ctx)
			continue
		}
		cursor.Close(ctx)

		// Lọc trùng: Mỗi từ chỉ lấy 1 document (loại từ phổ biến nhất do đã sort theo frequency)
		var words []WordDoc
		seen := make(map[string]bool)
		for _, w := range rawWords {
			wLower := strings.ToLower(strings.TrimSpace(w.Word))
			if !seen[wLower] {
				seen[wLower] = true
				words = append(words, w)
			}
		}

		if len(words) == 0 {
			fmt.Printf("No matching words found in DB for %s level. Skipping.\n", level)
			continue
		}

		totalSets := (len(words) + maxWordsPerSet - 1) / maxWordsPerSet
		for i := 0; i < totalSets; i++ {
			start := i * maxWordsPerSet
			end := start + maxWordsPerSet
			if end > len(words) {
				end = len(words)
			}

			batch := words[start:end]
			setName := fmt.Sprintf("Oxford %s - Part %d", level, i+1)
			description := fmt.Sprintf("Bộ từ vựng cốt lõi chuẩn CEFR Level %s (Từ %d đến %d).", level, start+1, end)

			createWordSet(ctx, appDB, userID, setName, description, batch)
		}
	}
}
