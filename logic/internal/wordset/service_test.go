package wordset_test

import (
	"context"
	"testing"

	"github.com/english/backend/internal/wordset"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type mockRepo struct {
	wordsets []wordset.WordSet
	items    []wordset.WordSetItem
}

func (m *mockRepo) CreateWordSet(ctx context.Context, set *wordset.WordSet) error {
	m.wordsets = append(m.wordsets, *set)
	return nil
}

func (m *mockRepo) GetWordSetsByUserID(ctx context.Context, userID bson.ObjectID) ([]wordset.WordSet, error) {
	return m.wordsets, nil
}

func (m *mockRepo) GetWordSetByID(ctx context.Context, id bson.ObjectID) (*wordset.WordSet, error) {
	return nil, nil // placeholder
}

func (m *mockRepo) UpdateWordSet(ctx context.Context, id bson.ObjectID, update bson.M) error {
	return nil
}

func (m *mockRepo) DeleteWordSetWithItems(ctx context.Context, id bson.ObjectID) error {
	return nil
}

func (m *mockRepo) CreateWordSetItems(ctx context.Context, items []wordset.WordSetItem) (int, error) {
	m.items = append(m.items, items...)
	return len(items), nil
}

func (m *mockRepo) GetWordSetsContainingWord(ctx context.Context, userID bson.ObjectID, wordIDs []bson.ObjectID) ([]wordset.WordSetItem, error) {
	return nil, nil
}

func (m *mockRepo) VerifyWordsExist(ctx context.Context, wordIDs []bson.ObjectID) (int, error) {
	return len(wordIDs), nil
}

func (m *mockRepo) CreateWordSetItem(ctx context.Context, item *wordset.WordSetItem) error {
	m.items = append(m.items, *item)
	return nil
}

func (m *mockRepo) GetWordSetItems(ctx context.Context, setID bson.ObjectID) ([]wordset.WordSetItem, error) {
	return m.items, nil
}

func (m *mockRepo) RemoveWordSetItem(ctx context.Context, setID, wordID bson.ObjectID) error {
	return nil
}

func (m *mockRepo) ReorderWordSetItems(ctx context.Context, setID bson.ObjectID, wordIDs []bson.ObjectID) error {
	return nil
}

func (m *mockRepo) EnsureIndexes(ctx context.Context) error {
	return nil
}

func (m *mockRepo) ExploreWordSets(ctx context.Context, req wordset.ExploreWordSetsRequest) ([]wordset.WordSet, int, error) {
	return nil, 0, nil
}

func (m *mockRepo) BatchRemoveWordSetItems(ctx context.Context, setID bson.ObjectID, wordIDs []bson.ObjectID) (int64, error) {
	return 0, nil
}

func TestCreateWordSet(t *testing.T) {
	repo := &mockRepo{}
	svc := wordset.NewService(repo, nil, nil, nil)
	userID := bson.NewObjectID()

	req := wordset.CreateWordSetRequest{
		Name: "",
	}
	_, err := svc.CreateWordSet(context.Background(), userID, req)
	if err == nil {
		t.Errorf("Expected error for empty name, got nil")
	}

	req = wordset.CreateWordSetRequest{
		Name: "IELTS Environment",
	}
	set, err := svc.CreateWordSet(context.Background(), userID, req)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if set.Visibility != wordset.WordSetVisibilityPrivate {
		t.Errorf("Expected visibility to be private, got %v", set.Visibility)
	}
}
