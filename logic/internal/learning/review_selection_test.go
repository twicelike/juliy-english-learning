package learning

import (
	"testing"

	"go.mongodb.org/mongo-driver/v2/bson"
)

func generateObjectIDs(count int) []bson.ObjectID {
	var ids []bson.ObjectID
	for i := 0; i < count; i++ {
		ids = append(ids, bson.NewObjectID())
	}
	return ids
}

func TestSelectReviewWords_InvalidLimit(t *testing.T) {
	_, err := SelectReviewWords(nil, nil, nil, ReviewStrategyContinue, 3)
	if err != ErrInvalidWordCount {
		t.Errorf("expected ErrInvalidWordCount, got %v", err)
	}

	_, err = SelectReviewWords(nil, nil, nil, ReviewStrategyContinue, 30)
	if err != ErrInvalidWordCount {
		t.Errorf("expected ErrInvalidWordCount, got %v", err)
	}
}

func TestSelectReviewWords_Empty(t *testing.T) {
	_, err := SelectReviewWords(nil, nil, nil, ReviewStrategyContinue, 10)
	if err != ErrNoReviewableWords {
		t.Errorf("expected ErrNoReviewableWords, got %v", err)
	}
}

func TestSelectReviewWords_Fresh(t *testing.T) {
	due := generateObjectIDs(5)
	learning := generateObjectIDs(15)
	newWords := generateObjectIDs(25)

	selected, err := SelectReviewWords(due, learning, newWords, ReviewStrategyFresh, 25)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(selected) != 25 {
		t.Errorf("expected 25 selected words, got %d", len(selected))
	}

	// Should prioritize due (5), then new (20)
	dueCount := 0
	learningCount := 0
	newCount := 0

	// Helper to count selected types
	for _, id := range selected {
		if contains(due, id) {
			dueCount++
		} else if contains(learning, id) {
			learningCount++
		} else if contains(newWords, id) {
			newCount++
		}
	}

	if dueCount != 5 {
		t.Errorf("expected 5 due words, got %d", dueCount)
	}
	if learningCount != 0 {
		t.Errorf("expected 0 learning words in Fresh strategy, got %d", learningCount)
	}
	if newCount != 20 {
		t.Errorf("expected 20 new words, got %d", newCount)
	}
}

func TestSelectReviewWords_Continue(t *testing.T) {
	due := generateObjectIDs(5)
	learning := generateObjectIDs(15)
	newWords := generateObjectIDs(25)

	selected, err := SelectReviewWords(due, learning, newWords, ReviewStrategyContinue, 25)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	dueCount := 0
	learningCount := 0
	newCount := 0

	for _, id := range selected {
		if contains(due, id) {
			dueCount++
		} else if contains(learning, id) {
			learningCount++
		} else if contains(newWords, id) {
			newCount++
		}
	}

	if dueCount != 5 {
		t.Errorf("expected 5 due words, got %d", dueCount)
	}
	if learningCount != 15 {
		t.Errorf("expected 15 learning words, got %d", learningCount)
	}
	if newCount != 5 {
		t.Errorf("expected 5 new words, got %d", newCount)
	}
}

func TestSelectReviewWords_FallbackToAvailable(t *testing.T) {
	due := generateObjectIDs(2)
	learning := generateObjectIDs(3)
	newWords := generateObjectIDs(3)

	selected, err := SelectReviewWords(due, learning, newWords, ReviewStrategyContinue, 25)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(selected) != 8 {
		t.Errorf("expected 8 selected words (all available), got %d", len(selected))
	}
}

func contains(slice []bson.ObjectID, item bson.ObjectID) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
