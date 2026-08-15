package wordset

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/english/backend/internal/auth"
	"github.com/english/backend/internal/dictionary"
	"github.com/english/backend/internal/learning"
	"github.com/english/backend/internal/language"
	"github.com/english/backend/internal/translation"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type AccessResult struct {
	Set     *WordSet
	IsOwner bool
}

type service struct {
	repo      Repository
	authRepo  auth.Repository
	dictSvc   dictionary.Service
	learnRepo learning.Repository
	transSvc  translation.Service
}

func NewService(repo Repository, authRepo auth.Repository, dictSvc dictionary.Service, learnRepo learning.Repository, transSvc translation.Service) Service {
	return &service{repo: repo, authRepo: authRepo, dictSvc: dictSvc, learnRepo: learnRepo, transSvc: transSvc}
}

func (s *service) checkOwnership(ctx context.Context, userID, setID bson.ObjectID) (*WordSet, error) {
	set, err := s.repo.GetWordSetByID(ctx, setID)
	if err != nil {
		return nil, err
	}
	if set == nil {
		return nil, fmt.Errorf("wordset not found") // 404 in handler
	}
	if set.UserID != userID {
		return nil, fmt.Errorf("unauthorized: you don't own this wordset") // 403 or 404 in handler
	}
	return set, nil
}

func (s *service) checkAccess(ctx context.Context, userID, setID bson.ObjectID) (AccessResult, error) {
	set, err := s.repo.GetWordSetByID(ctx, setID)
	if err != nil {
		return AccessResult{}, err
	}
	if set == nil {
		return AccessResult{}, fmt.Errorf("wordset not found")
	}

	isOwner := set.UserID == userID
	if isOwner {
		return AccessResult{Set: set, IsOwner: true}, nil
	}

	if set.Visibility == WordSetVisibilityPublic {
		return AccessResult{Set: set, IsOwner: false}, nil
	}

	// For private sets owned by others, return not found to avoid information leakage
	return AccessResult{}, fmt.Errorf("wordset not found")
}

func (s *service) CreateWordSet(ctx context.Context, userID bson.ObjectID, req CreateWordSetRequest) (*WordSet, error) {
	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" || len(req.Name) > 100 {
		return nil, fmt.Errorf("wordset name must be between 1 and 100 characters")
	}
	if len(req.Description) > 500 {
		return nil, fmt.Errorf("description cannot exceed 500 characters")
	}
	if req.Visibility != WordSetVisibilityPrivate && req.Visibility != WordSetVisibilityPublic {
		req.Visibility = WordSetVisibilityPrivate
	}

	if req.SourceLanguage == "" {
		req.SourceLanguage = "en"
	}
	if req.AnnotationLanguage == "" {
		req.AnnotationLanguage = "vi"
	}
	if !language.IsSupported(req.SourceLanguage) {
		return nil, fmt.Errorf("unsupported source language: %s", req.SourceLanguage)
	}
	if !language.IsSupported(req.AnnotationLanguage) {
		return nil, fmt.Errorf("unsupported annotation language: %s", req.AnnotationLanguage)
	}

	now := time.Now()
	set := &WordSet{
		ID:                 bson.NewObjectID(),
		UserID:             userID,
		Name:               req.Name,
		Description:        req.Description,
		Visibility:         req.Visibility,
		Topics:             req.Topics,
		Difficulty:         req.Difficulty,
		SourceLanguage:     req.SourceLanguage,
		AnnotationLanguage: req.AnnotationLanguage,
		WordCount:          0,
		CloneCount:         0,
		CreatedAt:          now,
		UpdatedAt:          now,
	}

	err := s.repo.CreateWordSet(ctx, set)
	if err != nil {
		return nil, err
	}
	return set, nil
}

func (s *service) GetWordSetsByUserID(ctx context.Context, userID bson.ObjectID, containsWordIDs []bson.ObjectID) (*WordSetListResponse, error) {
	sets, err := s.repo.GetWordSetsByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	
	// Map from setID → saved sense refs for the queried word(s)
	containsMap := make(map[string][]SavedSenseRef)
	if len(containsWordIDs) > 0 {
		items, _ := s.repo.GetWordSetsContainingWord(ctx, userID, containsWordIDs)
		for _, it := range items {
			setIDStr := it.SetID.Hex()
			containsMap[setIDStr] = append(containsMap[setIDStr], SavedSenseRef{
				WordID:     it.WordID.Hex(),
				SenseIndex: it.SenseIndex,
			})
		}
	}

	var listItems []WordSetListItem
	for _, set := range sets {
		setIDStr := set.ID.Hex()
		savedSenses := containsMap[setIDStr]
		listItems = append(listItems, WordSetListItem{
			WordSet:      set,
			ContainsWord: len(savedSenses) > 0,
			SavedSenses:  savedSenses,
		})
	}

	if listItems == nil {
		listItems = []WordSetListItem{}
	}

	return &WordSetListResponse{
		Items: listItems,
		Total: len(listItems),
	}, nil
}

func (s *service) ExploreWordSets(ctx context.Context, req ExploreWordSetsRequest) (*ExploreWordSetsResponse, error) {
	sets, total, err := s.repo.ExploreWordSets(ctx, req)
	if err != nil {
		return nil, err
	}

	var listItems []WordSetListItem
	for _, set := range sets {
		creatorName := "Unknown User"
		creator, err := s.authRepo.GetUserByID(ctx, set.UserID)
		if err == nil && creator != nil {
			creatorName = creator.Name
		}

		listItem := WordSetListItem{
			WordSet:     set,
			CreatorName: creatorName,
		}
		listItems = append(listItems, listItem)
	}

	if listItems == nil {
		listItems = []WordSetListItem{}
	}

	limit := req.Limit
	if limit < 1 {
		limit = 20
	}
	page := req.Page
	if page < 1 {
		page = 1
	}

	hasMore := (page * limit) < total

	return &ExploreWordSetsResponse{
		Items:   listItems,
		Page:    page,
		Limit:   limit,
		Total:   total,
		HasMore: hasMore,
	}, nil
}

func (s *service) GetWordSetDetail(ctx context.Context, userID, setID bson.ObjectID) (*WordSetDetailResponse, error) {
	access, err := s.checkAccess(ctx, userID, setID)
	if err != nil {
		return nil, err
	}

	items, err := s.repo.GetWordSetItems(ctx, setID)
	if err != nil {
		return nil, err
	}

	itemIDs := make([]string, 0, len(items))
	savedSenses := make([]SavedSenseRef, 0, len(items))
	for _, it := range items {
		itemIDs = append(itemIDs, it.WordID.Hex())
		savedSenses = append(savedSenses, SavedSenseRef{
			WordID:     it.WordID.Hex(),
			SenseIndex: it.SenseIndex,
		})
	}

	creatorName := ""
	if !access.IsOwner {
		creator, err := s.authRepo.GetUserByID(ctx, access.Set.UserID)
		if err == nil && creator != nil {
			creatorName = creator.Name
		} else {
			creatorName = "Unknown User"
		}
	}

	return &WordSetDetailResponse{
		WordSet:     *access.Set,
		ItemIDs:     itemIDs,
		SavedSenses: savedSenses,
		IsOwner:     access.IsOwner,
		CreatorName: creatorName,
	}, nil
}

func (s *service) CloneWordSet(ctx context.Context, userID, setID bson.ObjectID) (*WordSet, error) {
	access, err := s.checkAccess(ctx, userID, setID)
	if err != nil {
		return nil, err
	}

	if access.IsOwner {
		return nil, fmt.Errorf("you already own this word set")
	}

	sourceSet := access.Set
	now := time.Now()

	newSet := &WordSet{
		ID:          bson.NewObjectID(),
		UserID:      userID,
		Name:        sourceSet.Name + " (Copy)",
		Description: sourceSet.Description,
		Visibility:  WordSetVisibilityPrivate,
		WordCount:   0, // Let CreateWordSetItems automatically update the count
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	err = s.repo.CreateWordSet(ctx, newSet)
	if err != nil {
		return nil, err
	}

	items, err := s.repo.GetWordSetItems(ctx, setID)
	if err != nil {
		return nil, err
	}

	if len(items) > 0 {
		var newItems []WordSetItem
		for _, it := range items {
			newItems = append(newItems, WordSetItem{
				SetID:      newSet.ID,
				WordID:     it.WordID,
				SenseIndex: it.SenseIndex,
				Position:   it.Position,
				CreatedAt:  now,
			})
		}
		_, _ = s.repo.CreateWordSetItems(ctx, newItems)
	}

	// Increment source clone count
	_ = s.repo.UpdateWordSet(ctx, setID, bson.M{"clone_count": sourceSet.CloneCount + 1})

	// Refresh newSet to get the updated WordCount
	finalSet, err := s.repo.GetWordSetByID(ctx, newSet.ID)
	if err == nil {
		newSet = finalSet
	}

	return newSet, nil
}

func (s *service) UpdateWordSet(ctx context.Context, userID, setID bson.ObjectID, req UpdateWordSetRequest) (*WordSet, error) {
	set, err := s.checkOwnership(ctx, userID, setID)
	if err != nil {
		return nil, err
	}

	update := bson.M{"updated_at": time.Now()}
	
	if req.Name != nil {
		name := strings.TrimSpace(*req.Name)
		if name == "" || len(name) > 100 {
			return nil, fmt.Errorf("wordset name must be between 1 and 100 characters")
		}
		update["name"] = name
		set.Name = name
	}
	
	if req.Description != nil {
		if len(*req.Description) > 500 {
			return nil, fmt.Errorf("description cannot exceed 500 characters")
		}
		update["description"] = *req.Description
		set.Description = *req.Description
	}
	
	if req.Visibility != nil {
		if *req.Visibility != WordSetVisibilityPrivate && *req.Visibility != WordSetVisibilityPublic {
			return nil, fmt.Errorf("invalid visibility")
		}
		update["visibility"] = *req.Visibility
		set.Visibility = *req.Visibility
	}
	
	if req.Topics != nil {
		update["topics"] = *req.Topics
		set.Topics = *req.Topics
	}

	if req.Difficulty != nil {
		update["difficulty"] = *req.Difficulty
		set.Difficulty = *req.Difficulty
	}

	if req.AnnotationLanguage != nil {
		lang := strings.TrimSpace(*req.AnnotationLanguage)
		if lang == "" {
			return nil, fmt.Errorf("annotation language cannot be empty")
		}
		if !language.IsSupported(lang) {
			return nil, fmt.Errorf("unsupported annotation language: %s", lang)
		}
		update["annotation_language"] = lang
		set.AnnotationLanguage = lang
	}

	err = s.repo.UpdateWordSet(ctx, setID, update)
	if err != nil {
		return nil, err
	}
	set.UpdatedAt = update["updated_at"].(time.Time)
	return set, nil
}

func (s *service) DeleteWordSet(ctx context.Context, userID, setID bson.ObjectID) error {
	_, err := s.checkOwnership(ctx, userID, setID)
	if err != nil {
		return err
	}
	return s.repo.DeleteWordSetWithItems(ctx, setID)
}

func (s *service) AddWord(ctx context.Context, userID, setID, wordID bson.ObjectID, senseIndex int) error {
	_, err := s.checkOwnership(ctx, userID, setID)
	if err != nil {
		return err
	}

	items, err := s.repo.GetWordSetItems(ctx, setID)
	if err != nil {
		return err
	}
	pos := len(items)

	item := &WordSetItem{
		SetID:      setID,
		WordID:     wordID,
		SenseIndex: senseIndex,
		Position:   pos,
		CreatedAt:  time.Now(),
	}

	err = s.repo.CreateWordSetItem(ctx, item)
	if err == nil {
		set, _ := s.repo.GetWordSetByID(ctx, setID)
		if set != nil && s.transSvc != nil {
			_ = s.transSvc.EnqueueJobs(ctx, []string{wordID.Hex()}, set.AnnotationLanguage)
		}
	}
	
	return err
}

func (s *service) BulkAddWords(ctx context.Context, userID, setID bson.ObjectID, req BulkAddWordsRequest) (*BulkAddWordsResponse, error) {
	set, err := s.checkOwnership(ctx, userID, setID)
	if err != nil {
		return nil, err
	}

	if len(req.Items) == 0 {
		return nil, fmt.Errorf("items cannot be empty")
	}

	limit := 50
	if set.WordCount+len(req.Items) > limit {
		return nil, fmt.Errorf("Word set can contain at most %d words. Current: %d, Request: %d", limit, set.WordCount, len(req.Items))
	}

	existingItems, err := s.repo.GetWordSetItems(ctx, setID)
	if err != nil {
		return nil, err
	}

	existingMap := make(map[string]bool)
	for _, it := range existingItems {
		key := fmt.Sprintf("%s:%d", it.WordID.Hex(), it.SenseIndex)
		existingMap[key] = true
	}

	var newWordIDs []bson.ObjectID
	var itemsToInsert []WordSetItem
	pos := len(existingItems)

	for _, reqItem := range req.Items {
		wID, err := bson.ObjectIDFromHex(reqItem.WordID)
		if err != nil {
			return nil, fmt.Errorf("invalid word id: %s", reqItem.WordID)
		}
		// Generate composite key for existingMap check (wordID + senseIndex)
		key := fmt.Sprintf("%s:%d", reqItem.WordID, reqItem.SenseIndex)
		if existingMap[key] {
			continue // skip silently
		}
		newWordIDs = append(newWordIDs, wID)
		itemsToInsert = append(itemsToInsert, WordSetItem{
			SetID:      setID,
			WordID:     wID,
			SenseIndex: reqItem.SenseIndex,
			Position:   pos,
			CreatedAt:  time.Now(),
		})
		pos++
		existingMap[key] = true
	}

	if len(itemsToInsert) > 0 {
		// Deduplicate newWordIDs for verification
		uniqueWordIDs := make([]bson.ObjectID, 0)
		seenIDs := make(map[bson.ObjectID]bool)
		for _, id := range newWordIDs {
			if !seenIDs[id] {
				seenIDs[id] = true
				uniqueWordIDs = append(uniqueWordIDs, id)
			}
		}

		// Verify words exist via dictionary service (which queries the correct "dictionary" db)
		dtos, err := s.dictSvc.GetManyByIDs(ctx, uniqueWordIDs)
		if err != nil {
			return nil, fmt.Errorf("failed to verify words: %w", err)
		}
		if len(dtos) != len(uniqueWordIDs) {
			return nil, fmt.Errorf("some word IDs do not exist in the dictionary")
		}

		added, err := s.repo.CreateWordSetItems(ctx, itemsToInsert)
		if err != nil {
			return nil, err
		}
		set.WordCount += added
		
		// Enqueue translation jobs
		var wordIDStrs []string
		for _, id := range uniqueWordIDs {
			wordIDStrs = append(wordIDStrs, id.Hex())
		}
		if len(wordIDStrs) > 0 && s.transSvc != nil {
			_ = s.transSvc.EnqueueJobs(ctx, wordIDStrs, set.AnnotationLanguage)
		}
		
		return &BulkAddWordsResponse{
			Added:     added,
			WordCount: set.WordCount,
			Limit:     limit,
		}, nil
	}

	return &BulkAddWordsResponse{
		Added:     0,
		WordCount: set.WordCount,
		Limit:     limit,
	}, nil
}

func (s *service) RemoveWord(ctx context.Context, userID, setID, wordID bson.ObjectID) error {
	_, err := s.checkOwnership(ctx, userID, setID)
	if err != nil {
		return err
	}
	return s.repo.RemoveWordSetItem(ctx, setID, wordID)
}

func (s *service) BatchDeleteWords(ctx context.Context, userID, setID bson.ObjectID, req BatchDeleteWordsRequest) (*BatchDeleteWordsResponse, error) {
	set, err := s.checkOwnership(ctx, userID, setID)
	if err != nil {
		return nil, err
	}
	if len(req.WordIDs) == 0 {
		return nil, fmt.Errorf("word_ids must not be empty")
	}

	// Parse and deduplicate word IDs
	seen := make(map[bson.ObjectID]bool)
	var wordIDs []bson.ObjectID
	for _, idStr := range req.WordIDs {
		wID, err := bson.ObjectIDFromHex(idStr)
		if err != nil {
			return nil, fmt.Errorf("invalid word id: %s", idStr)
		}
		if !seen[wID] {
			seen[wID] = true
			wordIDs = append(wordIDs, wID)
		}
	}

	deletedCount, err := s.repo.BatchRemoveWordSetItems(ctx, setID, wordIDs)
	if err != nil {
		return nil, err
	}

	// Build deleted_word_ids list (those that actually existed in the set)
	// We return all requested IDs that were present (deletedCount is the truth)
	deletedIDs := make([]string, 0, deletedCount)
	for _, wID := range wordIDs {
		deletedIDs = append(deletedIDs, wID.Hex())
		if int64(len(deletedIDs)) >= deletedCount {
			break
		}
	}

	newCount := set.WordCount - int(deletedCount)
	if newCount < 0 {
		newCount = 0
	}

	return &BatchDeleteWordsResponse{
		DeletedWordIDs: deletedIDs,
		WordCount:      newCount,
	}, nil
}

func (s *service) RestoreWords(ctx context.Context, userID, setID bson.ObjectID, req RestoreWordsRequest) (*RestoreWordsResponse, error) {
	set, err := s.checkOwnership(ctx, userID, setID)
	if err != nil {
		return nil, err
	}
	if len(req.WordIDs) == 0 {
		return nil, fmt.Errorf("word_ids must not be empty")
	}

	// Parse and deduplicate
	seen := make(map[bson.ObjectID]bool)
	var wordIDs []bson.ObjectID
	for _, idStr := range req.WordIDs {
		wID, err := bson.ObjectIDFromHex(idStr)
		if err != nil {
			return nil, fmt.Errorf("invalid word id: %s", idStr)
		}
		if !seen[wID] {
			seen[wID] = true
			wordIDs = append(wordIDs, wID)
		}
	}

	// Get existing items to avoid duplicates
	existingItems, err := s.repo.GetWordSetItems(ctx, setID)
	if err != nil {
		return nil, err
	}
	existingMap := make(map[bson.ObjectID]bool)
	for _, it := range existingItems {
		existingMap[it.WordID] = true
	}

	// Build items to insert (skip duplicates already in set)
	now := time.Now()
	pos := len(existingItems)
	var toInsert []WordSetItem
	var restoredIDs []string

	for _, wID := range wordIDs {
		if existingMap[wID] {
			continue // already in set, skip
		}
		toInsert = append(toInsert, WordSetItem{
			SetID:      setID,
			WordID:     wID,
			SenseIndex: 0, // restore with original sense — caller must pass correct IDs
			Position:   pos,
			CreatedAt:  now,
		})
		restoredIDs = append(restoredIDs, wID.Hex())
		existingMap[wID] = true
		pos++
	}

	if len(toInsert) > 0 {
		added, err := s.repo.CreateWordSetItems(ctx, toInsert)
		if err != nil {
			return nil, err
		}
		_ = added
	}

	newCount := set.WordCount + len(restoredIDs)

	return &RestoreWordsResponse{
		RestoredWordIDs: restoredIDs,
		WordCount:       newCount,
	}, nil
}

func (s *service) ReorderWords(ctx context.Context, userID, setID bson.ObjectID, req ReorderWordsRequest) error {
	_, err := s.checkOwnership(ctx, userID, setID)
	if err != nil {
		return err
	}

	items, err := s.repo.GetWordSetItems(ctx, setID)
	if err != nil {
		return err
	}
	if len(req.WordIDs) != len(items) {
		return fmt.Errorf("provided word_ids count does not match the items in the set")
	}

	existingWords := make(map[string]bool)
	for _, it := range items {
		existingWords[it.WordID.Hex()] = true
	}

	var newOrder []bson.ObjectID
	seen := make(map[string]bool)
	for _, idStr := range req.WordIDs {
		if seen[idStr] {
			return fmt.Errorf("duplicate word_id in reorder request: %s", idStr)
		}
		if !existingWords[idStr] {
			return fmt.Errorf("word_id not found in this word set: %s", idStr)
		}
		seen[idStr] = true
		wID, _ := bson.ObjectIDFromHex(idStr)
		newOrder = append(newOrder, wID)
	}

	return s.repo.ReorderWordSetItems(ctx, setID, newOrder)
}

func (s *service) GetWordSetWords(ctx context.Context, userID, setID bson.ObjectID) ([]WordViewerItemDTO, error) {
	access, err := s.checkAccess(ctx, userID, setID)
	if err != nil {
		return nil, err
	}

	items, err := s.repo.GetWordSetItems(ctx, setID)
	if err != nil {
		return nil, err
	}

	if len(items) == 0 {
		return []WordViewerItemDTO{}, nil
	}

	var wordIDs []bson.ObjectID
	var itemRefs []learning.SessionItemRef
	for _, item := range items {
		wordIDs = append(wordIDs, item.WordID)
		itemRefs = append(itemRefs, learning.SessionItemRef{WordID: item.WordID, SenseIndex: item.SenseIndex})
	}

	dictWords, err := s.dictSvc.GetManyByIDs(ctx, wordIDs)
	if err != nil {
		return nil, err
	}
	
	progressMap, err := s.learnRepo.GetProgressForWords(ctx, userID, itemRefs)
	if err != nil {
		// Log error, but don't fail the entire view request
		progressMap = make(map[learning.LearningProgressKey]*learning.LearningProgress)
	}

	// Create map for easy lookup of WordDTO by ID
	dictMap := make(map[bson.ObjectID]*dictionary.WordDTO)
	for i := range dictWords {
		objID, _ := bson.ObjectIDFromHex(dictWords[i].ID)
		dictMap[objID] = &dictWords[i]
	}

	// WA9: Fetch Translations
	var translations []translation.WordTranslation
	if s.transSvc != nil {
		translations, _ = s.transSvc.GetTranslationsByWordIDs(ctx, wordIDs, access.Set.AnnotationLanguage)
	}
	
	transMap := make(map[string]translation.WordTranslation)
	for _, tr := range translations {
		key := fmt.Sprintf("%s:%d", tr.WordID.Hex(), tr.SenseIndex)
		transMap[key] = tr
	}

	var viewerItems []WordViewerItemDTO
	for _, item := range items {
		dw, ok := dictMap[item.WordID]
		if !ok {
			continue
		}

		key := learning.LearningProgressKey{WordID: item.WordID, SenseIndex: item.SenseIndex}
		state := learning.DetermineState(progressMap[key])
		
		candidates := []string{}
		transKey := fmt.Sprintf("%s:%d", item.WordID.Hex(), item.SenseIndex)
		if tr, ok := transMap[transKey]; ok {
			candidates = tr.Translations
		}
		
		viewerItems = append(viewerItems, WordViewerItemDTO{
			WordDTO:               *dw,
			SenseIndex:            item.SenseIndex,
			ProgressState:         state,
			TranslationCandidates: candidates,
			Annotation:            item.Annotation,
		})
	}

	return viewerItems, nil
}

func (s *service) UpdateAnnotation(ctx context.Context, userID, setID, wordID bson.ObjectID, text string) (*UserAnnotation, error) {
	set, err := s.checkOwnership(ctx, userID, setID)
	if err != nil {
		return nil, err
	}

	trimmedText := strings.TrimSpace(text)
	
	// If text is empty, clear annotation
	if trimmedText == "" {
		err := s.repo.UpdateWordSetItemAnnotation(ctx, setID, wordID, nil)
		return nil, err
	}

	// Determine Source (ai vs manual)
	source := "manual"
	if s.transSvc != nil {
		translations, err := s.transSvc.GetTranslationsByWordIDs(ctx, []bson.ObjectID{wordID}, set.AnnotationLanguage)
		if err == nil && len(translations) > 0 {
			for _, c := range translations[0].Translations {
				if strings.EqualFold(strings.TrimSpace(c), trimmedText) {
					source = "ai"
					trimmedText = c // Use the exact case from AI
					break
				}
			}
		}
	}

	annotation := &UserAnnotation{
		LanguageCode: set.AnnotationLanguage,
		Text:         trimmedText,
		Source:       source,
	}

	err = s.repo.UpdateWordSetItemAnnotation(ctx, setID, wordID, annotation)
	if err != nil {
		return nil, err
	}
	
	return annotation, nil
}

func (s *service) GetAnnotationsByUserAndWordIDs(ctx context.Context, userID bson.ObjectID, wordIDs []bson.ObjectID) (map[bson.ObjectID]*UserAnnotation, error) {
	items, err := s.repo.GetWordSetsContainingWord(ctx, userID, wordIDs)
	if err != nil {
		return nil, err
	}
	
	res := make(map[bson.ObjectID]*UserAnnotation)
	for _, item := range items {
		if item.Annotation != nil {
			// In case a word belongs to multiple sets, we just keep the first one we find or the latest.
			// Currently just overwrite, which is fine since they usually use the same annotation.
			res[item.WordID] = item.Annotation
		}
	}
	return res, nil
}
