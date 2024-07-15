//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen --config=./config.yaml ./api.yaml

package api

import (
	"fmt"
	"net/http"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
	"github.com/riverqueue/river"
	"github.com/vscodethemes/backend/internal/workers"
)

type Server struct {
	dbPool      *pgxpool.Pool
	riverClient *river.Client[pgx.Tx]
}

func NewServer(dbPool *pgxpool.Pool, riverClient *river.Client[pgx.Tx]) Server {
	return Server{
		dbPool:      dbPool,
		riverClient: riverClient,
	}
}

func (s Server) SyncExtensionBySlug(ctx echo.Context, extensionSlug string) error {
	requestCtx := ctx.Request().Context()

	fmt.Println("SyncExtensionBySlug", extensionSlug)

	tx, err := s.dbPool.Begin(requestCtx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		err := tx.Rollback(requestCtx)
		if err != nil {
			fmt.Printf("failed to rollback transaction: %v\n", err)
		}
	}()

	result, err := s.riverClient.InsertTx(requestCtx, tx, workers.SyncExtensionArgs{
		Slug: extensionSlug,
	}, nil)
	if err != nil {
		return fmt.Errorf("failed to insert job: %w", err)
	}

	if err := tx.Commit(requestCtx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return ctx.JSON(http.StatusOK, ExtensionSyncJob{
		Id:        result.Job.ID,
		State:     string(result.Job.State),
		CreatedAt: result.Job.CreatedAt,
	})
}
