package learning

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"time"

	"github.com/english/backend/internal/dictionary"
	"github.com/english/backend/internal/gamification"
	"github.com/english/backend/internal/user"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
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
	exSvc     ExerciseContentService
	enricher  SessionEnricher
}

func NewService(repo Repository, dictRepo dictionary.Repository, userRepo user.Repository, scheduler Scheduler, gamiSvc gamification.Service, exSvc ExerciseContentService, enricher SessionEnricher) Service {
	return &service{
		repo:      repo,
		dictRepo:  dictRepo,
		userRepo:  userRepo,
		scheduler: scheduler,
		gamiSvc:   gamiSvc,
		exSvc:     exSvc,
		enricher:  enricher,
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
	allWordRefs := make([]SessionItemRef, 0)
	seenWords := make(map[string]bool)

	for i := range wordDocs {
		w := wordDocs[i].Word
		if seenWords[w] {
			continue
		}
		seenWords[w] = true

		wordDocByID[wordDocs[i].ID] = &wordDocs[i]
		allWordRefs = append(allWordRefs, SessionItemRef{WordID: wordDocs[i].ID, SenseIndex: 0})
	}

	// 3. Get existing progress for all bookmarked words
	progressMap, err := s.repo.GetProgressForWords(ctx, uid, allWordRefs)
	if err != nil {
		return nil, err
	}

	// 4. Prioritise: due words first, then new words
	var dueRefs, newRefs []SessionItemRef
	for _, ref := range allWordRefs {
		key := LearningProgressKey{WordID: ref.WordID, SenseIndex: ref.SenseIndex}
		p, exists := progressMap[key]
		if !exists {
			newRefs = append(newRefs, ref)
		} else if IsDue(p) && !IsLearned(p) {
			dueRefs = append(dueRefs, ref)
		}
	}

	// Shuffle for variety
	rand.Shuffle(len(dueRefs), func(i, j int) { dueRefs[i], dueRefs[j] = dueRefs[j], dueRefs[i] })
	rand.Shuffle(len(newRefs), func(i, j int) { newRefs[i], newRefs[j] = newRefs[j], newRefs[i] })

	selectedRefs := append(dueRefs, newRefs...)
	if len(selectedRefs) > limit {
		selectedRefs = selectedRefs[:limit]
	}

	// 5. Build session items
	items := make([]SessionItemDTO, 0, len(selectedRefs))
	for _, ref := range selectedRefs {
		doc, ok := wordDocByID[ref.WordID]
		if !ok {
			continue
		}

		item, err := buildSessionItem(doc, ref.SenseIndex)
		if err != nil {
			continue
		}

		if mode == ModeQuiz {
			choices, err := s.buildQuizChoices(ctx, doc, ref.SenseIndex)
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

func (s *service) CreateNewWordsSession(ctx context.Context, userID string) (*SessionDTO, error) {
	uid, err := bson.ObjectIDFromHex(userID)
	if err != nil {
		return nil, err
	}

	// 1. Check for active/preparing session
	active, err := s.repo.GetActiveSession(ctx, uid, "new_words")
	if err != nil {
		return nil, err
	}
	if active != nil {
		return s.buildSessionDTO(ctx, active)
	}

	// 2. Fetch up to 10 pending words
	limit := 10
	pendingIDs, err := s.repo.GetPendingWordIDs(ctx, uid, limit)
	if err != nil {
		return nil, err
	}

	// 3. If pending < limit, fetch new available words
	var newAvailableRefs []SessionItemRef
	if len(pendingIDs) < limit {
		newLimit := limit - len(pendingIDs)
		
		bookmarks, err := s.userRepo.ListBookmarks(ctx, uid)
		if err == nil && len(bookmarks) > 0 {
			bookmarkWords := make([]string, 0, len(bookmarks))
			for _, b := range bookmarks {
				bookmarkWords = append(bookmarkWords, b.Word)
			}
			
			wordDocs, _ := s.dictRepo.GetManyByWords(ctx, bookmarkWords)
			
			allWordRefs := make([]SessionItemRef, 0)
			seenWords := make(map[string]bool)
			for i := range wordDocs {
				w := wordDocs[i].Word
				if seenWords[w] {
					continue
				}
				seenWords[w] = true
				allWordRefs = append(allWordRefs, SessionItemRef{WordID: wordDocs[i].ID, SenseIndex: 0})
			}
			
			progressMap, _ := s.repo.GetProgressForWords(ctx, uid, allWordRefs)
			for _, ref := range allWordRefs {
				key := LearningProgressKey{WordID: ref.WordID, SenseIndex: ref.SenseIndex}
				if _, exists := progressMap[key]; !exists {
					newAvailableRefs = append(newAvailableRefs, ref)
					if len(newAvailableRefs) == newLimit {
						break
					}
				}
			}
		}
	}

	if len(pendingIDs) == 0 && len(newAvailableRefs) == 0 {
		return nil, errors.New("no_new_words")
	}

	// 4. Combine words for the session (Snapshot)
	sessionRefs := append(pendingIDs, newAvailableRefs...)

	now := time.Now()
	// 5. Create session in "preparing" state. Do NOT create progress yet.
	session := &LearningSession{
		ID:           bson.NewObjectID(),
		UserID:       uid,
		Mode:         "new_words",
		Items:        sessionRefs,
		CurrentIndex: 0,
		Status:       SessionPreparing,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	err = s.repo.CreateSession(ctx, session)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			active, _ := s.repo.GetActiveSession(ctx, uid, "new_words")
			if active != nil {
				return s.buildSessionDTO(ctx, active)
			}
		}
		return nil, err
	}

	// 6. Async trigger AI content preparation
	go func(sid bson.ObjectID, refs []SessionItemRef) {
		bgCtx := context.Background()
		if s.exSvc != nil {
			wids := make([]bson.ObjectID, len(refs))
			for i, r := range refs {
				wids[i] = r.WordID
			}
			_, _ = s.exSvc.GetExercises(bgCtx, wids, ExerciseMultipleChoice, 1)
			_, _ = s.exSvc.GetExercises(bgCtx, wids, ExerciseFillBlank, 1)
		}
		_ = s.repo.UpdateSessionStatus(bgCtx, sid, SessionActive)
	}(session.ID, sessionRefs)

	// 7. Build and return DTO (will show status "preparing")
	return s.buildSessionDTO(ctx, session)
}

func (s *service) CreateReviewSession(ctx context.Context, userID string) (*SessionDTO, error) {
	uid, err := bson.ObjectIDFromHex(userID)
	if err != nil {
		return nil, err
	}

	// 1. Check for active/preparing session
	active, err := s.repo.GetActiveSession(ctx, uid, "review")
	if err != nil {
		return nil, err
	}
	if active != nil {
		return s.buildSessionDTO(ctx, active)
	}

	// 2. Fetch up to 20 due words
	limit := 20
	dueIDs, err := s.repo.GetDueWordIDs(ctx, uid, limit)
	if err != nil {
		return nil, err
	}
	if len(dueIDs) == 0 {
		return nil, errors.New("no_due_words")
	}

	now := time.Now()
	// 3. Create session in "preparing" state
	session := &LearningSession{
		ID:           bson.NewObjectID(),
		UserID:       uid,
		Mode:         "review",
		Items:        dueIDs,
		CurrentIndex: 0,
		Status:       SessionPreparing,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	err = s.repo.CreateSession(ctx, session)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			active, _ := s.repo.GetActiveSession(ctx, uid, "review")
			if active != nil {
				return s.buildSessionDTO(ctx, active)
			}
		}
		return nil, err
	}

	// 4. Asynchronously trigger content preparation
	go func(sid bson.ObjectID, refs []SessionItemRef) {
		bgCtx := context.Background()
		if s.exSvc != nil {
			wids := make([]bson.ObjectID, len(refs))
			for i, r := range refs {
				wids[i] = r.WordID
			}
			_, _ = s.exSvc.GetExercises(bgCtx, wids, ExerciseMultipleChoice, 1)
			_, _ = s.exSvc.GetExercises(bgCtx, wids, ExerciseFillBlank, 1)
		}
		_ = s.repo.UpdateSessionStatus(bgCtx, sid, SessionActive)
	}(session.ID, dueIDs)

	// 5. Build and return DTO
	return s.buildSessionDTO(ctx, session)
}

func (s *service) CreateWordSetReviewSession(ctx context.Context, userID string, setID string, strategy ReviewStrategy, wordCount int) (*SessionDTO, error) {
	uid, err := bson.ObjectIDFromHex(userID)
	if err != nil {
		return nil, err
	}
	sID, err := bson.ObjectIDFromHex(setID)
	if err != nil {
		return nil, err
	}

	// 1. Check for active session for this word set
	active, err := s.repo.GetActiveWordSetSession(ctx, uid, sID)
	if err != nil {
		return nil, err
	}

	// 2. Fetch word IDs for this set (and verify ownership)
	wordIDs, err := s.repo.GetWordIDsForWordSet(ctx, uid, sID)
	if err != nil {
		return nil, err
	}

	if active != nil {
		isValid := true
		if len(wordIDs) == 0 {
			isValid = false
		} else {
			currentWordIDs := make(map[string]bool)
			for _, ref := range wordIDs {
				key := fmt.Sprintf("%s:%d", ref.WordID.Hex(), ref.SenseIndex)
				currentWordIDs[key] = true
			}
			for _, item := range active.Items {
				key := fmt.Sprintf("%s:%d", item.WordID.Hex(), item.SenseIndex)
				if !currentWordIDs[key] {
					isValid = false
					break
				}
			}
		}

		if isValid {
			return s.buildSessionDTO(ctx, active)
		} else {
			_ = s.repo.UpdateSessionStatus(ctx, active.ID, SessionExpired)
		}
	}

	if len(wordIDs) == 0 {
		return nil, ErrNoReviewableWords
	}

	// 3. Get Learning Progress in batch
	progressMap, err := s.repo.GetProgressForWords(ctx, uid, wordIDs)
	if err != nil {
		return nil, err
	}

	// 4. Classify words
	var due, learning, newWords []SessionItemRef
	now := time.Now()
	for _, ref := range wordIDs {
		key := LearningProgressKey{WordID: ref.WordID, SenseIndex: ref.SenseIndex}
		p, exists := progressMap[key]
		if !exists {
			newWords = append(newWords, ref)
		} else {
			if p.Interval >= 21 { // Mastered
				continue
			}
			if p.NextReviewDate.Before(now) || p.NextReviewDate.Equal(now) {
				due = append(due, ref)
			} else {
				learning = append(learning, ref)
			}
		}
	}

	// 5. Select words
	selectedIDs, err := SelectReviewWords(due, learning, newWords, strategy, wordCount)
	if err != nil {
		return nil, err
	}

	// 6. Create session
	session := &LearningSession{
		ID:           bson.NewObjectID(),
		UserID:       uid,
		Mode:         "review",
		Source:       "word_set",
		WordSetID:    &sID,
		Strategy:     &strategy,
		WordCount:    wordCount,
		Items:        selectedIDs,
		CurrentIndex: 0,
		Status:       SessionPreparing,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	err = s.repo.CreateSession(ctx, session)
	if err != nil {
		// If duplicate key error, we just return the active one or fail
		if mongo.IsDuplicateKeyError(err) {
			active, _ := s.repo.GetActiveWordSetSession(ctx, uid, sID)
			if active != nil {
				return s.buildSessionDTO(ctx, active)
			}
		}
		return nil, err
	}

	// 7. Async trigger AI content preparation (Wait, we should get unique WordIDs for exercise gen, but for now we pass WordIDs)
	go func(sid bson.ObjectID, refs []SessionItemRef) {
		bgCtx := context.Background()
		if s.exSvc != nil {
			// Extract WordIDs for exercise generation
			wids := make([]bson.ObjectID, len(refs))
			for i, r := range refs {
				wids[i] = r.WordID
			}
			_, _ = s.exSvc.GetExercises(bgCtx, wids, ExerciseMultipleChoice, 1)
			_, _ = s.exSvc.GetExercises(bgCtx, wids, ExerciseFillBlank, 1)
		}
		_ = s.repo.UpdateSessionStatus(bgCtx, sid, SessionActive)
	}(session.ID, selectedIDs)

	// 8. Build DTO
	return s.buildSessionDTO(ctx, session)
}


func (s *service) GetSessionByID(ctx context.Context, sessionID string) (*SessionDTO, error) {
	sid, err := bson.ObjectIDFromHex(sessionID)
	if err != nil {
		return nil, err
	}

	session, err := s.repo.GetSessionByID(ctx, sid)
	if err != nil {
		return nil, err
	}
	if session == nil {
		return nil, errors.New("session not found")
	}

	return s.buildSessionDTO(ctx, session)
}

func (s *service) buildSessionDTO(ctx context.Context, session *LearningSession) (*SessionDTO, error) {
	// Extract unique WordIDs for queries
	wordIDs := make([]bson.ObjectID, len(session.Items))
	for i, ref := range session.Items {
		wordIDs[i] = ref.WordID
	}

	// Resolve dictionary words
	wordDocs, err := s.dictRepo.GetManyByIDs(ctx, wordIDs)
	if err != nil {
		return nil, err
	}

	wordDocByID := make(map[bson.ObjectID]*dictionary.WordDocument)
	for i := range wordDocs {
		wordDocByID[wordDocs[i].ID] = &wordDocs[i]
	}
	
	// Fetch ONLY Fill Blank Exercises via AI
	var fbExercises map[bson.ObjectID][]ExerciseContent
	if s.exSvc != nil {
		fbExercises, _ = s.exSvc.GetExercises(ctx, wordIDs, ExerciseFillBlank, 1)
	}

	items := make([]SessionItemDTO, 0, len(session.Items))
	for _, ref := range session.Items {
		id := ref.WordID
		doc, ok := wordDocByID[id]
		if !ok {
			continue
		}

		item, err := buildSessionItem(doc, ref.SenseIndex)
		if err != nil {
			continue // Skip corrupted item
		}
		
		// Inject MCQ via Database (Fast, no AI tokens)
		choices, err := s.buildQuizChoices(ctx, doc, ref.SenseIndex)
		if err == nil {
			item.Choices = choices
		}

		// Inject Fill Blank
		if fbExercises != nil && len(fbExercises[id]) > 0 {
			ex := fbExercises[id][0]
			item.FillBlankQuestion = ex.Question
			item.FillBlankAnswer = ex.Answer
		}

		items = append(items, item)
	}


	// Enrich translations and annotations
	if s.enricher != nil {
		targetLang := "vi" // Future: read from session or user config
		enrichedItems, err := s.enricher.EnrichItems(ctx, session.UserID, targetLang, items)
		if err == nil {
			items = enrichedItems
		}
	}

	return &SessionDTO{
		ID:           session.ID.Hex(),
		Mode:         SessionMode(session.Mode),
		Status:       session.Status,
		CurrentIndex: session.CurrentIndex,
		Items:        items,
	}, nil
}

// SubmitSessionResults aggregates evidence and submits it in a transaction
func (s *service) SubmitSessionResults(ctx context.Context, userID string, sessionID string, req *SubmitSessionRequest) error {
	uid, err := bson.ObjectIDFromHex(userID)
	if err != nil {
		return err
	}
	sid, err := bson.ObjectIDFromHex(sessionID)
	if err != nil {
		return err
	}

	// 1. Fetch the session
	session, err := s.repo.GetSessionByID(ctx, sid)
	if err != nil {
		return err
	}
	if session == nil {
		return errors.New("session not found")
	}

	// 2. Idempotency Check: if already completed, just return success
	if session.Status == SessionCompleted {
		return nil
	}
	if session.Status != SessionActive {
		return errors.New("session is not active")
	}

	// Make sure user owns it
	if session.UserID != uid {
		return errors.New("unauthorized session access")
	}

	now := time.Now().UTC()
	var progresses []*LearningProgress
	var reviews []*Review
	totalXP := 0
	
	// Pre-fetch all progresses for these words to update them
	refs := make([]SessionItemRef, len(req.Results))
	for i, r := range req.Results {
		w, _ := bson.ObjectIDFromHex(r.WordID)
		refs[i] = SessionItemRef{WordID: w, SenseIndex: r.SenseIndex}
	}
	progressMap, err := s.repo.GetProgressForWords(ctx, uid, refs)
	if err != nil {
		return err
	}

	// 3. Process each result
	for _, res := range req.Results {
		wid, err := bson.ObjectIDFromHex(res.WordID)
		if err != nil {
			continue
		}

		// A. Evaluate Evidence
		assessment := EvaluateLearningEvidence(res)
		
		// B. Retrieve or create Progress
		key := LearningProgressKey{WordID: wid, SenseIndex: res.SenseIndex}
		p, exists := progressMap[key]
		if !exists {
			p = &LearningProgress{
				ID:         bson.NewObjectID(),
				UserID:     uid,
				WordID:     wid,
				SenseIndex: res.SenseIndex,
				CreatedAt:  now,
			}
		}

		// C. Scheduler Calculation using Final Quality
		quality := Quality(assessment.FinalQuality)
		nextP := s.scheduler.CalculateNext(p, quality)
		nextP.UpdatedAt = now
		progresses = append(progresses, nextP)

		// D. Build Review Log
		rev := &Review{
			ID:             bson.NewObjectID(),
			UserID:         uid,
			WordID:         wid,
			SenseIndex:     res.SenseIndex,
			Quality:        quality,
			OldInterval:    p.Interval,
			NewInterval:    nextP.Interval,
			OldEF:          p.EaseFactor,
			NewEF:          nextP.EaseFactor,
			OldRepetition:  p.Repetition,
			NewRepetition:  nextP.Repetition,
			ReviewedAt:     now,
		}
		reviews = append(reviews, rev)

		// E. Gamification points:
		// Example: 10 XP per word learned/reviewed correctly
		if quality >= QualityHard {
			totalXP += 10
		}
	}

	// 4. Commit Transaction
	err = s.repo.SubmitSessionWithTransaction(ctx, sid, progresses, reviews, totalXP)
	if err != nil {
		return err
	}

	// 5. Update Gamification async
	if s.gamiSvc != nil && totalXP > 0 {
		go func() {
			bgCtx := context.Background()
			_ = s.gamiSvc.TrackActivity(bgCtx, userID, gamification.ActivityReview, true, totalXP)
		}()
	}

	return nil
}

func (s *service) DebugAdvanceTime(ctx context.Context, userID string, days int) error {
	oid, err := bson.ObjectIDFromHex(userID)
	if err != nil {
		return errors.New("invalid user id")
	}
	return s.repo.DebugAdvanceTime(ctx, oid, days)
}

// -----------------------------------------------------------------------------
// Private helpers
// -----------------------------------------------------------------------------

func buildSessionItem(doc *dictionary.WordDocument, senseIndex int) (SessionItemDTO, error) {
	if senseIndex < 0 || senseIndex >= len(doc.Senses) {
		return SessionItemDTO{}, errors.New("invalid sense_index")
	}

	// Extract IPA from first sound
	ipa := ""
	for _, s := range doc.Sounds {
		if s.IPA != "" {
			ipa = s.IPA
			break
		}
	}

	var defs []string
	var example string
	
	sense := doc.Senses[senseIndex]
	if len(sense.Glosses) > 0 && sense.Glosses[0] != "" {
		defs = append(defs, sense.Glosses[0])
		if len(sense.Examples) > 0 {
			example = sense.Examples[0].Text
		}
	}

	return SessionItemDTO{
		WordID:      doc.ID.Hex(),
		SenseIndex:  senseIndex,
		Word:        doc.Word,
		POS:         doc.POS,
		IPA:         ipa,
		Definitions: defs,
		Example:     example,
	}, nil
}

func (s *service) buildQuizChoices(ctx context.Context, doc *dictionary.WordDocument, senseIndex int) ([]ChoiceDTO, error) {
	if senseIndex < 0 || senseIndex >= len(doc.Senses) {
		return nil, errors.New("invalid sense_index")
	}

	// Correct answer: gloss of the specific sense
	correctDef := ""
	sense := doc.Senses[senseIndex]
	if len(sense.Glosses) > 0 && sense.Glosses[0] != "" {
		correctDef = sense.Glosses[0]
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
