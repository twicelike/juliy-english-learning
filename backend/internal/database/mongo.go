package database

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// Connect establishes a connection to MongoDB and verifies it with a Ping.
// Returns the mongo.Client so the caller can select the target database.
func Connect(uri string) (*mongo.Client, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(options.Client().ApplyURI(uri))
	if err != nil {
		return nil, fmt.Errorf("mongo.Connect: %w", err)
	}

	if err := Ping(ctx, client); err != nil {
		return nil, err
	}

	return client, nil
}

// Ping sends a ping command to verify the connection is alive.
func Ping(ctx context.Context, client *mongo.Client) error {
	if err := client.Ping(ctx, nil); err != nil {
		return fmt.Errorf("mongo.Ping: %w", err)
	}
	return nil
}

// Disconnect gracefully closes the MongoDB connection.
func Disconnect(client *mongo.Client) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return client.Disconnect(ctx)
}
