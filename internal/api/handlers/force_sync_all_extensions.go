package handlers

import (
	"context"
	"fmt"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/riverqueue/river"
	"github.com/vscodethemes/backend/internal/api/middleware"
	"github.com/vscodethemes/backend/internal/db"
	"github.com/vscodethemes/backend/internal/workers"
)

var ForceSyncAllExtensionsOperation = huma.Operation{
	OperationID: "post-extensions-force-sync",
	Method:      http.MethodPost,
	Path:        "/extensions/force-sync",
	Summary:     "Force Sync Extensions",
	Description: "Force sync all existing extensions.",
	Tags:        []string{"Extensions"},
	Errors:      []int{http.StatusBadRequest},
	Security: []map[string][]string{
		middleware.BearerAuthSecurity("extension:write"),
	},
}

type ForceSyncAllExtensionsInput struct{}

type ForceSyncAllExtensionsOutput struct {
	Body struct {
		ExtensionsToSync int `json:"extensionsToSync"`
	}
}

func (h Handler) ForceSyncAllExtensions(ctx context.Context, input *ForceSyncAllExtensionsInput) (*ForceSyncAllExtensionsOutput, error) {

	// Get all extensions from the database.
	queries := db.New(h.DBPool)
	extensions, err := queries.GetAllExtensionsForUpdate(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get all extensions: %w", err)
	}

	batch := []river.InsertManyParams{}
	for _, extension := range extensions {
		batch = append(batch, river.InsertManyParams{
			Args: workers.SyncExtensionArgs{
				PublisherName: extension.PublisherName,
				ExtensionName: extension.Name,
				Force:         true,
			},
			InsertOpts: &river.InsertOpts{
				Queue: workers.SyncExtensionLowPriorityQueue,
			},
		})
	}

	if len(batch) > 0 {
		if _, err = h.RiverClient.InsertMany(ctx, batch); err != nil {
			return nil, fmt.Errorf("failed to insert job: %w", err)
		}
	}

	resp := &ForceSyncAllExtensionsOutput{}
	resp.Body.ExtensionsToSync = len(extensions)

	return resp, nil
}
