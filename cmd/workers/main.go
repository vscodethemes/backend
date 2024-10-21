package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

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
	maxExtensions := flag.Int("max-extensions", 0, "Maximum number of extensions to scan, 0 for all")
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
	)
	if err != nil {
		log.Fatal(fmt.Errorf("failed to load object store config: %w", err))
	}

	objectStoreClient := s3.NewFromConfig(objectStoreCfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(*objectStoreEndpoint)
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
		Queues:       workers.QueueConfig(),
		PeriodicJobs: workers.PeriodicJobs(*maxExtensions),
		Workers:      workersRegistry,
		Logger: slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelWarn,
		})),
		ErrorHandler: &workers.ErrorHandler{},
	})
	if err != nil {
		log.Fatal(fmt.Errorf("failed to create river client: %w", err))
	}

	if err := riverClient.Start(context.Background()); err != nil {
		log.Fatal(fmt.Errorf("failed to start river client: %w", err))
	}

	fmt.Println("Waiting for jobs...")

	// Handle signals to gracefully stop the river client.
	// https://riverqueue.com/docs/graceful-shutdown
	sigintOrTerm := make(chan os.Signal, 1)
	signal.Notify(sigintOrTerm, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigintOrTerm
		fmt.Printf("Received SIGINT/SIGTERM; initiating soft stop (try to wait for jobs to finish)\n")

		softStopCtx, softStopCtxCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer softStopCtxCancel()

		go func() {
			select {
			case <-sigintOrTerm:
				fmt.Println("Received SIGINT/SIGTERM again; initiating hard stop (cancel everything)")
				softStopCtxCancel()
			case <-softStopCtx.Done():
				fmt.Println("Soft stop timeout; initiating hard stop (cancel everything)")
			}
		}()

		err := riverClient.Stop(softStopCtx)
		if err != nil && !errors.Is(err, context.DeadlineExceeded) && !errors.Is(err, context.Canceled) {
			panic(err)
		}
		if err == nil {
			fmt.Println("Soft stop succeeded")
			return
		}

		hardStopCtx, hardStopCtxCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer hardStopCtxCancel()

		// As long as all jobs respect context cancellation, StopAndCancel will
		// always work. However, in the case of a bug where a job blocks despite
		// being cancelled, it may be necessary to either ignore River's stop
		// result (what's shown here) or have a supervisor kill the process.
		err = riverClient.StopAndCancel(hardStopCtx)
		if err != nil && errors.Is(err, context.DeadlineExceeded) {
			fmt.Println("Hard stop timeout; ignoring stop procedure and exiting unsafely")
		} else if err != nil {
			panic(err)
		}

		// hard stop succeeded
	}()

	<-riverClient.Stopped()
}
