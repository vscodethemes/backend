package workers

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/gommon/log"
	"github.com/riverqueue/river"
	"github.com/vscodethemes/backend/internal/marketplace"
	"github.com/vscodethemes/backend/internal/marketplace/qo"
)

type ScanMostInstalled struct {
	MaxExtensions int
}

func (ScanMostInstalled) Kind() string {
	return "scanMostInstalled"
}

func (ScanMostInstalled) InsertOpts() river.InsertOpts {
	return river.InsertOpts{
		Queue:       ScanMostInstalledQueue,
		MaxAttempts: 5,
	}
}

type ScanMostInstalledWorker struct {
	river.WorkerDefaults[ScanMostInstalled]
	Marketplace *marketplace.Client
	DBPool      *pgxpool.Pool
}

func (w *ScanMostInstalledWorker) Work(ctx context.Context, job *river.Job[ScanMostInstalled]) error {
	client, err := river.ClientFromContextSafely[pgx.Tx](ctx)
	if err != nil {
		return fmt.Errorf("error getting client from context: %w", err)
	}

	extensionsScanned := 0
	pageNumber := 1
	pageSize := 50

	for extensionsScanned < job.Args.MaxExtensions {
		queryResults, err := w.Marketplace.NewQuery(ctx,
			qo.WithDirection(qo.DirectionDes),
			qo.WithSortBy(qo.SortByInstalls),
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
