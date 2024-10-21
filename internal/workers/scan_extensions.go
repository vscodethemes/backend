package workers

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/gommon/log"
	"github.com/riverqueue/river"
	"github.com/vscodethemes/backend/internal/db"
	"github.com/vscodethemes/backend/internal/marketplace"
	"github.com/vscodethemes/backend/internal/marketplace/qo"
)

type ScanPriority string

const (
	ScanPriorityHigh ScanPriority = "high"
	ScanPriorityLow  ScanPriority = "low"
)

type ScanExtensionsArgs struct {
	MaxExtensions            int                     `json:"maxExtensions"`
	SortBy                   qo.QueryOptionSortBy    `json:"sortBy"`
	SortDirection            qo.QueryOptionDirection `json:"sortDirection"`
	Priority                 ScanPriority            `json:"priority"`
	BatchSize                int                     `json:"batchSize"`
	StopAtEqualPublishedDate bool                    `json:"stopAtEqualPublishedDate"`
	Force                    bool                    `json:"force"`
}

func (ScanExtensionsArgs) Kind() string {
	return "scanExtensions"
}

func (ScanExtensionsArgs) InsertOpts() river.InsertOpts {
	return river.InsertOpts{
		Queue:       ScanExtensionsQueue,
		MaxAttempts: 1,
	}
}

type ScanExtensionsWorker struct {
	river.WorkerDefaults[ScanExtensionsArgs]
	Marketplace *marketplace.Client
	DBPool      *pgxpool.Pool
}

func (w *ScanExtensionsWorker) Timeout(*river.Job[ScanExtensionsArgs]) time.Duration {
	return 5 * time.Minute
}

func (w *ScanExtensionsWorker) Work(ctx context.Context, job *river.Job[ScanExtensionsArgs]) error {
	client, err := river.ClientFromContextSafely[pgx.Tx](ctx)
	if err != nil {
		return fmt.Errorf("error getting client from context: %w", err)
	}

	insertQueue := SyncExtensionLowPriorityQueue
	if job.Args.Priority == ScanPriorityHigh {
		insertQueue = SyncExtensionHighPriorityQueue
	}

	batchSize := 50
	if job.Args.BatchSize > 0 {
		batchSize = job.Args.BatchSize
	}

	queries := db.New(w.DBPool)

	extensionsScanned := 0
	pageNumber := 1
	stopScanning := false
	for !stopScanning {
		log.Infof("Scanning page %d", pageNumber)

		// Add a delay to avoid rate limiting from the martketplace API.
		time.Sleep(2 * time.Second)

		queryResults, err := w.Marketplace.NewQuery(ctx,
			qo.WithSortBy(job.Args.SortBy),
			qo.WithDirection(job.Args.SortDirection),
			qo.WithCriteria(qo.FilterTypeCategory, "Themes"),
			qo.WithCriteria(qo.FilterTypeUnknown8, "Microsoft.VisualStudio.Code"),
			qo.WithCriteria(qo.FilterTypeUnknown10, "target:\"Microsoft.VisualStudio.Code\" "),
			qo.WithCriteria(qo.FilterTypeUnknown12, "37888"),
			qo.WithPageNumber(pageNumber),
			qo.WithPageSize(batchSize),
		)
		if err != nil {
			return fmt.Errorf("failed to query marketplace: %w", err)
		}

		if len(queryResults) == 0 {
			log.Infof("No more extensions found, stopping scan")
			break
		}

		pageNumber++

		batch := []river.InsertManyParams{}
		for _, extension := range queryResults {
			if extensionsScanned >= job.Args.MaxExtensions {
				log.Infof("Reached max extensions, stopping scan")
				stopScanning = true
				break
			}

			// If the job is configured to stop at the first extension with the same published date,
			// check if the extension is up to date and stop scanning if it is. This is useful when
			// sorting by last updated data.
			if job.Args.StopAtEqualPublishedDate {
				isUpToDate, err := isExtensionUpToDate(ctx, queries, extension)
				if err != nil {
					return fmt.Errorf("failed to check if extension is up to date: %w", err)
				}

				if isUpToDate {
					log.Infof("Extension %s.%s is up to date, stopping scan", extension.Publisher.PublisherName, extension.ExtensionName)
					stopScanning = true
					break
				}
			}

			log.Debugf("Adding extension to batch: %s.%s", extension.Publisher.PublisherName, extension.ExtensionName)

			batch = append(batch, river.InsertManyParams{
				Args: SyncExtensionArgs{
					PublisherName: extension.Publisher.PublisherName,
					ExtensionName: extension.ExtensionName,
					Force:         job.Args.Force,
				},
				InsertOpts: &river.InsertOpts{
					Queue: insertQueue,
				},
			})

			extensionsScanned++
		}

		if len(batch) > 0 {
			if _, err = client.InsertMany(ctx, batch); err != nil {
				return fmt.Errorf("failed to insert job: %w", err)
			}
		}

		log.Infof("Scanned %d extensions in batch, %d total", len(batch), extensionsScanned)
	}

	return nil
}
