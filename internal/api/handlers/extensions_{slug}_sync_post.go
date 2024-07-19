package handlers

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/danielgtaylor/huma/v2"
	"github.com/jackc/pgx/v5"
	"github.com/riverqueue/river/rivertype"
	"github.com/vscodethemes/backend/internal/db"
	"github.com/vscodethemes/backend/internal/workers"
)

var SyncExtensionBySlugOperation = huma.Operation{
	OperationID: "post-extensions-sync-by-slug",
	Method:      http.MethodPost,
	Path:        "/extensions/{slug}/sync",
	Summary:     "Sync Extension",
	Description: "Sync an extension by it's slug (ie. sdras.night-owl). Returns the sync job.",
	Tags:        []string{"Extensions"},
	Errors:      []int{http.StatusBadRequest},
}

type SyncExtensionInput struct {
	Slug string `path:"slug" example:"sdras.night-owl" doc:"Slug of the extension to sync"`
}

type SyncExtensionOutput struct {
	Body struct {
		Job Job `json:"job"`
	}
}

func (h Handler) SyncExtensionBySlug(ctx context.Context, input *SyncExtensionInput) (*SyncExtensionOutput, error) {
	fmt.Println("SyncExtensionBySlug", input.Slug)

	// Validate slug format.
	slugParts := strings.Split(input.Slug, ".")
	if len(slugParts) != 2 {
		return nil, huma.NewError(http.StatusBadRequest, "Invalid slug format")
	}

	var job *rivertype.JobRow
	err := db.Tx(ctx, h.DBPool, func(tx pgx.Tx) error {
		result, err := h.RiverClient.InsertTx(ctx, tx, workers.SyncExtensionArgs{
			Slug: input.Slug,
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

	resp := &SyncExtensionOutput{}
	resp.Body.Job.Id = job.ID
	resp.Body.Job.State = string(job.State)
	resp.Body.Job.CreatedAt = job.CreatedAt

	return resp, nil
}
