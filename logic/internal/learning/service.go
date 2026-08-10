package learning

import (
	"context"
	"errors"
	"math/rand"
	"time"

	"github.com/english/backend/internal/dictionary"
	"github.com/english/backend/internal/gamification"
	"github.com/english/backend/internal/user"
	"go.mongodb.org/mongo-driver/v2/bson"
)

var (
	ErrWordNotFound = errors.New("word not found in dictionary")
)

type service struct {
	repo      Repository
	dictRepo  dictionary.Repository
	userRepo  user.Repository
	scheduler Scheduler
	gamiSvc   gamification.Service
}

func NewService(repo Repository, dictRepo dictionary.Repository, userRepo user.Repository, scheduler Scheduler, gamiSvc gamification.Service) Service {
	return &service{
		repo:      repo,
		dictRepo:  dictRepo,
		userRepo:  userRepo,
		scheduler: scheduler,
		gamiSvc:   gamiSvc,
	}
}

// GetDashboard returns new/review/learned counts in a single aggregation call.
func (s *service) GetDashboard(ctx context.Context, userID string) (*DashboardDTO, error) {
	uid, err := bson.ObjectIDFromHex(userID)
	if err != nil {
		return nil, err
	}

	// Get bookmarked words to compute "new" count
	bookmarks, err := s.userRepo.ListBookmarks(ctx, uid)
	if err != nil {
		return nil, err
	}

	// Resolve word strings -> word_ids via dictionary
	words := make([]string, 0, len(bookmarks))
	for _, b := range bookmarks {
		words = append(words, b.Word)
	}

	wordDocs, err := s.dictRepo.GetManyByWords(ctx, words)
	if err != nil {
		return nil, err
	}

	bookmarkedIDs := make([]bson.ObjectID, 0)
	seenWords := make(map[string]bool)
	for _, d := range wordDocs {
		if seenWords[d.Word] {
			continue
		}
		seenWords[d.Word] = true
		bookmarkedIDs = append(bookmarkedIDs, d.ID)
	}

	return s.repo.GetDashboardCounts(ctx, uid, bookmarkedIDs)
}

// GetSession builds a Flashcard or Quiz session from words due for review
// or new words that haven't been started yet.
func (s *service) GetSession(ctx context.Context, userID string, mode SessionMode, limit int) (*SessionDTO, error) {
	if limit <= 0 || limit > 20 {
		limit = 10
	}

	uid, err := bson.ObjectIDFromHex(userID)
	if err != nil {
		return nil, err
	}

	// 1. Fetch user's bookmarks
	bookmarks, err := s.userRepo.ListBookmarks(ctx, uid)
	if err != nil {
		return nil, err
	}
	if len(bookmarks) == 0 {
		return &SessionDTO{Mode: mode, Items: []SessionItemDTO{}}, nil
	}

	bookmarkWords := make([]string, 0, len(bookmarks))
	for _, b := range bookmarks {
		bookmarkWords = append(bookmarkWords, b.Word)
	}

	// 2. Resolve words -> dictionary documents
	wordDocs, err := s.dictRepo.GetManyByWords(ctx, bookmarkWords)
	if err != nil {
		return nil, err
	}

	wordDocByID := make(map[bson.ObjectID]*dictionary.WordDocument)
	allWordIDs := make([]bson.ObjectID, 0)
	seenWords := make(map[string]bool)

	for i := range wordDocs {
		w := wordDocs[i].Word
		if seenWords[w] {
			continue
		}
		seenWords[w] = true

		wordDocByID[wordDocs[i].ID] = &wordDocs[i]
		allWordIDs = append(allWordIDs, wordDocs[i].ID)
	}

	// 3. Get existing progress for all bookmarked words
	progressMap, err := s.repo.GetProgressForWords(ctx, uid, allWordIDs)
	if err != nil {
		return nil, err
	}

	// 4. Prioritise: due words first, then new words
	var dueIDs, newIDs []bson.ObjectID
	for _, id := range allWordIDs {
		p, exists := progressMap[id]
		if !exists {
			newIDs = append(newIDs, id)
		} else if IsDue(p) && !IsLearned(p) {
			dueIDs = append(dueIDs, id)
		}
	}

	// Shuffle for variety
	rand.Shuffle(len(dueIDs), func(i, j int) { dueIDs[i], dueIDs[j] = dueIDs[j], dueIDs[i] })
	rand.Shuffle(len(newIDs), func(i, j int) { newIDs[i], newIDs[j] = newIDs[j], newIDs[i] })

	selectedIDs := append(dueIDs, newIDs...)
	if len(selectedIDs) > limit {
		selectedIDs = selectedIDs[:limit]
	}

	// 5. Build session items
	items := make([]SessionItemDTO, 0, len(selectedIDs))
	for _, id := range selectedIDs {
		doc, ok := wordDocByID[id]
		if !ok {
			continue
		}

		item := buildSessionItem(doc)

		if mode == ModeQuiz {
			choices, err := s.buildQuizChoices(ctx, doc)
			if err != nil {
				// Non-fatal: fall back to flashcard item for this word
				item.Choices = nil
			} else {
				item.Choices = choices
			}
		}

		items = append(items, item)
	}

	return &SessionDTO{Mode: mode, Items: items}, nil
}

// SubmitProgress applies the Scheduler algorithm and upserts the progress document along with a review log.
func (s *service) SubmitProgress(ctx context.Context, userID string, req *SubmitProgressRequest) error {
	uid, err := bson.ObjectIDFromHex(userID)
	if err != nil {
		return err
	}
	wid, err := bson.ObjectIDFromHex(req.WordID)
	if err != nil {
		return err
	}

	// Verify word exists
	wordDoc, err := s.dictRepo.GetByID(ctx, wid)
	if err != nil {
		return err
	}
	if wordDoc == nil {
		return ErrWordNotFound
	}

	// Get current progress
	existing, err := s.repo.GetProgress(ctx, uid, wid)
	if err != nil {
		return err
	}

	var p *LearningProgress
	if existing != nil {
		p = existing
	} else {
		// New word
		p = &LearningProgress{
			UserID:    uid,
			WordID:    wid,
			CreatedAt: time.Now(),
		}
	}

	// Calculate next state
	nextP := s.scheduler.CalculateNext(p, req.Quality)

	// Build Review log
	rev := &Review{
		UserID:         uid,
		WordID:         wid,
		Quality:        req.Quality,
		ResponseTimeMs: req.ResponseTimeMs,
		OldInterval:    p.Interval,
		NewInterval:    nextP.Interval,
		OldEF:          p.EaseFactor,
		NewEF:          nextP.EaseFactor,
		OldRepetition:  p.Repetition,
		NewRepetition:  nextP.Repetition,
		ReviewedAt:     nextP.LastReviewDate,
	}

	err = s.repo.UpsertProgressAndLog(ctx, nextP, rev)
	if err != nil {
		return err
	}

	// Track Gamification
	isNewWord := (existing == nil)
	isCorrect := (req.Quality >= QualityHard)

	// Fire and forget, or at least don't block progress submission
	if s.gamiSvc != nil {
		act := gamification.ActivityReview
		if isNewWord {
			act = gamification.ActivityLearnNew
		}
		// Assuming we don't know if it's quiz or flashcard from this generic endpoint, default to review/learn
		// Hardcode daily goal 20 for now, could be fetched from user profile in the future.
		_ = s.gamiSvc.TrackActivity(ctx, userID, act, isCorrect, 20)
	}

	return nil
}

// -----------------------------------------------------------------------------
// Private helpers
// -----------------------------------------------------------------------------

func buildSessionItem(doc *dictionary.WordDocument) SessionItemDTO {
	// Extract IPA from first sound
	ipa := ""
	for _, s := range doc.Sounds {
		if s.IPA != "" {
			ipa = s.IPA
			break
		}
	}

	// Collect up to 3 definitions from senses
	var defs []string
	var example string
	for _, sense := range doc.Senses {
		for _, gloss := range sense.Glosses {
			if gloss != "" {
				defs = append(defs, gloss)
			}
		}
		if example == "" && len(sense.Examples) > 0 {
			example = sense.Examples[0].Text
		}
		if len(defs) >= 3 {
			break
		}
	}

	return SessionItemDTO{
		WordID:      doc.ID.Hex(),
		Word:        doc.Word,
		POS:         doc.POS,
		IPA:         ipa,
		Definitions: defs,
		Example:     example,
	}
}

func (s *service) buildQuizChoices(ctx context.Context, doc *dictionary.WordDocument) ([]ChoiceDTO, error) {
	// Correct answer: first gloss of first sense
	correctDef := ""
	for _, sense := range doc.Senses {
		if len(sense.Glosses) > 0 && sense.Glosses[0] != "" {
			correctDef = sense.Glosses[0]
			break
		}
	}
	if correctDef == "" {
		return nil, errors.New("no definition available")
	}

	// Distractors: same POS, different word
	distractors, err := s.dictRepo.GetDistractors(ctx, doc.POS, doc.ID, 3)
	if err != nil {
		return nil, err
	}

	choices := []ChoiceDTO{{Definition: correctDef, IsCorrect: true}}
	for _, d := range distractors {
		def := ""
		for _, sense := range d.Senses {
			if len(sense.Glosses) > 0 {
				def = sense.Glosses[0]
				break
			}
		}
		if def != "" {
			choices = append(choices, ChoiceDTO{Definition: def, IsCorrect: false})
		}
	}

	// Shuffle choices so correct answer isn't always first
	rand.Shuffle(len(choices), func(i, j int) { choices[i], choices[j] = choices[j], choices[i] })
	return choices, nil
}
