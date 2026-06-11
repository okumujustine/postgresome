package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/okumujustine/postgresome/internal/agent"
)

func main() {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL environment variable is required")
	}

	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	runner := agent.NewRunner(databaseURL, 30*time.Second)

	log.Println("starting Postgresome agent")

	if err := runner.Start(ctx); err != nil && err != context.Canceled {
		log.Fatalf("agent failed: %v", err)
	}

	log.Println("Postgresome agent stopped")
}
