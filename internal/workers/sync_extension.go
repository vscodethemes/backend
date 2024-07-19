package workers

import (
	"context"
	"fmt"
	"os"
	"path"

	"github.com/labstack/gommon/log"
	"github.com/riverqueue/river"
	"github.com/vscodethemes/backend/internal/downloader"
	"github.com/vscodethemes/backend/internal/marketplace"
	"github.com/vscodethemes/backend/internal/marketplace/qo"
)

type SyncExtensionArgs struct {
	Slug string `json:"slug"`
}

func (SyncExtensionArgs) Kind() string { return "syncExtension" }

type SyncExtensionWorker struct {
	river.WorkerDefaults[SyncExtensionArgs]
	Marketplace    *marketplace.Client
	Directory      string
	DisableCleanup bool
}

func (w *SyncExtensionWorker) Work(ctx context.Context, job *river.Job[SyncExtensionArgs]) error {
	extensionSlug := job.Args.Slug
	log.Infof("Syncing extension package: %s", extensionSlug)

	// Fetch extension from the marketplace API.
	results, err := w.Marketplace.NewQuery(ctx, qo.WithSlug(extensionSlug))
	if err != nil {
		return fmt.Errorf("failed to query marketplace: %w", err)
	}

	if len(results) == 0 {
		return fmt.Errorf("extension not found")
	}

	extension := results[0]

	// Ensure there's a package URL for the extension.
	packageUrl := extension.GetPackageURL()
	if packageUrl == "" {
		return fmt.Errorf("extension package not found")
	}

	// Create a directory for the job to download the package.
	jobDir := path.Join(w.Directory, "jobs", fmt.Sprintf("%d", job.ID))
	err = os.MkdirAll(jobDir, os.ModePerm)
	if err != nil {
		return fmt.Errorf("failed to create job dir: %w", err)
	}
	if !w.DisableCleanup {
		defer func() {
			log.Infof("Cleaning up job directory: %s", jobDir)
			os.RemoveAll(jobDir)
		}()
	}

	// Download the extension package.
	d := downloader.New(jobDir, extensionSlug)

	log.Infof("Downloading package: %s", packageUrl)
	err = d.Download(ctx, packageUrl)
	if err != nil {
		return fmt.Errorf("failed to download package: %w", err)
	}

	log.Infof("Extracting package: %s", d.PackagePath)
	err = d.Extract()
	if err != nil {
		return fmt.Errorf("failed to extract package: %w", err)
	}

	// TODO: Parse the extension.
	// TODO: Generate preview images.
	// TODO: Upload images.
	// TODO: Save the extension to the database.

	return nil
}
