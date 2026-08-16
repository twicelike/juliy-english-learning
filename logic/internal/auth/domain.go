package auth

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// User represents a user document in MongoDB.
type User struct {
	ID            bson.ObjectID `bson:"_id,omitempty" json:"id"`
	Email         string        `bson:"email" json:"email"`
	PasswordHash  string        `bson:"password_hash" json:"-"`
	DisplayName   string        `bson:"display_name" json:"display_name"`
	AvatarURL     string        `bson:"avatar_url" json:"avatar_url"`
	EmailVerified bool          `bson:"email_verified" json:"email_verified"`
	Status        string        `bson:"status" json:"status"`
	CreatedAt     time.Time     `bson:"created_at" json:"created_at"`
	UpdatedAt     time.Time     `bson:"updated_at" json:"updated_at"`
	LastLoginAt   *time.Time    `bson:"last_login_at,omitempty" json:"last_login_at,omitempty"`
}

// Session represents an active login session.
type Session struct {
	ID         bson.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID     bson.ObjectID `bson:"user_id" json:"user_id"`
	TokenHash  string        `bson:"token_hash" json:"-"`
	UserAgent  string        `bson:"user_agent" json:"user_agent"`
	IPAddress  string        `bson:"ip_address" json:"ip_address"`
	ExpiresAt  time.Time     `bson:"expires_at" json:"expires_at"`
	CreatedAt  time.Time     `bson:"created_at" json:"created_at"`
	LastUsedAt *time.Time    `bson:"last_used_at,omitempty" json:"last_used_at,omitempty"`
	RevokedAt  *time.Time    `bson:"revoked_at,omitempty" json:"revoked_at,omitempty"`
}

// AuthIdentity represents a linked OAuth provider identity.
type AuthIdentity struct {
	ID             bson.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID         bson.ObjectID `bson:"user_id" json:"user_id"`
	Provider       string        `bson:"provider" json:"provider"` // e.g., "google"
	ProviderUserID string        `bson:"provider_user_id" json:"provider_user_id"`
	CreatedAt      time.Time     `bson:"created_at" json:"created_at"`
	UpdatedAt      time.Time     `bson:"updated_at" json:"updated_at"`
}

// EmailVerificationToken represents an email verification token.
type EmailVerificationToken struct {
	ID        bson.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID    bson.ObjectID `bson:"user_id" json:"user_id"`
	TokenHash string        `bson:"token_hash" json:"-"`
	CreatedAt time.Time     `bson:"created_at" json:"created_at"`
	ExpiresAt time.Time     `bson:"expires_at" json:"expires_at"`
}

// PasswordResetToken represents a password reset token.
type PasswordResetToken struct {
	ID        bson.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID    bson.ObjectID `bson:"user_id" json:"user_id"`
	TokenHash string        `bson:"token_hash" json:"-"`
	CreatedAt time.Time     `bson:"created_at" json:"created_at"`
	ExpiresAt time.Time     `bson:"expires_at" json:"expires_at"`
}

// -----------------------------------------------------------------------------
// DTOs
// -----------------------------------------------------------------------------

type RegisterRequest struct {
	Email       string `json:"email" binding:"required,email"`
	Password    string `json:"password" binding:"required,min=8"`
	DisplayName string `json:"display_name" binding:"required"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type UserResponse struct {
	ID              string `json:"id"`
	Email           string `json:"email"`
	DisplayName     string `json:"display_name"`
	AvatarURL       string `json:"avatar_url,omitempty"`
	EmailVerified   bool   `json:"email_verified"`
	HasGoogleLinked bool   `json:"has_google_linked"`
}

type VerifyEmailRequest struct {
	Token string `json:"token" binding:"required"`
}

type ResendVerificationRequest struct {
	Email string `json:"email" binding:"required,email"`
}

type ForgotPasswordRequest struct {
	Email string `json:"email" binding:"required,email"`
}

type ResetPasswordRequest struct {
	Token       string `json:"token" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=8"`
}

type SessionResponse struct {
	ID        string     `json:"id"`
	UserAgent string     `json:"user_agent"`
	IPAddress string     `json:"ip_address"`
	CreatedAt time.Time  `json:"created_at"`
	ExpiresAt time.Time  `json:"expires_at"`
	IsCurrent bool       `json:"is_current"`
}
