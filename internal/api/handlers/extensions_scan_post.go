package handlers

import (
	"context"
	"fmt"
	"math"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/jackc/pgx/v5"
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
	Priority                 workers.ScanPriority `query:"priority" default:"low" example:"low" doc:"Priority of the scan, set to 'low' or 'high'."`
	SortBy                   string               `query:"sortBy" default:"lastUpdated" example:"lastUpdated" doc:"Type of scan to perform, set to 'lastUpdated' or 'mostInstalled'."`
	SortDirection            string               `query:"direction" default:"desc" example:"desc" doc:"Direction of the sort, set to 'asc' or 'desc'."`
	BatchSize                int                  `query:"batchSize" default:"50" example:"100" doc:"Number of extensions to scan in each batch."`
	MaxExtensions            int                  `query:"maxExtensions" example:"200" doc:"Maximum number of extensions to scan. If not provided, all extensions will be scanned."`
	StopAtEqualPublishedDate bool                 `query:"stopAtEqualPublishedDate" default:"false" example:"true" doc:"Stop scanning when the published date is equal to the last scanned extension."`
	ForceUpdate              bool                 `query:"forceUpdate" default:"false" example:"true" doc:"Force the extension to update even if publisehd date is equal."`
}

type ScanExtensions struct {
	Body struct {
		Job Job `json:"job"`
	}
}

func (h Handler) ScanExtensions(ctx context.Context, input *ScanExtensionsInput) (*ScanExtensions, error) {
	priority := workers.ScanPriorityLow
	if input.Priority == workers.ScanPriorityHigh {
		priority = workers.ScanPriorityHigh
	}

	sortBy := qo.SortByLastUpdated
	if input.SortBy == "mostInstalled" {
		sortBy = qo.SortByInstalls
	}

	sortDirection := qo.DirectionDesc
	if input.SortDirection == "asc" {
		sortDirection = qo.DirectionAsc
	}

	batchSize := 50
	if input.BatchSize > 0 {
		batchSize = input.BatchSize
	}

	maxExtensions := math.MaxInt
	if input.MaxExtensions != 0 {
		maxExtensions = input.MaxExtensions
	}

	var job *rivertype.JobRow
	err := pgx.BeginFunc(ctx, h.DBPool, func(tx pgx.Tx) error {
		result, err := h.RiverClient.InsertTx(ctx, tx, workers.ScanExtensionsArgs{
			Priority:                 priority,
			SortBy:                   sortBy,
			SortDirection:            sortDirection,
			BatchSize:                batchSize,
			MaxExtensions:            maxExtensions,
			StopAtEqualPublishedDate: input.StopAtEqualPublishedDate,
			Force:                    input.ForceUpdate,
		}, nil)
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
