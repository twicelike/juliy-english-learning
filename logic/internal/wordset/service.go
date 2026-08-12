package wordset

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/english/backend/internal/dictionary"
	"github.com/english/backend/internal/learning"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type service struct {
	repo      Repository
	dictSvc   dictionary.Service
	learnRepo learning.Repository
}

func NewService(repo Repository, dictSvc dictionary.Service, learnRepo learning.Repository) Service {
	return &service{repo: repo, dictSvc: dictSvc, learnRepo: learnRepo}
}

func (s *service) checkOwnership(ctx context.Context, userID, setID bson.ObjectID) (*WordSet, error) {
	set, err := s.repo.GetWordSetByID(ctx, setID)
	if err != nil {
		return nil, err
	}
	if set == nil {
		return nil, fmt.Errorf("wordset not found")
	}
	if set.UserID != userID {
		return nil, fmt.Errorf("unauthorized: you don't own this wordset")
	}
	return set, nil
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

	now := time.Now()
	set := &WordSet{
		ID:          bson.NewObjectID(),
		UserID:      userID,
		Name:        req.Name,
		Description: req.Description,
		Visibility:  req.Visibility,
		WordCount:   0,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	err := s.repo.CreateWordSet(ctx, set)
	if err != nil {
		return nil, err
	}
	return set, nil
}

func (s *service) GetWordSetsByUserID(ctx context.Context, userID bson.ObjectID, containsWordID *bson.ObjectID) (*WordSetListResponse, error) {
	sets, err := s.repo.GetWordSetsByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	
	containsMap := make(map[string]bool)
	if containsWordID != nil {
		containingSets, _ := s.repo.GetWordSetsContainingWord(ctx, userID, *containsWordID)
		for _, id := range containingSets {
			containsMap[id.Hex()] = true
		}
	}

	var listItems []WordSetListItem
	for _, set := range sets {
		listItems = append(listItems, WordSetListItem{
			WordSet:      set,
			ContainsWord: containsMap[set.ID.Hex()],
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

func (s *service) GetWordSetDetail(ctx context.Context, userID, setID bson.ObjectID) (*WordSetDetailResponse, error) {
	set, err := s.checkOwnership(ctx, userID, setID)
	if err != nil {
		return nil, err
	}

	items, err := s.repo.GetWordSetItems(ctx, setID)
	if err != nil {
		return nil, err
	}

	itemIDs := make([]string, 0, len(items))
	for _, it := range items {
		itemIDs = append(itemIDs, it.WordID.Hex())
	}

	return &WordSetDetailResponse{
		WordSet: *set,
		ItemIDs: itemIDs,
	}, nil
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

func (s *service) AddWord(ctx context.Context, userID, setID, wordID bson.ObjectID) error {
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
		SetID:     setID,
		WordID:    wordID,
		Position:  pos,
		CreatedAt: time.Now(),
	}

	return s.repo.CreateWordSetItem(ctx, item)
}

func (s *service) BulkAddWords(ctx context.Context, userID, setID bson.ObjectID, req BulkAddWordsRequest) (*BulkAddWordsResponse, error) {
	set, err := s.checkOwnership(ctx, userID, setID)
	if err != nil {
		return nil, err
	}

	if len(req.WordIDs) == 0 {
		return nil, fmt.Errorf("word_ids cannot be empty")
	}

	limit := 50
	if set.WordCount+len(req.WordIDs) > limit {
		return nil, fmt.Errorf("Word set can contain at most %d words. Current: %d, Request: %d", limit, set.WordCount, len(req.WordIDs))
	}

	existingItems, err := s.repo.GetWordSetItems(ctx, setID)
	if err != nil {
		return nil, err
	}

	existingMap := make(map[string]bool)
	for _, it := range existingItems {
		existingMap[it.WordID.Hex()] = true
	}

	var newWordIDs []bson.ObjectID
	var itemsToInsert []WordSetItem
	pos := len(existingItems)

	for _, idStr := range req.WordIDs {
		wID, err := bson.ObjectIDFromHex(idStr)
		if err != nil {
			return nil, fmt.Errorf("invalid word id: %s", idStr)
		}
		if existingMap[idStr] {
			continue // skip silently
		}
		newWordIDs = append(newWordIDs, wID)
		itemsToInsert = append(itemsToInsert, WordSetItem{
			SetID:     setID,
			WordID:    wID,
			Position:  pos,
			CreatedAt: time.Now(),
		})
		pos++
		existingMap[idStr] = true
	}

	if len(itemsToInsert) > 0 {
		// Verify words exist via dictionary service (which queries the correct "dictionary" db)
		dtos, err := s.dictSvc.GetManyByIDs(ctx, newWordIDs)
		if err != nil {
			return nil, fmt.Errorf("failed to verify words: %w", err)
		}
		if len(dtos) != len(newWordIDs) {
			return nil, fmt.Errorf("some word IDs do not exist in the dictionary")
		}

		added, err := s.repo.CreateWordSetItems(ctx, itemsToInsert)
		if err != nil {
			return nil, err
		}
		set.WordCount += added
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
	_, err := s.checkOwnership(ctx, userID, setID)
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
	for _, item := range items {
		wordIDs = append(wordIDs, item.WordID)
	}

	dictWords, err := s.dictSvc.GetManyByIDs(ctx, wordIDs)
	if err != nil {
		return nil, err
	}
	
	progressMap, err := s.learnRepo.GetProgressForWords(ctx, userID, wordIDs)
	if err != nil {
		// Log error, but don't fail the entire view request
		progressMap = make(map[bson.ObjectID]*learning.LearningProgress)
	}

	var viewerItems []WordViewerItemDTO
	for _, dw := range dictWords {
		var state string
		objID, err := bson.ObjectIDFromHex(dw.ID)
		if err == nil {
			state = learning.DetermineState(progressMap[objID])
		} else {
			state = "new"
		}
		
		viewerItems = append(viewerItems, WordViewerItemDTO{
			WordDTO:       dw,
			ProgressState: state,
		})
	}

	return viewerItems, nil
}
