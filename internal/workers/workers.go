package workers

import (
	"context"
	"fmt"

	"github.com/riverqueue/river"
)

func RegisterWorkers(workersRegistry *river.Workers) error {
	river.AddWorker(workersRegistry, &SyncExtensionWorker{})

	return nil
}

type SyncExtensionArgs struct {
	Slug string `json:"slug"`
}

func (SyncExtensionArgs) Kind() string { return "syncExtension" }

type SyncExtensionWorker struct {
	river.WorkerDefaults[SyncExtensionArgs]
}

func (w *SyncExtensionWorker) Work(ctx context.Context, job *river.Job[SyncExtensionArgs]) error {
	fmt.Printf("Sync extension: %+v\n", job.Args.Slug)
	return nil
}
