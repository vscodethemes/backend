package workers

import (
	"context"
	"fmt"
	"os"
	"path"
	"path/filepath"

	"github.com/labstack/gommon/log"
	"github.com/riverqueue/river"
	"github.com/vscodethemes/backend/internal/cli"
	"github.com/vscodethemes/backend/internal/downloader"
	"github.com/vscodethemes/backend/internal/marketplace"
	"github.com/vscodethemes/backend/internal/marketplace/qo"
	"golang.org/x/sync/errgroup"
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
	queryResults, err := w.Marketplace.NewQuery(ctx, qo.WithSlug(extensionSlug))
	if err != nil {
		return fmt.Errorf("failed to query marketplace: %w", err)
	}

	if len(queryResults) == 0 {
		return fmt.Errorf("extension not found")
	}

	extension := queryResults[0]

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

	extensionPath, err := filepath.Abs(d.ExtractDir)
	if err != nil {
		return fmt.Errorf("failed to get absolute path for extension: %w", err)
	}

	log.Infof("Reading extension info: %s", extensionPath)
	info, err := cli.GetInfo(ctx, extensionPath)
	if err != nil {
		return fmt.Errorf("failed to get info: %w", err)
	}

	imagesPath, err := filepath.Abs(path.Join(jobDir, "images"))
	if err != nil {
		return fmt.Errorf("failed to get absolute path for images: %w", err)
	}

	// Generate images for each theme concurrency, up to a max of 10 subroutines.
	group, ctx := errgroup.WithContext(ctx)
	group.SetLimit(10)

	imagesResults := make([]*cli.GenerateImagesResult, len(info.ThemeContributes))
	for i, theme := range info.ThemeContributes {
		group.Go(func() error {
			log.Infof("Generating images for theme: %s", theme.Path)
			result, err := cli.GenerateImages(ctx, extensionPath, theme, imagesPath)
			if err != nil {
				return fmt.Errorf("failed to generate images for %s: %w", theme.Path, err)
			}

			imagesResults[i] = result
			return nil
		})
	}

	if err := group.Wait(); err != nil {
		return err
	}

	// TODO: Upload images. Run concurrency up to a max of 10 subroutines.
	// _, err := s3.PutObject(image.SVGPath)
	// _, err := s3.PutObject(image.PNGPath)
	for _, result := range imagesResults {
		for _, language := range result.Languages {
			fmt.Println("SvgPath: ", language.SvgPath)
			fmt.Println("PngPath: ", language.PngPath)
		}
	}

	// TODO: Save the extension to the database.

	return nil
}
