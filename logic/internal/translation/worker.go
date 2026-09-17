package translation

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/english/backend/internal/dictionary"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type Worker interface {
	Start(ctx context.Context)
}

type worker struct {
	repo         Repository
	svc          Service
	dictSvc      dictionary.Service
	batchSize    int
	pollInterval time.Duration
	concurrency  int
	maxRetries   int
}

func NewWorker(repo Repository, svc Service, dictSvc dictionary.Service, batchSize int, pollInterval time.Duration, concurrency int, maxRetries int) Worker {
	if batchSize <= 0 {
		batchSize = 10
	}
	if pollInterval <= 0 {
		pollInterval = 2 * time.Second
	}
	if concurrency <= 0 {
		concurrency = 1
	}
	if maxRetries < 0 {
		maxRetries = 3
	}
	return &worker{
		repo:         repo,
		svc:          svc,
		dictSvc:      dictSvc,
		batchSize:    batchSize,
		pollInterval: pollInterval,
		concurrency:  concurrency,
		maxRetries:   maxRetries,
	}
}

func (w *worker) Start(ctx context.Context) {
	slog.Info("Translation worker started", "concurrency", w.concurrency, "batch_size", w.batchSize)
	for i := 0; i < w.concurrency; i++ {
		go w.runLoop(ctx, i)
	}
}

func (w *worker) runLoop(ctx context.Context, workerID int) {
	ticker := time.NewTicker(w.pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			slog.Info("Translation worker stopping", "worker_id", workerID)
			return
		case <-ticker.C:
			w.processBatch(ctx)
		}
	}
}

func (w *worker) processBatch(ctx context.Context) {
	jobs, err := w.repo.ClaimJobs(ctx, w.batchSize)
	if err != nil {
		slog.Error("Worker failed to claim jobs", "error", err)
		return
	}
	if len(jobs) == 0 {
		return
	}

	langGroups := make(map[string][]TranslationJob)
	for _, j := range jobs {
		langGroups[j.TargetLanguage] = append(langGroups[j.TargetLanguage], j)
	}

	for lang, groupJobs := range langGroups {
		w.processGroup(ctx, lang, groupJobs)
	}
}

func (w *worker) processGroup(ctx context.Context, targetLang string, jobs []TranslationJob) {
	var wordIDs []bson.ObjectID
	jobMap := make(map[bson.ObjectID]TranslationJob)
	for _, j := range jobs {
		wordIDs = append(wordIDs, j.WordID)
		jobMap[j.WordID] = j
	}

	// --- WA7 Deduplication ---
	existing, err := w.repo.GetTranslationsByWordIDs(ctx, wordIDs, targetLang)
	if err != nil {
		w.failJobs(ctx, jobs, fmt.Errorf("failed to check existing translations: %w", err))
		return
	}

	existingMap := make(map[bson.ObjectID]bool)
	for _, ex := range existing {
		existingMap[ex.WordID] = true
	}

	var missingWordIDs []bson.ObjectID
	for _, wID := range wordIDs {
		if existingMap[wID] {
			w.repo.UpdateJobStatus(ctx, jobMap[wID].ID, JobStatusCompleted, "", time.Time{})
		} else {
			missingWordIDs = append(missingWordIDs, wID)
		}
	}

	if len(missingWordIDs) == 0 {
		return // All words already translated
	}

	words, err := w.dictSvc.GetManyByIDs(ctx, missingWordIDs)
	if err != nil {
		w.failJobs(ctx, jobs, fmt.Errorf("failed to fetch words from dictionary: %w", err))
		return
	}

	wordMap := make(map[string]dictionary.WordDTO)
	for _, w := range words {
		wordMap[w.ID] = w
	}

	input := TranslateBatchInput{
		TargetLanguage: targetLang,
	}

	for _, wID := range missingWordIDs {
		wIDStr := wID.Hex()
		dictWord, ok := wordMap[wIDStr]
		if !ok {
			w.repo.UpdateJobStatus(ctx, jobMap[wID].ID, JobStatusFailed, "word not found in dictionary", time.Now().Add(8760*time.Hour))
			continue
		}

		var meanings []MeaningInput
		for _, s := range dictWord.Definitions {
			c := dictionary.SenseCategory(s.Category)
			// Cho qua nếu chưa enrich (c == "") hoặc thuộc category hợp lệ
			if c != "" && c != dictionary.CategoryPrimary && c != dictionary.CategoryCore && c != dictionary.CategoryUseful {
				continue
			}
			meanings = append(meanings, MeaningInput{
				SenseID:    s.SenseID,
				Definition: s.Definition,
			})
			if len(meanings) >= 5 {
				break
			}
		}
		
		if len(meanings) == 0 {
			w.repo.UpdateJobStatus(ctx, jobMap[wID].ID, JobStatusCompleted, "all senses ineligible (NICHE/RARE/IGNORE)", time.Time{})
			continue
		}

		input.Words = append(input.Words, TranslateWordInput{
			WordID:       wIDStr,
			Word:         dictWord.Word,
			PartOfSpeech: dictWord.POS,
			Meanings:     meanings,
		})
	}

	if len(input.Words) == 0 {
		return
	}

	translations, err := w.svc.TranslateBatch(ctx, input)
	if err != nil {
		w.failJobs(ctx, jobs, fmt.Errorf("AI translation failed: %w", err))
		return
	}

	if err := w.repo.SaveManyTranslations(ctx, translations); err != nil {
		w.failJobs(ctx, jobs, fmt.Errorf("failed to save translations: %w", err))
		return
	}

	translatedSet := make(map[bson.ObjectID]bool)
	for _, tr := range translations {
		translatedSet[tr.WordID] = true
	}

	for _, j := range jobs {
		if translatedSet[j.WordID] {
			w.repo.UpdateJobStatus(ctx, j.ID, JobStatusCompleted, "", time.Time{})
		} else {
			w.failJobSingle(ctx, j, fmt.Errorf("AI omitted this word from response"))
		}
	}
}

func (w *worker) failJobs(ctx context.Context, jobs []TranslationJob, err error) {
	for _, j := range jobs {
		w.failJobSingle(ctx, j, err)
	}
}

func (w *worker) failJobSingle(ctx context.Context, j TranslationJob, err error) {
	// ClaimJobs already ran $inc {attempts: 1} in the DB before returning the job object.
	// So j.Attempts is the pre-increment value; the actual DB count is j.Attempts+1.
	actualAttempts := j.Attempts + 1
	if actualAttempts >= w.maxRetries {
		w.repo.UpdateJobStatus(ctx, j.ID, JobStatusFailed, err.Error(), time.Now().Add(8760*time.Hour))
	} else {
		backoffMin := 1 << actualAttempts
		nextAttempt := time.Now().Add(time.Duration(backoffMin) * time.Minute)
		w.repo.UpdateJobStatus(ctx, j.ID, JobStatusFailed, err.Error(), nextAttempt)
	}
}
