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

	agentID := os.Getenv("AGENT_ID")
	if agentID == "" {
		log.Fatal("AGENT_ID environment variable is required")
	}

	agentName := os.Getenv("AGENT_NAME")
	if agentName == "" {
		log.Fatal("AGENT_NAME environment variable is required")
	}

	agentEnvironment := os.Getenv("AGENT_ENVIRONMENT")
	if agentEnvironment == "" {
		log.Fatal("AGENT_ENVIRONMENT environment variable is required")
	}

	apiBaseURL := os.Getenv("POSTGRESOME_API_URL")
	if apiBaseURL == "" {
		log.Fatal("POSTGRESOME_API_URL environment variable is required")
	}

	databaseInstanceID := os.Getenv("DATABASE_INSTANCE_ID")
	if databaseInstanceID == "" {
		log.Fatal("DATABASE_INSTANCE_ID environment variable is required")
	}

	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	runner := agent.NewRunner(agent.Config{
		DatabaseURL:        databaseURL,
		Interval:           30 * time.Second,
		AgentID:            agentID,
		AgentName:          agentName,
		AgentEnvironment:   agentEnvironment,
		APIBaseURL:         apiBaseURL,
		DatabaseInstanceID: databaseInstanceID,
	})

	log.Println("starting Postgresome agent")

	if err := runner.Start(ctx); err != nil && err != context.Canceled {
		log.Fatalf("agent failed: %v", err)
	}

	log.Println("Postgresome agent stopped")
}
