package handlers

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
)

var GetHealthOperation = huma.Operation{
	OperationID: "get-health",
	Method:      http.MethodGet,
	Path:        "/health",
	Summary:     "Health Check",
	Tags:        []string{"Misc"},
}

type GetHealthInput struct{}

type GetHealthOutput struct {
	Body struct{}
}

func (h Handler) GetHealth(ctx context.Context, input *GetHealthInput) (*GetHealthOutput, error) {
	_, err := h.DBPool.Exec(ctx, "SELECT 1")
	if err != nil {
		return nil, err
	}

	resp := &GetHealthOutput{}

	return resp, nil
}
