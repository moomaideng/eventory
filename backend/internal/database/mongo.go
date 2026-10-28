package database

import (
	"context"
	"errors"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.opentelemetry.io/contrib/instrumentation/go.mongodb.org/mongo-driver/v2/mongo/otelmongo"
)

// ConnectMongo initializes and validates a MongoDB client connection.
func ConnectMongo(ctx context.Context, uri string) (*mongo.Client, error) {
	if strings.TrimSpace(uri) == "" {
		return nil, errors.New("MONGO_URI is required")
	}

	// Keep command bodies out of spans: they hold user data (emails, form
	// answers). This is otelmongo's default today, but upstream plans to
	// re-enable it, so it stays explicit.
	monitor := otelmongo.NewMonitor(otelmongo.WithCommandAttributeDisabled(true))
	opts := options.Client().ApplyURI(uri).SetMonitor(monitor)
	client, err := mongo.Connect(opts)
	if err != nil {
		return nil, err
	}

	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := client.Ping(pingCtx, nil); err != nil {
		_ = client.Disconnect(context.Background())
		return nil, err
	}

	return client, nil
}
