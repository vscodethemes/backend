package handlers

import (
	"context"
	"fmt"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/jackc/pgx/v5"
	"github.com/riverqueue/river/rivertype"
	"github.com/vscodethemes/backend/internal/api/middleware"
	"github.com/vscodethemes/backend/internal/workers"
)

var SyncExtensionOperation = huma.Operation{
	OperationID: "post-extensions-sync",
	Method:      http.MethodPost,
	Path:        "/extensions/{publisher}/{name}/sync",
	Summary:     "Sync Extension",
	Description: "Sync an extension by it's slug. Returns the sync job.",
	Tags:        []string{"Extensions"},
	Errors:      []int{http.StatusBadRequest},
	Security: []map[string][]string{
		middleware.BearerAuthSecurity("extension:write"),
	},
}

type SyncExtensionInput struct {
	PublisherName string `path:"publisher" example:"sdras" doc:"The publisher name"`
	ExtensionName string `path:"name" example:"night-owl" doc:"The extension name"`
}

type SyncExtensionOutput struct {
	Body struct {
		Job Job `json:"job"`
	}
}

func (h Handler) SyncExtension(ctx context.Context, input *SyncExtensionInput) (*SyncExtensionOutput, error) {
	var job *rivertype.JobRow
	err := pgx.BeginFunc(ctx, h.DBPool, func(tx pgx.Tx) error {
		result, err := h.RiverClient.InsertTx(ctx, tx, workers.SyncExtensionArgs{
			PublisherName: input.PublisherName,
			ExtensionName: input.ExtensionName,
		}, nil)
		if err != nil {
			return fmt.Errorf("failed to insert job: %w", err)
		}

		job = result.Job

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to sync extension: %w", err)
	}

	if job == nil {
		return nil, huma.NewError(http.StatusNotFound, "Job not found")
	}

	resp := &SyncExtensionOutput{}
	resp.Body.Job = mapRiverJobToJob(*job)

	return resp, nil
}
