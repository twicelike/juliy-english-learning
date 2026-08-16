package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/english/backend/internal/mail"
	"go.mongodb.org/mongo-driver/v2/bson"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUserExists            = errors.New("email already registered")
	ErrInvalidCreds          = errors.New("invalid email or password")
	ErrInvalidToken          = errors.New("invalid or expired session token")
	ErrEmailAlreadyExists    = errors.New("google_email_exists")
	ErrEmailNotVerified      = errors.New("google_email_not_verified")
	ErrIdentityAlreadyLinked = errors.New("identity_already_linked")
)

type Service interface {
	Register(ctx context.Context, req *RegisterRequest, userAgent, ip string) (*UserResponse, string, error)
	Login(ctx context.Context, req *LoginRequest, userAgent, ip string) (*UserResponse, string, error)
	Logout(ctx context.Context, rawToken string) error
	LogoutAll(ctx context.Context, userID bson.ObjectID) error
	GetActiveSessions(ctx context.Context, userID bson.ObjectID, currentRawToken string) ([]SessionResponse, error)
	RevokeSession(ctx context.Context, sessionID bson.ObjectID, userID bson.ObjectID) error
	GetMe(ctx context.Context, userID bson.ObjectID) (*UserResponse, error)
	GetUserIDFromToken(ctx context.Context, rawToken string) (bson.ObjectID, error)
	LoginWithGoogle(ctx context.Context, userInfo *GoogleUserInfo, userAgent, ip string) (*UserResponse, string, error)
	LinkGoogleIdentity(ctx context.Context, userID bson.ObjectID, userInfo *GoogleUserInfo) error

	// Verification & Recovery
	VerifyEmail(ctx context.Context, rawToken string) error
	ResendVerification(ctx context.Context, email string) error
	ForgotPassword(ctx context.Context, email string) error
	ResetPassword(ctx context.Context, rawToken, newPassword string) error
}

type service struct {
	repo           Repository
	mailer         mail.Mailer
	sessionExpDays int
	frontendURL    string
}

func NewService(repo Repository, mailer mail.Mailer, sessionExpDays int, frontendURL string) Service {
	return &service{
		repo:           repo,
		mailer:         mailer,
		sessionExpDays: sessionExpDays,
		frontendURL:    frontendURL,
	}
}

func (s *service) Register(ctx context.Context, req *RegisterRequest, userAgent, ip string) (*UserResponse, string, error) {
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
		Email:         req.Email,
		PasswordHash:  string(hashed),
		DisplayName:   req.DisplayName,
		EmailVerified: false,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	if err := s.repo.CreateUser(ctx, user); err != nil {
		return nil, "", err
	}

	log.Printf("[AUTH] event=auth.register.success user_id=%s ip=%s\n", user.ID.Hex(), ip)

	// Generate and send verification email
	_ = s.sendVerificationEmail(ctx, user)

	return s.createSession(ctx, user, userAgent, ip)
}

func (s *service) Login(ctx context.Context, req *LoginRequest, userAgent, ip string) (*UserResponse, string, error) {
	user, err := s.repo.GetUserByEmail(ctx, req.Email)
	if err != nil {
		return nil, "", err
	}
	if user == nil {
		log.Printf("[AUTH] event=auth.login.failed ip=%s reason=user_not_found\n", ip)
		return nil, "", ErrInvalidCreds
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		log.Printf("[AUTH] event=auth.login.failed ip=%s reason=invalid_password\n", ip)
		return nil, "", ErrInvalidCreds
	}

	log.Printf("[AUTH] event=auth.login.success user_id=%s ip=%s\n", user.ID.Hex(), ip)

	return s.createSession(ctx, user, userAgent, ip)
}

func (s *service) Logout(ctx context.Context, rawToken string) error {
	if rawToken == "" {
		return nil
	}
	hash := hashToken(rawToken)
	
	err := s.repo.RevokeSession(ctx, hash)
	if err == nil {
		log.Printf("[AUTH] event=auth.session.revoked hash=%s\n", hash)
	}
	return err
}

func (s *service) LogoutAll(ctx context.Context, userID bson.ObjectID) error {
	err := s.repo.RevokeAllSessions(ctx, userID)
	if err == nil {
		log.Printf("[AUTH] event=auth.session.revoke_all user_id=%s\n", userID.Hex())
	}
	return err
}

func (s *service) GetActiveSessions(ctx context.Context, userID bson.ObjectID, currentRawToken string) ([]SessionResponse, error) {
	sessions, err := s.repo.GetSessionsByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	currentHash := ""
	if currentRawToken != "" {
		currentHash = hashToken(currentRawToken)
	}

	var res []SessionResponse
	for _, sess := range sessions {
		res = append(res, SessionResponse{
			ID:        sess.ID.Hex(),
			UserAgent: sess.UserAgent,
			IPAddress: sess.IPAddress,
			CreatedAt: sess.CreatedAt,
			ExpiresAt: sess.ExpiresAt,
			IsCurrent: sess.TokenHash == currentHash,
		})
	}
	return res, nil
}

func (s *service) RevokeSession(ctx context.Context, sessionID bson.ObjectID, userID bson.ObjectID) error {
	err := s.repo.RevokeSessionByID(ctx, sessionID, userID)
	if err == nil {
		log.Printf("[AUTH] event=auth.session.revoked id=%s user_id=%s\n", sessionID.Hex(), userID.Hex())
	}
	return err
}

func (s *service) GetMe(ctx context.Context, userID bson.ObjectID) (*UserResponse, error) {
	user, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("user not found")
	}
	return s.mapUserResponse(ctx, user), nil
}

func (s *service) GetUserIDFromToken(ctx context.Context, rawToken string) (bson.ObjectID, error) {
	if rawToken == "" {
		return bson.NilObjectID, ErrInvalidToken
	}
	hash := hashToken(rawToken)
	session, err := s.repo.GetSessionByHash(ctx, hash)
	if err != nil {
		return bson.NilObjectID, err
	}
	if session == nil || session.RevokedAt != nil || session.ExpiresAt.Before(time.Now()) {
		return bson.NilObjectID, ErrInvalidToken
	}
	return session.UserID, nil
}

func (s *service) LoginWithGoogle(ctx context.Context, userInfo *GoogleUserInfo, userAgent, ip string) (*UserResponse, string, error) {
	if !userInfo.EmailVerified {
		return nil, "", ErrEmailNotVerified
	}

	identity, err := s.repo.GetAuthIdentity(ctx, "google", userInfo.Sub)
	if err != nil {
		return nil, "", err
	}

	if identity != nil {
		user, err := s.repo.GetUserByID(ctx, identity.UserID)
		if err != nil {
			return nil, "", err
		}
		if user == nil {
			return nil, "", errors.New("user not found for identity")
		}
		return s.createSession(ctx, user, userAgent, ip)
	}

	existingUser, err := s.repo.GetUserByEmail(ctx, userInfo.Email)
	if err != nil {
		return nil, "", err
	}

	if existingUser != nil {
		return nil, "", ErrEmailAlreadyExists
	}

	now := time.Now()
	user := &User{
		Email:         userInfo.Email,
		DisplayName:   userInfo.Name,
		AvatarURL:     userInfo.Picture,
		EmailVerified: true,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	if err := s.repo.CreateUser(ctx, user); err != nil {
		return nil, "", err
	}

	newIdentity := &AuthIdentity{
		UserID:         user.ID,
		Provider:       "google",
		ProviderUserID: userInfo.Sub,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	if err := s.repo.CreateAuthIdentity(ctx, newIdentity); err != nil {
		return nil, "", err
	}

	return s.createSession(ctx, user, userAgent, ip)
}

func (s *service) LinkGoogleIdentity(ctx context.Context, userID bson.ObjectID, userInfo *GoogleUserInfo) error {
	identity, err := s.repo.GetAuthIdentity(ctx, "google", userInfo.Sub)
	if err != nil {
		return err
	}
	if identity != nil {
		if identity.UserID != userID {
			return ErrIdentityAlreadyLinked
		}
		return nil
	}

	now := time.Now()
	newIdentity := &AuthIdentity{
		UserID:         userID,
		Provider:       "google",
		ProviderUserID: userInfo.Sub,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	return s.repo.CreateAuthIdentity(ctx, newIdentity)
}

// -----------------------------------------------------------------------------
// Verification & Recovery
// -----------------------------------------------------------------------------

func (s *service) VerifyEmail(ctx context.Context, rawToken string) error {
	hash := hashToken(rawToken)
	token, err := s.repo.GetEmailVerificationToken(ctx, hash)
	if err != nil {
		return err
	}
	if token == nil || token.ExpiresAt.Before(time.Now()) {
		return errors.New("invalid or expired token")
	}

	if err := s.repo.UpdateUserEmailVerified(ctx, token.UserID, true); err != nil {
		return err
	}

	log.Printf("[AUTH] event=auth.email.verified user_id=%s\n", token.UserID.Hex())

	_ = s.repo.DeleteEmailVerificationToken(ctx, hash)
	return nil
}

func (s *service) ResendVerification(ctx context.Context, email string) error {
	user, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil {
		return err
	}
	if user == nil || user.EmailVerified {
		return nil // Do nothing if user doesn't exist or already verified (prevents enumeration)
	}

	// Revoke old tokens
	_ = s.repo.DeleteAllEmailVerificationTokensForUser(ctx, user.ID)

	return s.sendVerificationEmail(ctx, user)
}

func (s *service) ForgotPassword(ctx context.Context, email string) error {
	user, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil {
		return err
	}
	if user == nil {
		return nil // Always return success
	}

	// Delete old tokens
	_ = s.repo.DeleteAllPasswordResetTokensForUser(ctx, user.ID)

	rawToken, err := generateRandomToken()
	if err != nil {
		return err
	}

	now := time.Now()
	resetToken := &PasswordResetToken{
		UserID:    user.ID,
		TokenHash: hashToken(rawToken),
		CreatedAt: now,
		ExpiresAt: now.Add(15 * time.Minute),
	}

	if err := s.repo.CreatePasswordResetToken(ctx, resetToken); err != nil {
		return err
	}

	resetLink := fmt.Sprintf("%s/reset-password?token=%s", s.frontendURL, rawToken)
	body := fmt.Sprintf("<p>Click the link below to reset your password:</p><p><a href='%s'>%s</a></p><p>This link expires in 15 minutes.</p>", resetLink, resetLink)
	
	go s.mailer.SendEmail([]string{user.Email}, "Password Reset", body)

	return nil
}

func (s *service) ResetPassword(ctx context.Context, rawToken, newPassword string) error {
	hash := hashToken(rawToken)
	token, err := s.repo.GetPasswordResetToken(ctx, hash)
	if err != nil {
		return err
	}
	if token == nil || token.ExpiresAt.Before(time.Now()) {
		return errors.New("invalid or expired token")
	}

	hashedPwd, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	if err := s.repo.UpdateUserPassword(ctx, token.UserID, string(hashedPwd)); err != nil {
		return err
	}

	log.Printf("[AUTH] event=auth.password.reset user_id=%s\n", token.UserID.Hex())

	_ = s.repo.DeletePasswordResetToken(ctx, hash)
	_ = s.repo.RevokeAllSessions(ctx, token.UserID)
	log.Printf("[AUTH] event=auth.session.revoke_all user_id=%s reason=password_reset\n", token.UserID.Hex())

	return nil
}

// -----------------------------------------------------------------------------
// Helper Methods
// -----------------------------------------------------------------------------

func (s *service) sendVerificationEmail(ctx context.Context, user *User) error {
	rawToken, err := generateRandomToken()
	if err != nil {
		return err
	}

	now := time.Now()
	token := &EmailVerificationToken{
		UserID:    user.ID,
		TokenHash: hashToken(rawToken),
		CreatedAt: now,
		ExpiresAt: now.Add(15 * time.Minute),
	}

	if err := s.repo.CreateEmailVerificationToken(ctx, token); err != nil {
		return err
	}

	verifyLink := fmt.Sprintf("%s/verify-email?token=%s", s.frontendURL, rawToken)
	body := fmt.Sprintf("<p>Please click the link below to verify your email:</p><p><a href='%s'>%s</a></p><p>This link expires in 15 minutes.</p>", verifyLink, verifyLink)
	
	go s.mailer.SendEmail([]string{user.Email}, "Verify your email", body)

	return nil
}

func (s *service) createSession(ctx context.Context, user *User, userAgent, ip string) (*UserResponse, string, error) {
	// Generate 32-byte secure random token
	rawToken, err := generateRandomToken()
	if err != nil {
		return nil, "", err
	}

	// Hash the token before storing
	tokenHash := hashToken(rawToken)
	expiresAt := time.Now().AddDate(0, 0, s.sessionExpDays)

	session := &Session{
		UserID:    user.ID,
		TokenHash: tokenHash,
		UserAgent: userAgent,
		IPAddress: ip,
		ExpiresAt: expiresAt,
		CreatedAt: time.Now(),
	}

	if err := s.repo.CreateSession(ctx, session); err != nil {
		return nil, "", err
	}

	log.Printf("[AUTH] event=auth.session.created user_id=%s ip=%s\n", user.ID.Hex(), ip)

	resp := s.mapUserResponse(ctx, user)

	return resp, rawToken, nil
}

func (s *service) mapUserResponse(ctx context.Context, user *User) *UserResponse {
	identities, _ := s.repo.GetAuthIdentitiesByUserID(ctx, user.ID)
	hasGoogle := false
	for _, id := range identities {
		if id.Provider == "google" {
			hasGoogle = true
			break
		}
	}

	return &UserResponse{
		ID:              user.ID.Hex(),
		Email:           user.Email,
		DisplayName:     user.DisplayName,
		AvatarURL:       user.AvatarURL,
		EmailVerified:   user.EmailVerified,
		HasGoogleLinked: hasGoogle,
	}
}

func generateRandomToken() (string, error) {
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
