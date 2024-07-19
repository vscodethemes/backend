package workers

import (
	"context"
	"fmt"

	"github.com/riverqueue/river"
	"github.com/vscodethemes/backend/internal/marketplace"
)

type SyncExtensionArgs struct {
	Slug string `json:"slug"`
}

func (SyncExtensionArgs) Kind() string { return "syncExtension" }

type SyncExtensionWorker struct {
	river.WorkerDefaults[SyncExtensionArgs]
	marketplace *marketplace.Client
}

func (w *SyncExtensionWorker) Work(ctx context.Context, job *river.Job[SyncExtensionArgs]) error {
	fmt.Printf("Sync extension: %+v\n", job.Args.Slug)

	// Fetch extension from the marketplace API.
	results, err := w.marketplace.NewQuery(ctx, marketplace.WithSlug(job.Args.Slug))
	if err != nil {
		return fmt.Errorf("failed to query marketplace: %w", err)
	}

	fmt.Println("Results:", results)

	// TODO: Download the extension.
	// TODO: Defer cleanup of downloaded extension.
	// TODO: Extract the extension.
	// TODO: Parse the extension.
	// TODO: Generate preview images.
	// TODO: Upload images.
	// TODO: Save the extension to the database.

	return nil
}
