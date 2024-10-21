package workers

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/gommon/log"
	"github.com/riverqueue/river"
	"github.com/vscodethemes/backend/internal/db"
	"github.com/vscodethemes/backend/internal/marketplace"
	"github.com/vscodethemes/backend/internal/marketplace/qo"
)

type UpdateExtensionStatsArgs struct {
	ExtensionName string `json:"extensionName"`
	PublisherName string `json:"publisherName"`
}

func (UpdateExtensionStatsArgs) Kind() string {
	return "updateExtensionStats"
}

func (UpdateExtensionStatsArgs) InsertOpts() river.InsertOpts {
	return river.InsertOpts{
		Queue:       UpdateExtenstionStatsQueue,
		MaxAttempts: 5,
	}
}

type UpdateExtensionStatsWorker struct {
	river.WorkerDefaults[UpdateExtensionStatsArgs]
	Marketplace *marketplace.Client
	DBPool      *pgxpool.Pool
}

func (w *UpdateExtensionStatsWorker) Timeout(*river.Job[UpdateExtensionStatsArgs]) time.Duration {
	return 1 * time.Minute
}

func (w *UpdateExtensionStatsWorker) Work(ctx context.Context, job *river.Job[UpdateExtensionStatsArgs]) error {
	extensionSlug := fmt.Sprintf("%s.%s", job.Args.PublisherName, job.Args.ExtensionName)
	log.Infof("Updating extension stats: %s", extensionSlug)

	// Add a delay to avoid rate limiting from the martketplace API.
	time.Sleep(2 * time.Second)

	// Fetch extension from the marketplace API.
	queryResults, err := w.Marketplace.NewQuery(ctx, qo.WithSlug(extensionSlug))
	if err != nil {
		return fmt.Errorf("failed to query marketplace: %w", err)
	}

	if len(queryResults) == 0 {
		return fmt.Errorf("extension not found")
	}

	extension := queryResults[0]

	upsertExtensionParams, err := convertUpsertExtensionParams(extension)
	if err != nil {
		return fmt.Errorf("failed to convert upsert extension params: %w", err)
	}

	log.Infof("Saving extension to database")

	queries := db.New(w.DBPool)
	if _, err = queries.UpsertExtension(ctx, upsertExtensionParams); err != nil {
		return fmt.Errorf("failed to upsert extension stats: %w", err)
	}

	log.Infof("Extension stats saved to database")

	return nil
}
