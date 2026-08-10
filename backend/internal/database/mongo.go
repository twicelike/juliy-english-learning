package database

import (
	"context"
	"fmt"
	"time"

	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/v2/event"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// Connect establishes a connection to MongoDB and verifies it with a Ping.
// Returns the mongo.Client so the caller can select the target database.
func Connect(uri string) (*mongo.Client, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	logSlowQuery := func(duration time.Duration, commandName string) {
		if duration > 100*time.Millisecond {
			logger := log.Warn()
			if duration > 1000*time.Millisecond {
				logger = log.Error()
			} else if duration > 300*time.Millisecond {
				logger = log.Error()
			}
			logger.Str("command", commandName).
				Dur("duration", duration).
				Msg("Slow MongoDB Query")
		}
	}

	monitor := &event.CommandMonitor{
		Started: func(ctx context.Context, evt *event.CommandStartedEvent) {},
		Succeeded: func(ctx context.Context, evt *event.CommandSucceededEvent) {
			logSlowQuery(evt.Duration, evt.CommandName)
		},
		Failed: func(ctx context.Context, evt *event.CommandFailedEvent) {
			logSlowQuery(evt.Duration, evt.CommandName)
		},
	}

	clientOptions := options.Client().
		ApplyURI(uri).
		SetMaxPoolSize(100).
		SetMinPoolSize(10).
		SetMaxConnIdleTime(60 * time.Second).
		SetMonitor(monitor)

	client, err := mongo.Connect(clientOptions)
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
