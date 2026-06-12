package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/okumujustine/postgresome/internal/api"
	"github.com/okumujustine/postgresome/internal/storage"
)

func main() {
	addr := os.Getenv("API_ADDR")
	if addr == "" {
		addr = ":9090"
	}

	databaseURL := os.Getenv("POSTGRESOME_DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("POSTGRESOME_DATABASE_URL environment variable is required")
	}

	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	pool, err := storage.Connect(ctx, databaseURL)
	if err != nil {
		log.Fatalf("failed to connect to Postgresome database: %v", err)
	}
	defer pool.Close()

	log.Println("connected to Postgresome database")

	server := api.NewServer(addr, pool)

	log.Printf("starting Postgresome API on %s", addr)

	if err := server.Start(ctx); err != nil && err != context.Canceled {
		log.Fatalf("api server failed: %v", err)
	}

	log.Println("Postgresome API stopped")
}
