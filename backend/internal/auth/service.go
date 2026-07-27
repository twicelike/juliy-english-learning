package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUserExists    = errors.New("email already registered")
	ErrInvalidCreds  = errors.New("invalid email or password")
	ErrInvalidToken  = errors.New("invalid or expired refresh token")
)

type Service interface {
	Register(ctx context.Context, req *RegisterRequest) (*AuthResponse, string, error)
	Login(ctx context.Context, req *LoginRequest) (*AuthResponse, string, error)
	Refresh(ctx context.Context, refreshToken string) (*AuthResponse, string, error)
	Logout(ctx context.Context, refreshToken string) error
}

type service struct {
	repo           Repository
	jwtSecret      []byte
	jwtExpMins     int
	refreshExpDays int
}

func NewService(repo Repository, jwtSecret string, jwtExpMins, refreshExpDays int) Service {
	return &service{
		repo:           repo,
		jwtSecret:      []byte(jwtSecret),
		jwtExpMins:     jwtExpMins,
		refreshExpDays: refreshExpDays,
	}
}

func (s *service) Register(ctx context.Context, req *RegisterRequest) (*AuthResponse, string, error) {
	// Check if user exists
	existing, err := s.repo.GetUserByEmail(ctx, req.Email)
	if err != nil {
		return nil, "", err
	}
	if existing != nil {
		return nil, "", ErrUserExists
	}

	// Hash password
	hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, "", err
	}

	now := time.Now()
	user := &User{
		Email:        req.Email,
		PasswordHash: string(hashed),
		Name:         req.Name,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if err := s.repo.CreateUser(ctx, user); err != nil {
		return nil, "", err
	}

	return s.createAuthSession(ctx, user)
}

func (s *service) Login(ctx context.Context, req *LoginRequest) (*AuthResponse, string, error) {
	user, err := s.repo.GetUserByEmail(ctx, req.Email)
	if err != nil {
		return nil, "", err
	}
	if user == nil {
		return nil, "", ErrInvalidCreds
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, "", ErrInvalidCreds
	}

	return s.createAuthSession(ctx, user)
}

func (s *service) Refresh(ctx context.Context, refreshToken string) (*AuthResponse, string, error) {
	hash := hashToken(refreshToken)
	
	session, err := s.repo.GetSessionByHash(ctx, hash)
	if err != nil {
		return nil, "", err
	}
	if session == nil || session.ExpiresAt.Before(time.Now()) {
		return nil, "", ErrInvalidToken
	}

	user, err := s.repo.GetUserByID(ctx, session.UserID)
	if err != nil {
		return nil, "", err
	}
	if user == nil {
		return nil, "", ErrInvalidToken
	}

	// Delete old session
	_ = s.repo.DeleteSession(ctx, hash)

	// Create new session
	return s.createAuthSession(ctx, user)
}

func (s *service) Logout(ctx context.Context, refreshToken string) error {
	if refreshToken == "" {
		return nil
	}
	hash := hashToken(refreshToken)
	return s.repo.DeleteSession(ctx, hash)
}

// -----------------------------------------------------------------------------
// Helper Methods
// -----------------------------------------------------------------------------

func (s *service) createAuthSession(ctx context.Context, user *User) (*AuthResponse, string, error) {
	// Generate JWT Access Token
	accessToken, err := s.generateAccessToken(user.ID.Hex())
	if err != nil {
		return nil, "", err
	}

	// Generate Refresh Token
	refreshToken, err := generateOpaqueToken()
	if err != nil {
		return nil, "", err
	}

	// Hash the refresh token before storing
	rtHash := hashToken(refreshToken)
	expiresAt := time.Now().AddDate(0, 0, s.refreshExpDays)

	session := &Session{
		UserID:           user.ID,
		RefreshTokenHash: rtHash,
		ExpiresAt:        expiresAt,
		CreatedAt:        time.Now(),
	}

	if err := s.repo.CreateSession(ctx, session); err != nil {
		return nil, "", err
	}

	resp := &AuthResponse{
		AccessToken: accessToken,
		User: UserResponse{
			ID:    user.ID.Hex(),
			Email: user.Email,
			Name:  user.Name,
		},
	}

	return resp, refreshToken, nil
}

func (s *service) generateAccessToken(userID string) (string, error) {
	claims := jwt.MapClaims{
		"sub": userID,
		"exp": time.Now().Add(time.Duration(s.jwtExpMins) * time.Minute).Unix(),
		"iat": time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.jwtSecret)
}

func generateOpaqueToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func hashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}
