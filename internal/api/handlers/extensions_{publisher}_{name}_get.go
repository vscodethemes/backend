package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/jackc/pgx/v5"
	"github.com/vscodethemes/backend/internal/api/middleware"
	"github.com/vscodethemes/backend/internal/db"
)

var GetExtensionOperation = huma.Operation{
	OperationID: "get-extension",
	Method:      http.MethodGet,
	Path:        "/extensions/{publisher}/{name}",
	Summary:     "Get Extension",
	Description: "Get an extension and it's themes.",
	Tags:        []string{"Extensions"},
	Errors:      []int{http.StatusBadRequest, http.StatusNotFound},
	Security: []map[string][]string{
		middleware.BearerAuthSecurity("extension:read"),
	},
}

type GetExtensionInput struct {
	PublisherName string `path:"publisher" example:"sdras" doc:"The publisher name"`
	ExtensionName string `path:"name" example:"night-owl" doc:"The extension name"`
	Language      string `query:"language" default:"js" example:"js" doc:"The language to return themes for"`
}

type GetExtensionOutput struct {
	Body struct {
		Extension Extension `json:"extension"`
	}
}

type Extension struct {
	Name                 string  `json:"name"`
	DisplayName          string  `json:"displayName"`
	PublisherName        string  `json:"publisherName"`
	PublisherDisplayName string  `json:"publisherDisplayName"`
	ShortDescription     *string `json:"shortDescription"`
	Themes               []Theme `json:"themes"`
}

type Theme struct {
	Name             string `json:"name"`
	DisplayName      string `json:"displayName"`
	EditorBackground string `json:"editorBackground"`
	URL              string `json:"url"`
}

func (h Handler) GetExtension(ctx context.Context, input *GetExtensionInput) (*GetExtensionOutput, error) {
	queries := db.New(h.DBPool)

	result, err := queries.GetExtension(ctx, db.GetExtensionParams{
		PublisherName: input.PublisherName,
		ExtensionName: input.ExtensionName,
		Language:      input.Language,
	})

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, huma.Error404NotFound("Extension not found")
		}
		return nil, fmt.Errorf("failed to get extension: %w", err)
	}

	extension := &Extension{
		Name:                 result.Name,
		DisplayName:          result.DisplayName,
		PublisherName:        result.PublisherName,
		PublisherDisplayName: result.PublisherDisplayName,
	}

	if result.ShortDescription.Valid {
		extension.ShortDescription = &result.ShortDescription.String
	}

	err = json.Unmarshal(result.Themes, &extension.Themes)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal themes: %w", err)
	}

	resp := &GetExtensionOutput{}
	resp.Body.Extension = *extension

	return resp, nil
}
