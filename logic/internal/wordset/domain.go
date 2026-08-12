package wordset

import (
	"context"
	"time"

	"github.com/english/backend/internal/dictionary"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type WordSetVisibility string

const (
	WordSetVisibilityPrivate WordSetVisibility = "private"
	WordSetVisibilityPublic  WordSetVisibility = "public"
)

type WordSet struct {
	ID          bson.ObjectID     `bson:"_id,omitempty" json:"id"`
	UserID      bson.ObjectID     `bson:"user_id" json:"-"`
	Name        string            `bson:"name" json:"name"`
	Description string            `bson:"description" json:"description"`
	Visibility  WordSetVisibility `bson:"visibility" json:"visibility"`
	WordCount   int               `bson:"word_count" json:"word_count"`
	CreatedAt   time.Time         `bson:"created_at" json:"created_at"`
	UpdatedAt   time.Time         `bson:"updated_at" json:"updated_at"`
}

type WordSetItem struct {
	ID        bson.ObjectID `bson:"_id,omitempty" json:"id"`
	SetID     bson.ObjectID `bson:"set_id" json:"set_id"`
	WordID    bson.ObjectID `bson:"word_id" json:"word_id"`
	Position  int           `bson:"position" json:"position"`
	CreatedAt time.Time     `bson:"created_at" json:"created_at"`
}

// DTOs
type WordSetDetailResponse struct {
	WordSet `bson:",inline"`
	ItemIDs []string `json:"item_ids"` // string for frontend convenience
}

type WordSetListItem struct {
	WordSet      `bson:",inline"`
	ContainsWord bool `json:"contains_word"`
}

type WordSetListResponse struct {
	Items []WordSetListItem `json:"items"`
	Total int               `json:"total"`
}

type CreateWordSetRequest struct {
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Visibility  WordSetVisibility `json:"visibility"`
}

type UpdateWordSetRequest struct {
	Name        *string            `json:"name"`
	Description *string            `json:"description"`
	Visibility  *WordSetVisibility `json:"visibility"`
}

type WordViewerItemDTO struct {
	dictionary.WordDTO
	ProgressState string `json:"progress_state"`
}

type AddWordRequest struct {
	WordID string `json:"word_id"`
}

type ReorderWordsRequest struct {
	WordIDs []string `json:"word_ids"`
}

type BulkAddWordsRequest struct {
	WordIDs []string `json:"word_ids"`
}

type BulkAddWordsResponse struct {
	Added     int `json:"added"`
	WordCount int `json:"word_count"`
	Limit     int `json:"limit"`
}

type Repository interface {
	CreateWordSet(ctx context.Context, set *WordSet) error
	GetWordSetsByUserID(ctx context.Context, userID bson.ObjectID) ([]WordSet, error)
	GetWordSetByID(ctx context.Context, id bson.ObjectID) (*WordSet, error)
	UpdateWordSet(ctx context.Context, id bson.ObjectID, update bson.M) error
	DeleteWordSetWithItems(ctx context.Context, id bson.ObjectID) error

	CreateWordSetItem(ctx context.Context, item *WordSetItem) error
	CreateWordSetItems(ctx context.Context, items []WordSetItem) (int, error)
	GetWordSetItems(ctx context.Context, setID bson.ObjectID) ([]WordSetItem, error)
	GetWordSetsContainingWord(ctx context.Context, userID, wordID bson.ObjectID) ([]bson.ObjectID, error)
	VerifyWordsExist(ctx context.Context, wordIDs []bson.ObjectID) (int, error)
	RemoveWordSetItem(ctx context.Context, setID, wordID bson.ObjectID) error
	ReorderWordSetItems(ctx context.Context, setID bson.ObjectID, wordIDs []bson.ObjectID) error

	EnsureIndexes(ctx context.Context) error
}

type Service interface {
	CreateWordSet(ctx context.Context, userID bson.ObjectID, req CreateWordSetRequest) (*WordSet, error)
	GetWordSetsByUserID(ctx context.Context, userID bson.ObjectID, containsWordID *bson.ObjectID) (*WordSetListResponse, error)
	GetWordSetDetail(ctx context.Context, userID, setID bson.ObjectID) (*WordSetDetailResponse, error)
	UpdateWordSet(ctx context.Context, userID, setID bson.ObjectID, req UpdateWordSetRequest) (*WordSet, error)
	DeleteWordSet(ctx context.Context, userID, setID bson.ObjectID) error

	AddWord(ctx context.Context, userID, setID, wordID bson.ObjectID) error
	BulkAddWords(ctx context.Context, userID, setID bson.ObjectID, req BulkAddWordsRequest) (*BulkAddWordsResponse, error)
	RemoveWord(ctx context.Context, userID, setID, wordID bson.ObjectID) error
	ReorderWords(ctx context.Context, userID, setID bson.ObjectID, req ReorderWordsRequest) error

	GetWordSetWords(ctx context.Context, userID, setID bson.ObjectID) ([]WordViewerItemDTO, error)
}
