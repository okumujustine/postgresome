package main

import (
	"context"
	"log"
	"os"
	"strings"

	"github.com/okumujustine/postgresome/backend/internal/devseed"
	"github.com/okumujustine/postgresome/backend/internal/secrets"
	"github.com/okumujustine/postgresome/backend/internal/storage"
)

func main() {
	if !strings.EqualFold(os.Getenv("POSTGRESOME_ENABLE_DEV_SEED"), "true") {
		log.Println("development seed disabled; skipping")
		return
	}

	databaseURL := os.Getenv("POSTGRESOME_DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("POSTGRESOME_DATABASE_URL environment variable is required")
	}

	ctx := context.Background()

	pool, err := storage.Connect(ctx, databaseURL)
	if err != nil {
		log.Fatalf("failed to connect to Postgresome database: %v", err)
	}
	defer pool.Close()

	connectionProtector, err := secrets.NewConnectionProtectorFromEnv()
	if err != nil {
		log.Fatalf("failed to initialize source secret protection: %v", err)
	}

	if err := devseed.Run(ctx, pool, connectionProtector); err != nil {
		log.Fatalf("development seed failed: %v", err)
	}

	log.Println("development seed complete")
}
