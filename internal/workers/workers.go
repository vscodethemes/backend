package workers

import (
	"github.com/riverqueue/river"
	"github.com/vscodethemes/backend/internal/marketplace"
)

func RegisterWorkers(workersRegistry *river.Workers, directory string, disableCleanup bool) error {
	river.AddWorker(workersRegistry, &SyncExtensionWorker{
		Marketplace:    marketplace.NewClient(),
		Directory:      directory,
		DisableCleanup: disableCleanup,
	})

	return nil
}
