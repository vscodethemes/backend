package workers

import (
	"context"
	"fmt"
	"os"
	"path"
	"path/filepath"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/labstack/gommon/log"
	gonanoid "github.com/matoous/go-nanoid/v2"
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
	Marketplace       *marketplace.Client
	Directory         string
	DisableCleanup    bool
	ObjectStoreClient *s3.Client
	ObjectStoreBucket string
	CDNBaseUrl        string
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
	group, imagesCtx := errgroup.WithContext(ctx)
	group.SetLimit(10)

	imagesResults := make([]*cli.GenerateImagesResult, len(info.ThemeContributes))
	for i, theme := range info.ThemeContributes {
		group.Go(func() error {
			log.Infof("Generating images for theme: %s", theme.Path)
			result, err := cli.GenerateImages(imagesCtx, extensionPath, theme, imagesPath)
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

	// Upload images for each theme concurrency, up to a max of 10 subroutines.
	group, uploadCtx := errgroup.WithContext(ctx)
	group.SetLimit(10)

	for _, result := range imagesResults {
		group.Go(func() error {
			for _, language := range result.Languages {
				file, err := os.Open(language.SvgPath)
				if err != nil {
					return fmt.Errorf("failed to open file: %w", err)
				}
				defer file.Close()

				id, err := gonanoid.New()
				if err != nil {
					return fmt.Errorf("failed to generate nanoid: %w", err)
				}

				svgFileName := fmt.Sprintf("%s.svg", id)
				svgObjectKey := fmt.Sprintf("%s/%s", extensionSlug, svgFileName)

				log.Debugf("Uploading SVG image at %s to %s", language.SvgPath, svgObjectKey)

				_, err = w.ObjectStoreClient.PutObject(uploadCtx, &s3.PutObjectInput{
					Bucket:       aws.String(w.ObjectStoreBucket),
					Key:          aws.String(svgObjectKey),
					Body:         file,
					ContentType:  aws.String("image/svg+xml"),
					CacheControl: aws.String("public, max-age=31536000"),
				})
				if err != nil {
					return fmt.Errorf("failed to upload svg file to %s: %w", svgObjectKey, err)
				}

				svgImageUrl := fmt.Sprintf("%s/%s", w.CDNBaseUrl, svgObjectKey)
				log.Infof("SVG image uploaded: %s", svgImageUrl)

				file, err = os.Open(language.PngPath)
				if err != nil {
					return fmt.Errorf("failed to open file: %w", err)
				}

				pngFileName := fmt.Sprintf("%s.png", id)
				pngObjectKey := fmt.Sprintf("%s/%s", extensionSlug, pngFileName)

				log.Debugf("Uploading PNG image at %s to %s", language.PngPath, pngObjectKey)

				_, err = w.ObjectStoreClient.PutObject(uploadCtx, &s3.PutObjectInput{
					Bucket:       aws.String(w.ObjectStoreBucket),
					Key:          aws.String(pngObjectKey),
					Body:         file,
					ContentType:  aws.String("image/png"),
					CacheControl: aws.String("public, max-age=31536000"),
				})
				if err != nil {
					return fmt.Errorf("failed to upload png file to %s: %w", pngObjectKey, err)
				}

				pngImageUrl := fmt.Sprintf("%s/%s", w.CDNBaseUrl, pngObjectKey)
				log.Infof("PNG image uploaded: %s", pngImageUrl)
			}

			return nil
		})
	}

	if err := group.Wait(); err != nil {
		return err
	}

	// TODO: Map the full image urls to each langauge.

	// TODO: Save the extension to the database.

	return nil
}
