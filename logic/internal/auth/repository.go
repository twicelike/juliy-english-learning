package auth

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type Repository interface {
	// Users
	CreateUser(ctx context.Context, user *User) error
	GetUserByEmail(ctx context.Context, email string) (*User, error)
	GetUserByID(ctx context.Context, id bson.ObjectID) (*User, error)
	ListUsers(ctx context.Context, offset, limit int) ([]User, int64, error)
	UpdateUserRoleStatus(ctx context.Context, id bson.ObjectID, role, status string) error

	// Sessions
	CreateSession(ctx context.Context, session *Session) error
	GetSessionByHash(ctx context.Context, hash string) (*Session, error)
	GetSessionsByUserID(ctx context.Context, userID bson.ObjectID) ([]Session, error)
	RevokeSession(ctx context.Context, hash string) error
	RevokeSessionByID(ctx context.Context, sessionID bson.ObjectID, userID bson.ObjectID) error
	RevokeAllSessions(ctx context.Context, userID bson.ObjectID) error

	// Auth Identities
	CreateAuthIdentity(ctx context.Context, identity *AuthIdentity) error
	GetAuthIdentity(ctx context.Context, provider string, providerUserID string) (*AuthIdentity, error)
	GetAuthIdentitiesByUserID(ctx context.Context, userID bson.ObjectID) ([]AuthIdentity, error)

	// Verification Tokens
	CreateEmailVerificationToken(ctx context.Context, token *EmailVerificationToken) error
	GetEmailVerificationToken(ctx context.Context, hash string) (*EmailVerificationToken, error)
	DeleteEmailVerificationToken(ctx context.Context, hash string) error
	DeleteAllEmailVerificationTokensForUser(ctx context.Context, userID bson.ObjectID) error

	// Password Reset Tokens
	CreatePasswordResetToken(ctx context.Context, token *PasswordResetToken) error
	GetPasswordResetToken(ctx context.Context, hash string) (*PasswordResetToken, error)
	DeletePasswordResetToken(ctx context.Context, hash string) error
	DeleteAllPasswordResetTokensForUser(ctx context.Context, userID bson.ObjectID) error

	// User Updates
	UpdateUserEmailVerified(ctx context.Context, userID bson.ObjectID, verified bool) error
	UpdateUserPassword(ctx context.Context, userID bson.ObjectID, passwordHash string) error

	EnsureIndexes(ctx context.Context) error
}

type repository struct {
	userColl          *mongo.Collection
	sessionColl       *mongo.Collection
	identityColl      *mongo.Collection
	emailVerifColl    *mongo.Collection
	passwordResetColl *mongo.Collection
}

func NewRepository(db *mongo.Database) Repository {
	return &repository{
		userColl:          db.Collection("users"),
		sessionColl:       db.Collection("sessions"),
		identityColl:      db.Collection("auth_identities"),
		emailVerifColl:    db.Collection("email_verification_tokens"),
		passwordResetColl: db.Collection("password_reset_tokens"),
	}
}

// -----------------------------------------------------------------------------
// Users
// -----------------------------------------------------------------------------

func (r *repository) CreateUser(ctx context.Context, user *User) error {
	res, err := r.userColl.InsertOne(ctx, user)
	if err != nil {
		return err
	}
	if id, ok := res.InsertedID.(bson.ObjectID); ok {
		user.ID = id
	}
	return nil
}

func (r *repository) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	var user User
	err := r.userColl.FindOne(ctx, bson.M{"email": email}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func (r *repository) GetUserByID(ctx context.Context, id bson.ObjectID) (*User, error) {
	var user User
	err := r.userColl.FindOne(ctx, bson.M{"_id": id}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func (r *repository) ListUsers(ctx context.Context, offset, limit int) ([]User, int64, error) {
	filter := bson.M{}
	
	total, err := r.userColl.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}}).SetSkip(int64(offset)).SetLimit(int64(limit))
	cursor, err := r.userColl.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var users []User
	if err = cursor.All(ctx, &users); err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

func (r *repository) UpdateUserRoleStatus(ctx context.Context, id bson.ObjectID, role, status string) error {
	update := bson.M{"$set": bson.M{
		"role": role,
		"status": status,
		"updated_at": time.Now(),
	}}
	_, err := r.userColl.UpdateOne(ctx, bson.M{"_id": id}, update)
	return err
}

func (r *repository) UpdateUserEmailVerified(ctx context.Context, userID bson.ObjectID, verified bool) error {
	now := time.Now()
	_, err := r.userColl.UpdateOne(
		ctx,
		bson.M{"_id": userID},
		bson.M{"$set": bson.M{"email_verified": verified, "updated_at": now}},
	)
	return err
}

func (r *repository) UpdateUserPassword(ctx context.Context, userID bson.ObjectID, passwordHash string) error {
	now := time.Now()
	_, err := r.userColl.UpdateOne(
		ctx,
		bson.M{"_id": userID},
		bson.M{"$set": bson.M{"password_hash": passwordHash, "updated_at": now}},
	)
	return err
}

// -----------------------------------------------------------------------------
// Sessions
// -----------------------------------------------------------------------------

func (r *repository) CreateSession(ctx context.Context, session *Session) error {
	res, err := r.sessionColl.InsertOne(ctx, session)
	if err != nil {
		return err
	}
	if id, ok := res.InsertedID.(bson.ObjectID); ok {
		session.ID = id
	}
	return nil
}

func (r *repository) GetSessionByHash(ctx context.Context, hash string) (*Session, error) {
	var session Session
	err := r.sessionColl.FindOne(ctx, bson.M{"token_hash": hash}).Decode(&session)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &session, nil
}

func (r *repository) GetSessionsByUserID(ctx context.Context, userID bson.ObjectID) ([]Session, error) {
	var sessions []Session
	cursor, err := r.sessionColl.Find(ctx, bson.M{
		"user_id":    userID,
		"revoked_at": nil,
		"expires_at": bson.M{"$gt": time.Now()},
	})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	if err := cursor.All(ctx, &sessions); err != nil {
		return nil, err
	}
	return sessions, nil
}

func (r *repository) RevokeSession(ctx context.Context, hash string) error {
	now := time.Now()
	_, err := r.sessionColl.UpdateOne(
		ctx,
		bson.M{"token_hash": hash},
		bson.M{"$set": bson.M{"revoked_at": now}},
	)
	return err
}

func (r *repository) RevokeSessionByID(ctx context.Context, sessionID bson.ObjectID, userID bson.ObjectID) error {
	now := time.Now()
	_, err := r.sessionColl.UpdateOne(
		ctx,
		bson.M{"_id": sessionID, "user_id": userID},
		bson.M{"$set": bson.M{"revoked_at": now}},
	)
	return err
}

func (r *repository) RevokeAllSessions(ctx context.Context, userID bson.ObjectID) error {
	now := time.Now()
	_, err := r.sessionColl.UpdateMany(
		ctx,
		bson.M{"user_id": userID, "revoked_at": nil},
		bson.M{"$set": bson.M{"revoked_at": now}},
	)
	return err
}

// -----------------------------------------------------------------------------
// Auth Identities
// -----------------------------------------------------------------------------

func (r *repository) CreateAuthIdentity(ctx context.Context, identity *AuthIdentity) error {
	res, err := r.identityColl.InsertOne(ctx, identity)
	if err != nil {
		return err
	}
	if id, ok := res.InsertedID.(bson.ObjectID); ok {
		identity.ID = id
	}
	return nil
}

func (r *repository) GetAuthIdentity(ctx context.Context, provider string, providerUserID string) (*AuthIdentity, error) {
	var identity AuthIdentity
	err := r.identityColl.FindOne(ctx, bson.M{
		"provider":         provider,
		"provider_user_id": providerUserID,
	}).Decode(&identity)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &identity, nil
}

func (r *repository) GetAuthIdentitiesByUserID(ctx context.Context, userID bson.ObjectID) ([]AuthIdentity, error) {
	var identities []AuthIdentity
	cursor, err := r.identityColl.Find(ctx, bson.M{"user_id": userID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	if err := cursor.All(ctx, &identities); err != nil {
		return nil, err
	}
	return identities, nil
}

// -----------------------------------------------------------------------------
// Verification & Reset Tokens
// -----------------------------------------------------------------------------

func (r *repository) CreateEmailVerificationToken(ctx context.Context, token *EmailVerificationToken) error {
	res, err := r.emailVerifColl.InsertOne(ctx, token)
	if err != nil {
		return err
	}
	if id, ok := res.InsertedID.(bson.ObjectID); ok {
		token.ID = id
	}
	return nil
}

func (r *repository) GetEmailVerificationToken(ctx context.Context, hash string) (*EmailVerificationToken, error) {
	var token EmailVerificationToken
	err := r.emailVerifColl.FindOne(ctx, bson.M{"token_hash": hash}).Decode(&token)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &token, nil
}

func (r *repository) DeleteEmailVerificationToken(ctx context.Context, hash string) error {
	_, err := r.emailVerifColl.DeleteOne(ctx, bson.M{"token_hash": hash})
	return err
}

func (r *repository) DeleteAllEmailVerificationTokensForUser(ctx context.Context, userID bson.ObjectID) error {
	_, err := r.emailVerifColl.DeleteMany(ctx, bson.M{"user_id": userID})
	return err
}

func (r *repository) CreatePasswordResetToken(ctx context.Context, token *PasswordResetToken) error {
	res, err := r.passwordResetColl.InsertOne(ctx, token)
	if err != nil {
		return err
	}
	if id, ok := res.InsertedID.(bson.ObjectID); ok {
		token.ID = id
	}
	return nil
}

func (r *repository) GetPasswordResetToken(ctx context.Context, hash string) (*PasswordResetToken, error) {
	var token PasswordResetToken
	err := r.passwordResetColl.FindOne(ctx, bson.M{"token_hash": hash}).Decode(&token)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &token, nil
}

func (r *repository) DeletePasswordResetToken(ctx context.Context, hash string) error {
	_, err := r.passwordResetColl.DeleteOne(ctx, bson.M{"token_hash": hash})
	return err
}

func (r *repository) DeleteAllPasswordResetTokensForUser(ctx context.Context, userID bson.ObjectID) error {
	_, err := r.passwordResetColl.DeleteMany(ctx, bson.M{"user_id": userID})
	return err
}

// -----------------------------------------------------------------------------
// Indexes
// -----------------------------------------------------------------------------

func (r *repository) EnsureIndexes(ctx context.Context) error {
	// Drop legacy indexes from previous session schema (ignore errors if they don't exist)
	_ = r.sessionColl.Indexes().DropOne(ctx, "refresh_token_hash_1")
	_ = r.sessionColl.Indexes().DropOne(ctx, "access_token_hash_1")

	// Users collection: email UNIQUE
	if _, err := r.userColl.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "email", Value: 1}},
		Options: options.Index().SetUnique(true),
	}); err != nil {
		return err
	}

	// Sessions collection: token_hash UNIQUE
	if _, err := r.sessionColl.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "token_hash", Value: 1}},
		Options: options.Index().SetUnique(true),
	}); err != nil {
		return err
	}

	// Sessions collection: user_id
	if _, err := r.sessionColl.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "user_id", Value: 1}},
	}); err != nil {
		return err
	}

	// Sessions collection: expires_at TTL
	if _, err := r.sessionColl.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "expires_at", Value: 1}},
		Options: options.Index().SetExpireAfterSeconds(0),
	}); err != nil {
		return err
	}

	// Auth Identities: provider + provider_user_id UNIQUE
	if _, err := r.identityColl.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{
			{Key: "provider", Value: 1},
			{Key: "provider_user_id", Value: 1},
		},
		Options: options.Index().SetUnique(true),
	}); err != nil {
		return err
	}

	// Auth Identities: user_id
	if _, err := r.identityColl.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "user_id", Value: 1}},
	}); err != nil {
		return err
	}

	// Email Verification Tokens: token_hash UNIQUE
	if _, err := r.emailVerifColl.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "token_hash", Value: 1}},
		Options: options.Index().SetUnique(true),
	}); err != nil {
		return err
	}

	// Email Verification Tokens: expires_at TTL
	if _, err := r.emailVerifColl.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "expires_at", Value: 1}},
		Options: options.Index().SetExpireAfterSeconds(0),
	}); err != nil {
		return err
	}

	// Password Reset Tokens: token_hash UNIQUE
	if _, err := r.passwordResetColl.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "token_hash", Value: 1}},
		Options: options.Index().SetUnique(true),
	}); err != nil {
		return err
	}

	// Password Reset Tokens: expires_at TTL
	if _, err := r.passwordResetColl.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "expires_at", Value: 1}},
		Options: options.Index().SetExpireAfterSeconds(0),
	}); err != nil {
		return err
	}

	return nil
}
