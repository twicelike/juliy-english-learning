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

type WordSetDifficulty string

const (
	DifficultyBeginner     WordSetDifficulty = "Beginner"
	DifficultyIntermediate WordSetDifficulty = "Intermediate"
	DifficultyAdvanced     WordSetDifficulty = "Advanced"
)

var PredefinedTopics = []string{
	"Academic",
	"Business",
	"Conversation",
	"Daily Life",
	"Education",
	"IELTS",
	"Technology",
	"TOEIC",
	"Travel",
	"Work",
}

type WordSet struct {
	ID          bson.ObjectID     `bson:"_id,omitempty" json:"id"`
	UserID      bson.ObjectID     `bson:"user_id" json:"-"`
	Name        string            `bson:"name" json:"name"`
	Description string            `bson:"description" json:"description"`
	Visibility  WordSetVisibility `bson:"visibility" json:"visibility"`
	Topics      []string          `bson:"topics" json:"topics"`
	Difficulty  WordSetDifficulty `bson:"difficulty" json:"difficulty"`
	WordCount   int               `bson:"word_count" json:"word_count"`
	CloneCount  int               `bson:"clone_count" json:"clone_count"`
	CreatedAt   time.Time         `bson:"created_at" json:"created_at"`
	UpdatedAt   time.Time         `bson:"updated_at" json:"updated_at"`
}

type WordSetItem struct {
	ID         bson.ObjectID `bson:"_id,omitempty" json:"id"`
	SetID      bson.ObjectID `bson:"set_id" json:"set_id"`
	WordID     bson.ObjectID `bson:"word_id" json:"word_id"`
	SenseIndex int           `bson:"sense_index" json:"sense_index"`
	Position   int           `bson:"position" json:"position"`
	CreatedAt  time.Time     `bson:"created_at" json:"created_at"`
}

// DTOs
type WordSetDetailResponse struct {
	WordSet     `bson:",inline"`
	ItemIDs     []string        `json:"item_ids"` // string for frontend convenience
	SavedSenses []SavedSenseRef `json:"saved_senses"`
	IsOwner     bool            `json:"is_owner"`
	CreatorName string          `json:"creator_name,omitempty"`
}

type WordSetListItem struct {
	WordSet      `bson:",inline"`
	ContainsWord bool            `json:"contains_word"`
	SavedSenses  []SavedSenseRef `json:"saved_senses,omitempty"`
	CreatorName  string          `json:"creator_name,omitempty"`
}

type SavedSenseRef struct {
	WordID     string `json:"word_id"`
	SenseIndex int    `json:"sense_index"`
}

type WordSetListResponse struct {
	Items []WordSetListItem `json:"items"`
	Total int               `json:"total"`
}

type CreateWordSetRequest struct {
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Visibility  WordSetVisibility `json:"visibility"`
	Topics      []string          `json:"topics"`
	Difficulty  WordSetDifficulty `json:"difficulty"`
}

type UpdateWordSetRequest struct {
	Name        *string            `json:"name"`
	Description *string            `json:"description"`
	Visibility  *WordSetVisibility `json:"visibility"`
	Topics      *[]string          `json:"topics"`
	Difficulty  *WordSetDifficulty `json:"difficulty"`
}

type WordViewerItemDTO struct {
	dictionary.WordDTO
	SenseIndex    int    `json:"sense_index"`
	ProgressState string `json:"progress_state"`
}

type AddWordRequest struct {
	WordID     string `json:"word_id"`
	SenseIndex int    `json:"sense_index"`
}

type ReorderWordsRequest struct {
	WordIDs []string `json:"word_ids"`
}

type BulkAddWordsRequest struct {
	Items []AddWordRequest `json:"items"`
}

type BulkAddWordsResponse struct {
	Added     int `json:"added"`
	WordCount int `json:"word_count"`
	Limit     int `json:"limit"`
}

type BatchDeleteWordsRequest struct {
	WordIDs []string `json:"word_ids"`
}

type BatchDeleteWordsResponse struct {
	DeletedWordIDs []string `json:"deleted_word_ids"`
	WordCount      int      `json:"word_count"`
}

type RestoreWordsRequest struct {
	WordIDs []string `json:"word_ids"`
}

type RestoreWordsResponse struct {
	RestoredWordIDs []string `json:"restored_word_ids"`
	WordCount       int      `json:"word_count"`
}

type ExploreWordSetsRequest struct {
	Query      string `form:"q"`
	Topic      string `form:"topic"`
	Difficulty string `form:"difficulty"`
	Sort       string `form:"sort"`
	Page       int    `form:"page"`
	Limit      int    `form:"limit"`
}

type ExploreWordSetsResponse struct {
	Items   []WordSetListItem `json:"items"`
	Page    int               `json:"page"`
	Limit   int               `json:"limit"`
	Total   int               `json:"total"`
	HasMore bool              `json:"has_more"`
}

type Repository interface {
	CreateWordSet(ctx context.Context, set *WordSet) error
	GetWordSetsByUserID(ctx context.Context, userID bson.ObjectID) ([]WordSet, error)
	ExploreWordSets(ctx context.Context, req ExploreWordSetsRequest) ([]WordSet, int, error)
	GetWordSetByID(ctx context.Context, id bson.ObjectID) (*WordSet, error)
	UpdateWordSet(ctx context.Context, id bson.ObjectID, update bson.M) error
	DeleteWordSetWithItems(ctx context.Context, id bson.ObjectID) error

	CreateWordSetItem(ctx context.Context, item *WordSetItem) error
	CreateWordSetItems(ctx context.Context, items []WordSetItem) (int, error)
	GetWordSetItems(ctx context.Context, setID bson.ObjectID) ([]WordSetItem, error)
	GetWordSetsContainingWord(ctx context.Context, userID bson.ObjectID, wordIDs []bson.ObjectID) ([]WordSetItem, error)
	VerifyWordsExist(ctx context.Context, wordIDs []bson.ObjectID) (int, error)
	RemoveWordSetItem(ctx context.Context, setID, wordID bson.ObjectID) error
	BatchRemoveWordSetItems(ctx context.Context, setID bson.ObjectID, wordIDs []bson.ObjectID) (int64, error)
	ReorderWordSetItems(ctx context.Context, setID bson.ObjectID, wordIDs []bson.ObjectID) error

	EnsureIndexes(ctx context.Context) error
}

type Service interface {
	CreateWordSet(ctx context.Context, userID bson.ObjectID, req CreateWordSetRequest) (*WordSet, error)
	GetWordSetsByUserID(ctx context.Context, userID bson.ObjectID, containsWordIDs []bson.ObjectID) (*WordSetListResponse, error)
	GetWordSetDetail(ctx context.Context, userID, setID bson.ObjectID) (*WordSetDetailResponse, error)
	ExploreWordSets(ctx context.Context, req ExploreWordSetsRequest) (*ExploreWordSetsResponse, error)
	UpdateWordSet(ctx context.Context, userID, setID bson.ObjectID, req UpdateWordSetRequest) (*WordSet, error)
	DeleteWordSet(ctx context.Context, userID, setID bson.ObjectID) error
	CloneWordSet(ctx context.Context, userID, setID bson.ObjectID) (*WordSet, error)

	AddWord(ctx context.Context, userID, setID, wordID bson.ObjectID, senseIndex int) error
	BulkAddWords(ctx context.Context, userID, setID bson.ObjectID, req BulkAddWordsRequest) (*BulkAddWordsResponse, error)
	RemoveWord(ctx context.Context, userID, setID, wordID bson.ObjectID) error
	BatchDeleteWords(ctx context.Context, userID, setID bson.ObjectID, req BatchDeleteWordsRequest) (*BatchDeleteWordsResponse, error)
	RestoreWords(ctx context.Context, userID, setID bson.ObjectID, req RestoreWordsRequest) (*RestoreWordsResponse, error)
	ReorderWords(ctx context.Context, userID, setID bson.ObjectID, req ReorderWordsRequest) error

	GetWordSetWords(ctx context.Context, userID, setID bson.ObjectID) ([]WordViewerItemDTO, error)
}
