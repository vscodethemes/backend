package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"log/slog"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/riverqueue/river"
	"github.com/riverqueue/river/riverdriver/riverpgxv5"
	"github.com/vscodethemes/backend/internal/workers"
)

func main() {
	dbUrl := flag.String("db-url", "", "Database URL")
	flag.Parse()

	if *dbUrl == "" {
		log.Fatal("Database URL is required")
	}

	dbPool, err := pgxpool.New(context.Background(), *dbUrl)
	if err != nil {
		log.Fatal(fmt.Errorf("failed to create db pool: %w", err))
	}
	defer dbPool.Close()

	// Register Workers.
	workersRegistry := river.NewWorkers()
	err = workers.RegisterWorkers(workersRegistry)
	if err != nil {
		log.Fatal(fmt.Errorf("failed to register workers: %w", err))
	}

	// Create river client.
	riverClient, err := river.NewClient(riverpgxv5.New(dbPool), &river.Config{
		Queues: map[string]river.QueueConfig{
			river.QueueDefault: {MaxWorkers: 100},
		},
		Workers: workersRegistry,
		Logger: slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelWarn,
		})),
	})
	if err != nil {
		log.Fatal(fmt.Errorf("failed to create river client: %w", err))
	}

	if err := riverClient.Start(context.Background()); err != nil {
		log.Fatal(fmt.Errorf("failed to start river client: %w", err))
	}

	fmt.Println("Waiting for jobs...")

	// TODO: Handle signals to gracefully stop the river client.
	// https://riverqueue.com/docs/graceful-shutdown

	<-riverClient.Stopped()
}
