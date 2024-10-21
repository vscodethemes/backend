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
)

type UpdateAllExtensionsStatsArgs struct{}

func (UpdateAllExtensionsStatsArgs) Kind() string {
	return "updateAllExtensionsStats"
}

func (UpdateAllExtensionsStatsArgs) InsertOpts() river.InsertOpts {
	return river.InsertOpts{
		Queue:       UpdateExtenstionStatsQueue,
		MaxAttempts: 5,
	}
}

type UpdateAllExtensionsStatsWorker struct {
	river.WorkerDefaults[UpdateAllExtensionsStatsArgs]
	DBPool *pgxpool.Pool
}

func (w *UpdateAllExtensionsStatsWorker) Timeout(*river.Job[UpdateAllExtensionsStatsArgs]) time.Duration {
	return 5 * time.Minute
}

func (w *UpdateAllExtensionsStatsWorker) Work(ctx context.Context, job *river.Job[UpdateAllExtensionsStatsArgs]) error {
	fmt.Println("Getting all extensions for update")

	client, err := river.ClientFromContextSafely[pgx.Tx](ctx)
	if err != nil {
		return fmt.Errorf("error getting client from context: %w", err)
	}

	// Get all extensions from the database.
	queries := db.New(w.DBPool)
	extensions, err := queries.GetAllExtensionsForUpdate(ctx)
	if err != nil {
		return fmt.Errorf("failed to get all extensions: %w", err)
	}

	// Update the stats for each extension.
	batch := []river.InsertManyParams{}
	for _, extension := range extensions {
		batch = append(batch, river.InsertManyParams{
			Args: UpdateExtensionStatsArgs{
				PublisherName: extension.PublisherName,
				ExtensionName: extension.Name,
			},
			InsertOpts: &river.InsertOpts{
				Queue: UpdateExtenstionStatsQueue,
			},
		})
	}

	if len(batch) > 0 {
		if _, err = client.InsertMany(ctx, batch); err != nil {
			return fmt.Errorf("failed to insert job: %w", err)
		}
	}

	log.Infof("Updating %d extensions in batch", len(batch))

	return nil
}
