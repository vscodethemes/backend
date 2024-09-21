package handlers

import (
	"context"
	"fmt"
	"math"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/jackc/pgx/v5"
	"github.com/riverqueue/river"
	"github.com/riverqueue/river/rivertype"
	"github.com/vscodethemes/backend/internal/api/middleware"
	"github.com/vscodethemes/backend/internal/marketplace/qo"
	"github.com/vscodethemes/backend/internal/workers"
)

var ScanExtensionsOperation = huma.Operation{
	OperationID: "post-extensions-scan",
	Method:      http.MethodPost,
	Path:        "/extensions/scan",
	Summary:     "Scan Extensions",
	Description: "Scan extensions from the marketplace.",
	Tags:        []string{"Extensions"},
	Errors:      []int{http.StatusBadRequest},
	Security: []map[string][]string{
		middleware.BearerAuthSecurity("extension:write"),
	},
}

type ScanExtensionsInput struct {
	Type          string `query:"type" default:"lastUpdated" example:"lastUpdated" doc:"Type of scan to perform."`
	MaxExtensions int    `query:"maxExtensions" example:"50" doc:"Maximum number of extensions to scan. If not provided, all extensions will be scanned."`
}

type ScanExtensions struct {
	Body struct {
		Job Job `json:"job"`
	}
}

func (h Handler) ScanExtensions(ctx context.Context, input *ScanExtensionsInput) (*ScanExtensions, error) {
	maxExtensions := math.MaxInt
	if input.MaxExtensions != 0 {
		maxExtensions = input.MaxExtensions
	}

	var jobArgs river.JobArgs
	if input.Type == "lastUpdated" || input.Type == "" {
		jobArgs = workers.ScanExtensionsArgs{
			MaxExtensions: maxExtensions,
			SortBy:        qo.SortByLastUpdated,
			SortDirection: qo.DirectionDesc,
		}
	} else if input.Type == "mostInstalled" {
		jobArgs = workers.ScanExtensionsArgs{
			MaxExtensions: maxExtensions,
			SortBy:        qo.SortByInstalls,
			SortDirection: qo.DirectionDesc,
		}
	} else {
		return nil, huma.Error400BadRequest(fmt.Sprintf("Unknown scan type: %s", input.Type))
	}

	var job *rivertype.JobRow
	err := pgx.BeginFunc(ctx, h.DBPool, func(tx pgx.Tx) error {
		result, err := h.RiverClient.InsertTx(ctx, tx, jobArgs, nil)
		if err != nil {
			return fmt.Errorf("failed to insert job: %w", err)
		}

		job = result.Job

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to scan extensions: %w", err)
	}

	if job == nil {
		return nil, huma.NewError(http.StatusNotFound, "Job not found")
	}

	resp := &ScanExtensions{}
	resp.Body.Job = mapRiverJobToJob(*job)

	return resp, nil
}
