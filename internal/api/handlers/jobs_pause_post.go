package handlers

import (
	"context"
	"fmt"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/vscodethemes/backend/internal/api/middleware"
)

var PauseJobsOperation = huma.Operation{
	OperationID: "post-jobs-pause",
	Method:      http.MethodPost,
	Path:        "/jobs/pause",
	Summary:     "Pause Jobs",
	Description: "Pause a specific queue of jobs or all queues.",
	Tags:        []string{"Jobs"},
	Errors:      []int{http.StatusBadRequest},
	Security: []map[string][]string{
		middleware.BearerAuthSecurity("jobs:write"),
	},
}

type PauseJobsInput struct {
	QueueName string `query:"queue" example:"sync-extension-priority" doc:"The queue name or '*' for all queues"`
}

type PauseJobsOutput struct {
	Body struct{}
}

func (h Handler) PauseJobs(ctx context.Context, input *PauseJobsInput) (*PauseJobsOutput, error) {
	if err := h.RiverClient.QueuePause(ctx, input.QueueName, nil); err != nil {
		return nil, fmt.Errorf("failed to pause queue '%s': %w", input.QueueName, err)
	}

	return &PauseJobsOutput{}, nil
}
