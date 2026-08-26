package match

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type SwipeAction string

const (
	SwipeKeep SwipeAction = "keep"
	SwipePass SwipeAction = "pass"
)

type UserSwipe struct {
	ID        bson.ObjectID `bson:"_id,omitempty"`
	UserID    bson.ObjectID `bson:"user_id"`
	WordID    bson.ObjectID `bson:"word_id"`
	Action    SwipeAction   `bson:"action"`
	CreatedAt time.Time     `bson:"created_at"`
}

type SwipeRequest struct {
	WordID string      `json:"word_id"`
	Action SwipeAction `json:"action"`
}

type MatchCardDTO struct {
	WordID      string `json:"word_id"`
	Word        string `json:"word"`
	POS         string `json:"pos"`
	SenseID     string `json:"sense_id"`
	Definition  string `json:"definition"`
	Translation string `json:"translation"`
	// Community signals
	TotalSwipes int `json:"total_swipes,omitempty"`
	MatchRate   int `json:"match_rate,omitempty"` // Percentage 0-100
}

type Repository interface {
	EnsureIndexes(ctx context.Context) error
	RecordSwipe(ctx context.Context, swipe *UserSwipe) error
	GetSwipedWordIDs(ctx context.Context, userID bson.ObjectID) ([]bson.ObjectID, error)
	GetCommunitySignal(ctx context.Context, wordID bson.ObjectID) (likes int, total int, err error)
}

type Service interface {
	GetRecommendations(ctx context.Context, userID string, limit int) ([]MatchCardDTO, error)
	Swipe(ctx context.Context, userID string, req *SwipeRequest) error
}
