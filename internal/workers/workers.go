package workers

import (
	"github.com/riverqueue/river"
	"github.com/vscodethemes/backend/internal/marketplace"
)

func RegisterWorkers(workersRegistry *river.Workers) error {
	river.AddWorker(workersRegistry, &SyncExtensionWorker{
		marketplace: marketplace.NewClient(),
	})

	return nil
}
