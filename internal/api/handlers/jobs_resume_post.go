package handlers

import (
	"context"
	"fmt"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/vscodethemes/backend/internal/api/middleware"
)

var ResumeJobsOperation = huma.Operation{
	OperationID: "post-jobs-resume",
	Method:      http.MethodPost,
	Path:        "/jobs/resume",
	Summary:     "Resume Jobs",
	Description: "Resume a specific queue of jobs or all queues.",
	Tags:        []string{"Jobs"},
	Errors:      []int{http.StatusBadRequest},
	Security: []map[string][]string{
		middleware.BearerAuthSecurity("jobs:write"),
	},
}

type ResumeJobsInput struct {
	QueueName string `query:"queue" example:"sync-extension-priority" doc:"The queue name or '*' for all queues"`
}

type ResumeJobsOutput struct {
	Body struct{}
}

func (h Handler) ResumeJobs(ctx context.Context, input *ResumeJobsInput) (*ResumeJobsOutput, error) {
	if err := h.RiverClient.QueueResume(ctx, input.QueueName, nil); err != nil {
		return nil, fmt.Errorf("failed to resume queue '%s': %w", input.QueueName, err)
		// return nil, fmt.Errorf("failed to resume queue '%s': %w", input.QueueName, err)
	}

	return &ResumeJobsOutput{}, nil
}
