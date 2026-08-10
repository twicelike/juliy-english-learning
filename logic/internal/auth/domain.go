package auth

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// User represents a user document in MongoDB.
type User struct {
	ID           bson.ObjectID `bson:"_id,omitempty"`
	Email        string        `bson:"email"`
	PasswordHash string        `bson:"password_hash"`
	Name         string        `bson:"name"`
	CreatedAt    time.Time     `bson:"created_at"`
	UpdatedAt    time.Time     `bson:"updated_at"`
}

// Session represents a refresh token session document in MongoDB.
type Session struct {
	ID               bson.ObjectID `bson:"_id,omitempty"`
	UserID           bson.ObjectID `bson:"user_id"`
	RefreshTokenHash string        `bson:"refresh_token_hash"`
	ExpiresAt        time.Time     `bson:"expires_at"`
	CreatedAt        time.Time     `bson:"created_at"`
}

// -----------------------------------------------------------------------------
// DTOs
// -----------------------------------------------------------------------------

type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
	Name     string `json:"name" binding:"required"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type AuthResponse struct {
	AccessToken string       `json:"access_token"`
	User        UserResponse `json:"user"`
}

type UserResponse struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	Name  string `json:"name"`
}
