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
	ID int64 `path:"id" example:"0" doc:"The ID of the job"`
}

type GetJobOutput struct {
	Body struct {
		Job Job `json:"job"`
	}
}

type Job struct {
	ID          int64             `json:"id"`
	Attempt     int               `json:"attempt"`
	AttemptedAt *time.Time        `json:"attemptedAt"`
	CreatedAt   time.Time         `json:"createdAt"`
	Errors      []JobAttemptError `json:"errors"`
	FinalizedAt *time.Time        `json:"finalizedAt"`
	MaxAttempts int               `json:"maxAttempts"`
	State       string            `json:"state"`
}

type JobAttemptError struct {
	At      time.Time `json:"at"`
	Attempt int       `json:"attempt"`
	Error   string    `json:"error"`
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

	if job == nil {
		return nil, huma.NewError(http.StatusNotFound, "Job not found")
	}

	resp := &GetJobOutput{}
	resp.Body.Job = mapRiverJobToJob(*job)

	return resp, nil
}

func mapRiverJobToJob(riverJob rivertype.JobRow) Job {
	job := Job{
		ID:          riverJob.ID,
		Attempt:     riverJob.Attempt,
		AttemptedAt: riverJob.AttemptedAt,
		CreatedAt:   riverJob.CreatedAt,
		FinalizedAt: riverJob.FinalizedAt,
		MaxAttempts: riverJob.MaxAttempts,
		State:       string(riverJob.State),
	}

	for _, riverError := range riverJob.Errors {
		job.Errors = append(job.Errors, JobAttemptError{
			At:      riverError.At,
			Attempt: riverError.Attempt,
			Error:   riverError.Error,
		})
	}

	return job
}
