package user

import (
	"context"
	"errors"
	"time"

	"github.com/english/backend/internal/auth"
	"go.mongodb.org/mongo-driver/v2/bson"
)

var (
	ErrUserNotFound = errors.New("user not found")
)

type Service interface {
	GetProfile(ctx context.Context, userID string) (*auth.UserResponse, error)
	AddBookmark(ctx context.Context, userID string, word string, bType string) error
	RemoveBookmark(ctx context.Context, userID string, word string) error
	ListBookmarks(ctx context.Context, userID string) ([]BookmarkResponse, error)
}

type service struct {
	userRepo auth.Repository
	repo     Repository
}

func NewService(userRepo auth.Repository, repo Repository) Service {
	return &service{
		userRepo: userRepo,
		repo:     repo,
	}
}

func (s *service) GetProfile(ctx context.Context, userID string) (*auth.UserResponse, error) {
	oid, err := bson.ObjectIDFromHex(userID)
	if err != nil {
		return nil, err
	}

	u, err := s.userRepo.GetUserByID(ctx, oid)
	if err != nil {
		return nil, err
	}
	if u == nil {
		return nil, ErrUserNotFound
	}

	return &auth.UserResponse{
		ID:            u.ID.Hex(),
		Email:         u.Email,
		DisplayName:   u.DisplayName,
		AvatarURL:     u.AvatarURL,
		EmailVerified: u.EmailVerified,
	}, nil
}

func (s *service) AddBookmark(ctx context.Context, userID string, word string, bType string) error {
	oid, err := bson.ObjectIDFromHex(userID)
	if err != nil {
		return err
	}

	if bType == "" {
		bType = "word" // Default type
	}

	bookmark := &SavedWord{
		UserID:  oid,
		Word:    word,
		Type:    bType,
		AddedAt: time.Now(),
	}

	return s.repo.AddBookmark(ctx, bookmark)
}

func (s *service) RemoveBookmark(ctx context.Context, userID string, word string) error {
	oid, err := bson.ObjectIDFromHex(userID)
	if err != nil {
		return err
	}

	return s.repo.RemoveBookmark(ctx, oid, word)
}

func (s *service) ListBookmarks(ctx context.Context, userID string) ([]BookmarkResponse, error) {
	oid, err := bson.ObjectIDFromHex(userID)
	if err != nil {
		return nil, err
	}

	words, err := s.repo.ListBookmarks(ctx, oid)
	if err != nil {
		return nil, err
	}

	var dtos []BookmarkResponse
	for _, w := range words {
		dtos = append(dtos, BookmarkResponse{
			Word:    w.Word,
			Type:    w.Type,
			AddedAt: w.AddedAt,
		})
	}

	if dtos == nil {
		dtos = []BookmarkResponse{}
	}

	return dtos, nil
}
