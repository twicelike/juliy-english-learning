package user

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// SavedWord represents a bookmarked word for a user.
type SavedWord struct {
	ID      bson.ObjectID `bson:"_id,omitempty"`
	UserID  bson.ObjectID `bson:"user_id"`
	Word    string        `bson:"word"`
	Type    string        `bson:"type"` // "word" or "expression"
	AddedAt time.Time     `bson:"added_at"`
}

// -----------------------------------------------------------------------------
// DTOs
// -----------------------------------------------------------------------------

type BookmarkRequest struct {
	Word string `json:"word" binding:"required"`
	Type string `json:"type"` // optional, default "word"
}

type BookmarkResponse struct {
	Word    string    `json:"word"`
	Type    string    `json:"type"`
	AddedAt time.Time `json:"added_at"`
}
