package workers

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/gommon/log"
	"github.com/riverqueue/river"
	"github.com/vscodethemes/backend/internal/marketplace"
	"github.com/vscodethemes/backend/internal/marketplace/qo"
)

type ScanExtensionsArgs struct {
	MaxExtensions int
	SortBy        qo.QueryOptionSortBy
	SortDirection qo.QueryOptionDirection
}

func (ScanExtensionsArgs) Kind() string {
	return "scanExtensions"
}

func (ScanExtensionsArgs) InsertOpts() river.InsertOpts {
	return river.InsertOpts{
		Queue:       ScanExtensionsQueue,
		MaxAttempts: 5,
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

	extensionsScanned := 0
	pageNumber := 1
	pageSize := 50

	for extensionsScanned < job.Args.MaxExtensions {
		queryResults, err := w.Marketplace.NewQuery(ctx,
			qo.WithSortBy(job.Args.SortBy),
			qo.WithDirection(job.Args.SortDirection),
			qo.WithCriteria(qo.FilterTypeCategory, "Themes"),
			qo.WithCriteria(qo.FilterTypeUnknown8, "Microsoft.VisualStudio.Code"),
			qo.WithCriteria(qo.FilterTypeUnknown10, "target:\"Microsoft.VisualStudio.Code\" "),
			qo.WithCriteria(qo.FilterTypeUnknown12, "37888"),
			qo.WithPageNumber(pageNumber),
			qo.WithPageSize(pageSize),
		)
		if err != nil {
			return fmt.Errorf("failed to query marketplace: %w", err)
		}

		if len(queryResults) == 0 {
			break
		}

		pageNumber++

		batch := []river.InsertManyParams{}
		for _, extension := range queryResults {
			extensionsScanned++

			if extensionsScanned > job.Args.MaxExtensions {
				continue
			}

			log.Infof("Adding extension to batch: %s.%s", extension.Publisher.PublisherName, extension.ExtensionName)

			// TODO: Add option to stop scanning if we reach an extension exists with the same published date. This
			// would be used when scanning by most recently published periodically.

			batch = append(batch, river.InsertManyParams{
				Args: SyncExtensionArgs{
					PublisherName: extension.Publisher.PublisherName,
					ExtensionName: extension.ExtensionName,
				},
			})
		}

		count, err := client.InsertMany(ctx, batch)
		if err != nil {
			return fmt.Errorf("failed to insert job: %w", err)
		}

		log.Infof("Scanned %d extensions", count)
	}

	return nil
}
