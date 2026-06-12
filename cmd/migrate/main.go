package main

import (
	"context"
	"log"
	"os"

	"github.com/okumujustine/postgresome/internal/migrate"
	"github.com/okumujustine/postgresome/internal/storage"
)

const migrationsDir = "migrations"

func main() {
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

	log.Println("connected to Postgresome database")

	if err := migrate.Run(ctx, pool, migrationsDir); err != nil {
		log.Fatalf("migration failed: %v", err)
	}

	log.Println("migrations complete")
}
