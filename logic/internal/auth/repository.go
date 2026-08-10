package auth

import (
	"context"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type Repository interface {
	CreateUser(ctx context.Context, user *User) error
	GetUserByEmail(ctx context.Context, email string) (*User, error)
	GetUserByID(ctx context.Context, id bson.ObjectID) (*User, error)
	
	CreateSession(ctx context.Context, session *Session) error
	GetSessionByHash(ctx context.Context, hash string) (*Session, error)
	DeleteSession(ctx context.Context, hash string) error
	
	EnsureIndexes(ctx context.Context) error
}

type repository struct {
	userColl    *mongo.Collection
	sessionColl *mongo.Collection
}

func NewRepository(db *mongo.Database) Repository {
	return &repository{
		userColl:    db.Collection("users"),
		sessionColl: db.Collection("sessions"),
	}
}

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
	err := r.sessionColl.FindOne(ctx, bson.M{"refresh_token_hash": hash}).Decode(&session)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &session, nil
}

func (r *repository) DeleteSession(ctx context.Context, hash string) error {
	_, err := r.sessionColl.DeleteOne(ctx, bson.M{"refresh_token_hash": hash})
	return err
}

func (r *repository) EnsureIndexes(ctx context.Context) error {
	// Users collection: email UNIQUE
	_, err := r.userColl.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "email", Value: 1}},
		Options: options.Index().SetUnique(true),
	})
	if err != nil {
		return err
	}

	// Sessions collection: refresh_token_hash UNIQUE
	_, err = r.sessionColl.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "refresh_token_hash", Value: 1}},
		Options: options.Index().SetUnique(true),
	})
	if err != nil {
		return err
	}

	// Sessions collection: user_id
	_, err = r.sessionColl.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "user_id", Value: 1}},
	})
	if err != nil {
		return err
	}

	// Sessions collection: expires_at TTL
	_, err = r.sessionColl.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "expires_at", Value: 1}},
		Options: options.Index().SetExpireAfterSeconds(0),
	})
	return err
}
