package match

import (
	"context"
	"math/rand"
	"time"

	"github.com/english/backend/internal/dictionary"
	"github.com/english/backend/internal/learning"
	"github.com/english/backend/internal/translation"
	"github.com/english/backend/internal/user"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type service struct {
	repo         Repository
	dictRepo     dictionary.Repository
	learningRepo learning.Repository
	userService  user.Service
	transRepo    translation.Repository
}

func NewService(repo Repository, dictRepo dictionary.Repository, learningRepo learning.Repository, userService user.Service, transRepo translation.Repository) Service {
	return &service{
		repo:         repo,
		dictRepo:     dictRepo,
		learningRepo: learningRepo,
		userService:  userService,
		transRepo:    transRepo,
	}
}

func (s *service) GetRecommendations(ctx context.Context, userID string, limit int) ([]MatchCardDTO, error) {
	uid, err := bson.ObjectIDFromHex(userID)
	if err != nil {
		return nil, err
	}

	// 1. Get words already swiped by the user
	swipedIDs, err := s.repo.GetSwipedWordIDs(ctx, uid)
	if err != nil {
		return nil, err
	}

	// 2. Fetch candidate words (top frequent words excluding swiped ones)
	// We fetch more than limit because some words might be completely learned or ineligible.
	fetchLimit := limit * 10
	if fetchLimit < 1000 {
		fetchLimit = 1000
	}
	candidates, err := s.dictRepo.GetTopWordsNotIn(ctx, swipedIDs, fetchLimit)
	if err != nil {
		return nil, err
	}

	// 3. Prepare learning progress map for Progressive Discovery
	var queryRefs []learning.SessionItemRef
	for _, doc := range candidates {
		queryRefs = append(queryRefs, learning.SessionItemRef{WordID: doc.ID})
	}
	progressMap, err := s.learningRepo.GetProgressForWords(ctx, uid, queryRefs)
	if err != nil {
		return nil, err
	}

	// 4. Batch-fetch translations for these candidate words
	var wordIDs []bson.ObjectID
	for _, doc := range candidates {
		wordIDs = append(wordIDs, doc.ID)
	}
	translations, err := s.transRepo.GetTranslationsByWordIDs(ctx, wordIDs, "vi")
	if err != nil {
		return nil, err
	}
	translationMap := make(map[string]string) // senseID -> translation text
	for _, tr := range translations {
		if len(tr.Translations) > 0 {
			translationMap[tr.SenseID] = tr.Translations[0] // Pick the primary translation
		}
	}

	// Shuffle candidates to add randomness within the top frequent words pool
	// This makes the deck feel fresh and spaced out, while still limiting to top N words.
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	r.Shuffle(len(candidates), func(i, j int) {
		candidates[i], candidates[j] = candidates[j], candidates[i]
	})

	// 5. Resolve display sense for each candidate
	cards := make([]MatchCardDTO, 0, limit)
	seenSpellings := make(map[string]bool)

	for _, doc := range candidates {
		// Prevent showing the exact same spelling (e.g. noun and verb forms) in the same batch
		if seenSpellings[doc.Word] {
			continue
		}

		// Group progress for this word
		wordProgresses := make(map[string]*learning.LearningProgress)
		for key, p := range progressMap {
			if key.WordID == doc.ID {
				wordProgresses[key.SenseID] = p
			}
		}

		// Use the Sense-aware discovery policy
		nextSense := learning.DiscoverNextSense(&doc, wordProgresses)
		if nextSense == nil {
			// All eligible senses are mastered, or no eligible senses exist.
			continue
		}

		// Also check if this sense is currently being learned
		if _, exists := wordProgresses[nextSense.ID]; exists {
			// If it exists, they are actively learning it (DiscoverNextSense returns the active frontier).
			// We shouldn't recommend it in Match since Match is for top-of-funnel discovery.
			continue
		}

		// Check if we have a Vietnamese translation for this sense
		viTranslation, hasTranslation := translationMap[nextSense.ID]
		if !hasTranslation || viTranslation == "" {
			// Skip words missing primary sense translations to keep Match UI instantaneous and complete
			continue
		}

		// Compute community signal (only if we have enough sample size, e.g. >= 10 interactions)
		likes, total, _ := s.repo.GetCommunitySignal(ctx, doc.ID)
		matchRate := 0
		if total >= 10 {
			matchRate = int((float64(likes) / float64(total)) * 100)
		}

		definition := ""
		if len(nextSense.Glosses) > 0 {
			definition = nextSense.Glosses[0]
		}

		cards = append(cards, MatchCardDTO{
			WordID:      doc.ID.Hex(),
			Word:        doc.Word,
			POS:         doc.POS,
			SenseID:     nextSense.ID,
			Definition:  definition,
			Translation: viTranslation,
			TotalSwipes: total,
			MatchRate:   matchRate,
		})
		seenSpellings[doc.Word] = true

		if len(cards) >= limit {
			break
		}
	}

	return cards, nil
}

func (s *service) Swipe(ctx context.Context, userID string, req *SwipeRequest) error {
	uid, err := bson.ObjectIDFromHex(userID)
	if err != nil {
		return err
	}
	wid, err := bson.ObjectIDFromHex(req.WordID)
	if err != nil {
		return err
	}

	// 1. Record the swipe action (upsert to handle duplicates cleanly)
	swipe := &UserSwipe{
		UserID: uid,
		WordID: wid,
		Action: req.Action,
	}
	err = s.repo.RecordSwipe(ctx, swipe)
	if err != nil {
		return err
	}

	return nil
}
