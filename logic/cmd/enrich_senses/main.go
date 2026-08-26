package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/english/backend/internal/ai/client"
	"github.com/english/backend/internal/config"
	"github.com/english/backend/internal/dictionary"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

const EnrichmentVersion = "v1"

// ----------------------------------------------------------------------------
// API Payloads
// ----------------------------------------------------------------------------

type AIEnrichmentRequest struct {
	Words []AIWordContext `json:"words"`
}

type AIWordContext struct {
	WordID string         `json:"word_id"`
	Word   string         `json:"word"`
	POS    string         `json:"pos"`
	Senses []AISenseInput `json:"senses"`
}

type AISenseInput struct {
	SenseID    string   `json:"sense_id"`
	Definition string   `json:"definition"`
	Examples   []string `json:"examples"`
}

type AIEnrichmentResponse struct {
	Words []AIWordResult `json:"words"`
}

type AIWordResult struct {
	WordID string          `json:"word_id"`
	Senses []AISenseResult `json:"senses"`
}

type AISenseResult struct {
	SenseID        string                   `json:"sense_id"`
	QualityScore   int                      `json:"quality_score"`
	FrequencyScore int                      `json:"frequency_score"`
	Category       dictionary.SenseCategory `json:"category"`
}

// ----------------------------------------------------------------------------
// Validations
// ----------------------------------------------------------------------------

func isValidCategory(c dictionary.SenseCategory) bool {
	switch c {
	case dictionary.CategoryPrimary, dictionary.CategoryCore, dictionary.CategoryUseful,
		dictionary.CategoryNiche, dictionary.CategoryRare, dictionary.CategoryIgnore:
		return true
	}
	return false
}

// ----------------------------------------------------------------------------
// Prompt Builder
// ----------------------------------------------------------------------------

func buildPrompt(req AIEnrichmentRequest) string {
	jsonData, _ := json.Marshal(req)
	return fmt.Sprintf(`You are an expert lexicographer and language learning curriculum designer.
I am providing you with a batch of English words and their senses in JSON format.
Your task is to classify and score EVERY sense provided based on its usefulness for an English learner.

Rules:
1. QualityScore (0-100): How good, clear, and distinct is this definition? (e.g. 100 = perfect core meaning, 10 = obscure, archaic, or circular definition).
2. FrequencyScore (0-100): How common is this specific sense in modern, everyday English? (100 = extremely common, 10 = almost never used).
3. Category: Must be exactly one of: PRIMARY, CORE, USEFUL, NICHE, RARE, IGNORE.
   - PRIMARY: The most dominant, frequent, and important meaning of the word. Usually only one per word.
   - CORE: Other very common meanings that a B1-B2 learner must know.
   - USEFUL: Meanings a C1 learner should know.
   - NICHE: Highly specific domain meanings (e.g. heraldry, obscure science).
   - RARE: Archaic or obsolete meanings.
   - IGNORE: Definitions that are just references, misspellings, or completely useless for learners.

Output ONLY valid JSON matching the exact expected schema. Do not include markdown formatting like json blocks.
Input JSON:
%s
`, string(jsonData))
}

// ----------------------------------------------------------------------------
// Worker Logic
// ----------------------------------------------------------------------------

type WorkerPool struct {
	aiClient client.GeminiClient // Using Gemini client abstraction directly for the script
	coll     *mongo.Collection

	processed uint64
	success   uint64
	failed    uint64

	tokenLimit int
}

func (p *WorkerPool) ProcessBatch(ctx context.Context, batch []dictionary.WordDocument) error {
	if len(batch) == 0 {
		return nil
	}

	// 1. Build the AI input
	var req AIEnrichmentRequest
	for _, doc := range batch {
		wordCtx := AIWordContext{
			WordID: doc.ID.Hex(),
			Word:   doc.Word,
			POS:    doc.POS,
		}
		for _, sense := range doc.Senses {
			def := ""
			if len(sense.Glosses) > 0 {
				def = sense.Glosses[0]
			}
			var ex []string
			for i, e := range sense.Examples {
				if i >= 2 {
					break
				}
				ex = append(ex, e.Text)
			}
			wordCtx.Senses = append(wordCtx.Senses, AISenseInput{
				SenseID:    sense.ID,
				Definition: def,
				Examples:   ex,
			})
		}
		req.Words = append(req.Words, wordCtx)
	}

	prompt := buildPrompt(req)

	// 2. Call AI with retry
	var resp AIEnrichmentResponse
	maxRetries := 3
	var err error

	for i := 0; i < maxRetries; i++ {
		err = p.aiClient.GenerateJSON(ctx, prompt, nil, &resp)
		if err == nil {
			break
		}
		if strings.Contains(err.Error(), "429") || strings.Contains(err.Error(), "503") {
			log.Printf("Rate limit hit, backing off... (%v)", err)
			time.Sleep(time.Duration(i*2+2) * time.Second) // backoff
			continue
		}
		break // Not a retriable error
	}

	if err != nil {
		atomic.AddUint64(&p.failed, uint64(len(batch)))
		return fmt.Errorf("AI generation failed after retries: %w", err)
	}

	// 3. Validate and Update DB
	now := time.Now().UTC()

	// Build a map of results for fast lookup
	resultMap := make(map[string]map[string]AISenseResult)
	for _, w := range resp.Words {
		resultMap[w.WordID] = make(map[string]AISenseResult)
		for _, s := range w.Senses {
			resultMap[w.WordID][s.SenseID] = s
		}
	}

	var bulkOps []mongo.WriteModel

	for _, doc := range batch {
		wordResults, ok := resultMap[doc.ID.Hex()]
		if !ok {
			log.Printf("AI missed word_id %s", doc.ID.Hex())
			atomic.AddUint64(&p.failed, 1)
			continue
		}

		updateFields := bson.M{}
		isValid := true

		for i, sense := range doc.Senses {
			res, ok := wordResults[sense.ID]
			if !ok {
				log.Printf("AI missed sense_id %s for word %s", sense.ID, doc.ID.Hex())
				isValid = false
				break
			}

			if res.QualityScore < 0 || res.QualityScore > 100 || res.FrequencyScore < 0 || res.FrequencyScore > 100 {
				log.Printf("Invalid scores for sense %s: Q:%d F:%d", sense.ID, res.QualityScore, res.FrequencyScore)
				isValid = false
				break
			}
			if !isValidCategory(res.Category) {
				log.Printf("Invalid category for sense %s: %s", sense.ID, res.Category)
				isValid = false
				break
			}

			// Positional update using array index in memory
			quality := dictionary.SenseQuality{
				QualityScore:      res.QualityScore,
				FrequencyScore:    res.FrequencyScore,
				Category:          res.Category,
				EnrichmentVersion: EnrichmentVersion,
				EnrichedAt:        &now,
			}
			updateFields[fmt.Sprintf("senses.%d.quality", i)] = quality
		}

		if isValid && len(updateFields) > 0 {
			update := mongo.NewUpdateOneModel().
				SetFilter(bson.M{"_id": doc.ID}).
				SetUpdate(bson.M{"$set": updateFields})
			bulkOps = append(bulkOps, update)
			atomic.AddUint64(&p.success, 1)
		} else {
			atomic.AddUint64(&p.failed, 1)
		}
		atomic.AddUint64(&p.processed, 1)
	}

	if len(bulkOps) > 0 {
		_, err := p.coll.BulkWrite(ctx, bulkOps, options.BulkWrite().SetOrdered(false))
		if err != nil {
			return fmt.Errorf("bulk write failed: %w", err)
		}
	}

	return nil
}

func (p *WorkerPool) Start(ctx context.Context, numWorkers int, jobs <-chan []dictionary.WordDocument) {
	var wg sync.WaitGroup
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				case batch, ok := <-jobs:
					if !ok {
						return // channel closed
					}
					if err := p.ProcessBatch(ctx, batch); err != nil {
						log.Printf("[Worker %d] Failed to process batch: %v", workerID, err)
					}
				}
			}
		}(i)
	}
	wg.Wait()
}

// ----------------------------------------------------------------------------
// Main
// ----------------------------------------------------------------------------

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Configuration
	numWorkers := 5
	if w, err := strconv.Atoi(os.Getenv("ENRICH_WORKERS")); err == nil && w > 0 {
		numWorkers = w
	}
	aiProvider := os.Getenv("ENRICH_AI_PROVIDER") // gemini is default
	if aiProvider == "" {
		aiProvider = "gemini"
	}
	aiModel := os.Getenv("ENRICH_AI_MODEL")
	if aiModel == "" {
		aiModel = "gemini-2.0-flash" // low cost, high throughput
	}
	tokenLimit := 8000
	if tl, err := strconv.Atoi(os.Getenv("ENRICH_BATCH_TOKEN_LIMIT")); err == nil && tl > 0 {
		tokenLimit = tl
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle Graceful Shutdown
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs
		log.Println("\nReceived shutdown signal. Stopping gracefully...")
		cancel()
	}()

	clientOptions := options.Client().ApplyURI(cfg.MongoURI)
	mongoClient, err := mongo.Connect(clientOptions)
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer mongoClient.Disconnect(context.Background())

	coll := mongoClient.Database("dictionary").Collection("words")

	// Init AI Client
	log.Printf("Initializing AI Client (Provider: %s, Model: %s)", aiProvider, aiModel)
	geminiClient, err := client.NewGeminiClient(ctx, cfg.GeminiAPIKey, aiModel)
	if err != nil {
		log.Fatalf("Failed to init AI client: %v", err)
	}
	defer geminiClient.Close()

	pool := &WorkerPool{
		aiClient:   *geminiClient,
		coll:       coll,
		tokenLimit: tokenLimit,
	}

	// Valid POS filtering
	validPOS := []string{
		"noun", "verb", "adjective", "adverb",
		"phrasal verb", "preposition", "conjunction", "pronoun",
	}
	posFilter := bson.A{}
	for _, p := range validPOS {
		posFilter = append(posFilter, p)
	}

	// Filter: valid POS AND (senses exist) AND (any sense lacks current EnrichmentVersion)
	filter := bson.M{
		"pos": bson.M{"$in": posFilter},
		"senses": bson.M{
			"$elemMatch": bson.M{
				"$or": []bson.M{
					{"quality": bson.M{"$exists": false}},
					{"quality": nil},
					{"quality.enrichment_version": bson.M{"$ne": EnrichmentVersion}},
				},
			},
		},
	}

	// For stats reporting
	totalMatch, err := coll.CountDocuments(ctx, filter)
	if err != nil {
		log.Fatalf("Failed to count documents: %v", err)
	}
	log.Printf("Found %d words requiring enrichment.", totalMatch)

	// Setup pipeline
	jobs := make(chan []dictionary.WordDocument, numWorkers*2)
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		pool.Start(ctx, numWorkers, jobs)
	}()

	cursor, err := coll.Find(ctx, filter)
	if err != nil {
		log.Fatalf("Cursor error: %v", err)
	}

	var batch []dictionary.WordDocument
	var currentTokens int

	// Progress reporter
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				p := atomic.LoadUint64(&pool.processed)
				s := atomic.LoadUint64(&pool.success)
				f := atomic.LoadUint64(&pool.failed)
				rem := uint64(totalMatch) - p
				log.Printf("[Progress] processed=%d success=%d failed=%d remaining=%d", p, s, f, rem)
			}
		}
	}()

	log.Printf("Starting feed to %d workers...", numWorkers)
outer:
	for cursor.Next(ctx) {
		var doc dictionary.WordDocument
		if err := cursor.Decode(&doc); err != nil {
			log.Printf("Failed to decode: %v", err)
			continue
		}

		// Estimate tokens: roughly 4 chars per token.
		// Word context approx: word + POS + definitions (say 100 chars per sense)
		chars := len(doc.Word) + len(doc.POS)
		for _, s := range doc.Senses {
			for _, g := range s.Glosses {
				chars += len(g)
			}
		}
		estimatedTokens := chars / 4
		if estimatedTokens < 20 {
			estimatedTokens = 20 // min baseline
		}

		if currentTokens+estimatedTokens > tokenLimit && len(batch) > 0 {
			// Send batch
			select {
			case jobs <- batch:
			case <-ctx.Done():
				break outer
			}
			batch = nil
			currentTokens = 0
		}

		batch = append(batch, doc)
		currentTokens += estimatedTokens
	}

	if len(batch) > 0 {
		select {
		case jobs <- batch:
		case <-ctx.Done():
		}
	}

	close(jobs)
	log.Println("Finished reading from DB. Waiting for workers to finish...")
	wg.Wait()

	log.Println("Enrichment run completed (or interrupted).")
}
