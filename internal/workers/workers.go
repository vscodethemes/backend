package workers

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/riverqueue/river"
	"github.com/riverqueue/river/rivertype"
	"github.com/vscodethemes/backend/internal/marketplace"
)

// Workers

type RegisterWorkersConfig struct {
	Registry          *river.Workers
	Directory         string
	DisableCleanup    bool
	ObjectStoreClient *s3.Client
	ObjectStoreBucket string
	CDNBaseUrl        string
	DBPool            *pgxpool.Pool
}

func RegisterWorkers(cfg RegisterWorkersConfig) error {
	river.AddWorker(cfg.Registry, &ScanExtensionsWorker{
		Marketplace: marketplace.NewClient(),
		DBPool:      cfg.DBPool,
	})

	river.AddWorker(cfg.Registry, &SyncExtensionWorker{
		Marketplace:       marketplace.NewClient(),
		Directory:         cfg.Directory,
		DisableCleanup:    cfg.DisableCleanup,
		ObjectStoreClient: cfg.ObjectStoreClient,
		ObjectStoreBucket: cfg.ObjectStoreBucket,
		CDNBaseUrl:        cfg.CDNBaseUrl,
		DBPool:            cfg.DBPool,
	})

	return nil
}

// Queues

const (
	ScanExtensionsQueue        = "scan-extensions"
	SyncExtensionPriorityQueue = "sync-extension-priority"
	SyncExtensionBackfillQueue = "sync-extension-backfill"
)

func QueueConfig() map[string]river.QueueConfig {
	return map[string]river.QueueConfig{
		SyncExtensionPriorityQueue: {MaxWorkers: 1},
		SyncExtensionBackfillQueue: {MaxWorkers: 1},
		ScanExtensionsQueue:        {MaxWorkers: 1},
	}
}

// Error handling

type ErrorHandler struct{}

func (*ErrorHandler) HandleError(ctx context.Context, job *rivertype.JobRow, err error) *river.ErrorHandlerResult {
	fmt.Printf("Job errored with: %s\n", err)
	return nil
}

func (*ErrorHandler) HandlePanic(ctx context.Context, job *rivertype.JobRow, panicVal any, trace string) *river.ErrorHandlerResult {
	fmt.Printf("Job panicked with: %v\n", panicVal)
	fmt.Printf("Stack trace: %s\n", trace)
	return nil
}
