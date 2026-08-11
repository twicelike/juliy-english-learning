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
}

func NewService(repo Repository, dictRepo dictionary.Repository, userRepo user.Repository, scheduler Scheduler, gamiSvc gamification.Service, exSvc ExerciseContentService) Service {
	return &service{
		repo:      repo,
		dictRepo:  dictRepo,
		userRepo:  userRepo,
		scheduler: scheduler,
		gamiSvc:   gamiSvc,
		exSvc:     exSvc,
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
	var newAvailableIDs []bson.ObjectID
	if len(pendingIDs) < limit {
		newLimit := limit - len(pendingIDs)
		
		bookmarks, err := s.userRepo.ListBookmarks(ctx, uid)
		if err == nil && len(bookmarks) > 0 {
			bookmarkWords := make([]string, 0, len(bookmarks))
			for _, b := range bookmarks {
				bookmarkWords = append(bookmarkWords, b.Word)
			}
			
			wordDocs, _ := s.dictRepo.GetManyByWords(ctx, bookmarkWords)
			
			allWordIDs := make([]bson.ObjectID, 0)
			seenWords := make(map[string]bool)
			for i := range wordDocs {
				w := wordDocs[i].Word
				if seenWords[w] {
					continue
				}
				seenWords[w] = true
				allWordIDs = append(allWordIDs, wordDocs[i].ID)
			}
			
			progressMap, _ := s.repo.GetProgressForWords(ctx, uid, allWordIDs)
			for _, id := range allWordIDs {
				if _, exists := progressMap[id]; !exists {
					newAvailableIDs = append(newAvailableIDs, id)
					if len(newAvailableIDs) == newLimit {
						break
					}
				}
			}
		}
	}

	if len(pendingIDs) == 0 && len(newAvailableIDs) == 0 {
		return nil, errors.New("no_new_words")
	}

	// 4. Combine words for the session (Snapshot)
	sessionWordIDs := append(pendingIDs, newAvailableIDs...)

	now := time.Now()
	// 5. Create session in "preparing" state. Do NOT create progress yet.
	session := &LearningSession{
		ID:           bson.NewObjectID(),
		UserID:       uid,
		Mode:         "new_words",
		WordIDs:      sessionWordIDs,
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

	// 6. Asynchronously trigger content preparation
	go func(sid bson.ObjectID, wids []bson.ObjectID) {
		bgCtx := context.Background()
		if s.exSvc != nil {
			_, _ = s.exSvc.GetExercises(bgCtx, wids, ExerciseMultipleChoice, 1)
			_, _ = s.exSvc.GetExercises(bgCtx, wids, ExerciseFillBlank, 1)
		}
		_ = s.repo.UpdateSessionStatus(bgCtx, sid, SessionActive)
	}(session.ID, sessionWordIDs)

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
		WordIDs:      dueIDs,
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
	go func(sid bson.ObjectID, wids []bson.ObjectID) {
		bgCtx := context.Background()
		if s.exSvc != nil {
			_, _ = s.exSvc.GetExercises(bgCtx, wids, ExerciseMultipleChoice, 1)
			_, _ = s.exSvc.GetExercises(bgCtx, wids, ExerciseFillBlank, 1)
		}
		_ = s.repo.UpdateSessionStatus(bgCtx, sid, SessionActive)
	}(session.ID, dueIDs)

	// 5. Build and return DTO
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
	// Resolve dictionary words
	wordDocs, err := s.dictRepo.GetManyByIDs(ctx, session.WordIDs)
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
		fbExercises, _ = s.exSvc.GetExercises(ctx, session.WordIDs, ExerciseFillBlank, 1)
	}

	items := make([]SessionItemDTO, 0, len(session.WordIDs))
	for _, id := range session.WordIDs {
		doc, ok := wordDocByID[id]
		if !ok {
			continue
		}

		item := buildSessionItem(doc)
		
		// Inject MCQ via Database (Fast, no AI tokens)
		choices, err := s.buildQuizChoices(ctx, doc)
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
	wordIDs := make([]bson.ObjectID, len(req.Results))
	for i, r := range req.Results {
		w, _ := bson.ObjectIDFromHex(r.WordID)
		wordIDs[i] = w
	}
	progressMap, err := s.repo.GetProgressForWords(ctx, uid, wordIDs)
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
		p, exists := progressMap[wid]
		if !exists {
			p = &LearningProgress{
				ID:        bson.NewObjectID(),
				UserID:    uid,
				WordID:    wid,
				CreatedAt: now,
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
