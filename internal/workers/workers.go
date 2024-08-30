package workers

import (
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/riverqueue/river"
	"github.com/vscodethemes/backend/internal/marketplace"
)

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
