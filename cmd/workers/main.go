package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"log/slog"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/riverqueue/river"
	"github.com/riverqueue/river/riverdriver/riverpgxv5"
	"github.com/vscodethemes/backend/internal/workers"
)

func main() {
	dbUrl := flag.String("database-url", "", "Database URL")
	dir := flag.String("dir", "/tmp", "Directory")
	objectStoreEndpoint := flag.String("object-store-endpoint", "http://s3.localhost.localstack.cloud:4566", "Object store endpoint")
	objectStoreBucket := flag.String("object-store-bucket", "images", "Object store bucket to upload images to")
	objectStoreRegion := flag.String("object-store-region", "us-east-1", "Object store region")
	objectStoreAccessKeyID := flag.String("object-store-access-key-id", "test", "Object store access key ID")
	objectStoreAccessKeySecret := flag.String("object-store-access-key-secret", "test", "Object store access key secret")
	cdnBaseUrl := flag.String("cdn-base-url", "http://s3.localhost.localstack.cloud:4566/images", "CDN base URL")
	disableCleanup := flag.Bool("disable-cleanup", false, "Disable cleanup")
	flag.Parse()

	if *dbUrl == "" {
		log.Fatal("Database URL is required")
	}

	dbPool, err := pgxpool.New(context.Background(), *dbUrl)
	if err != nil {
		log.Fatal(fmt.Errorf("failed to create db pool: %w", err))
	}
	defer dbPool.Close()

	objectStoreCreds := credentials.NewStaticCredentialsProvider(*objectStoreAccessKeyID, *objectStoreAccessKeySecret, "")

	objectStoreCfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithCredentialsProvider(objectStoreCreds),
		config.WithRegion(*objectStoreRegion),
		// For R2
		// config.WithRegion("auto"),
	)
	if err != nil {
		log.Fatal(fmt.Errorf("failed to load object store config: %w", err))
	}

	objectStoreClient := s3.NewFromConfig(objectStoreCfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(*objectStoreEndpoint)
		// For R2
		// o.BaseEndpoint = aws.String(fmt.Sprintf("https://%s.r2.cloudflarestorage.com", accountId))
	})

	// Register Workers.
	workersRegistry := river.NewWorkers()
	err = workers.RegisterWorkers(workers.RegisterWorkersConfig{
		Registry:          workersRegistry,
		Directory:         *dir,
		DisableCleanup:    *disableCleanup,
		ObjectStoreClient: objectStoreClient,
		ObjectStoreBucket: *objectStoreBucket,
		CDNBaseUrl:        *cdnBaseUrl,
		DBPool:            dbPool,
	})
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
