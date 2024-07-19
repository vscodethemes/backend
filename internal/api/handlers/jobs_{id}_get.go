package handlers

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/riverqueue/river/rivertype"
)

var GetJobByIDOperation = huma.Operation{
	OperationID: "get-job-by-id",
	Method:      http.MethodGet,
	Path:        "/jobs/{id}",
	Summary:     "Get Job",
	Description: "Get the state of a job by it's ID.",
	Tags:        []string{"Jobs"},
	Errors:      []int{http.StatusNotFound},
}

type GetJobInput struct {
	ID int64 `path:"id" example:"1" doc:"The ID of the job"`
}

type GetJobOutput struct {
	Body struct {
		Job Job `json:"job"`
	}
}

type Job struct {
	Id        int64     `json:"id"`
	State     string    `json:"state"`
	CreatedAt time.Time `json:"created_at"`
}

func (h Handler) GetJobByID(ctx context.Context, input *GetJobInput) (*GetJobOutput, error) {
	job, err := h.RiverClient.JobGet(ctx, input.ID)
	if err != nil {
		if errors.Is(err, rivertype.ErrNotFound) {
			return nil, huma.NewError(http.StatusNotFound, "Job not found")
		} else {
			return nil, fmt.Errorf("failed to get job: %w", err)
		}
	}

	resp := &GetJobOutput{}
	resp.Body.Job.Id = job.ID
	resp.Body.Job.State = string(job.State)
	resp.Body.Job.CreatedAt = job.CreatedAt

	return resp, nil
}
