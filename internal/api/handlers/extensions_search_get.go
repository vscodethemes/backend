package handlers

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/vscodethemes/backend/internal/api/middleware"
	"github.com/vscodethemes/backend/internal/colors"
	"github.com/vscodethemes/backend/internal/db"
)

var SearchExtensionsOperation = huma.Operation{
	OperationID: "search-extensions",
	Method:      http.MethodGet,
	Path:        "/extensions/search",
	Summary:     "Search Extensions",
	Description: "Search extensions by text or color.",
	Tags:        []string{"Extensions"},
	Errors:      []int{http.StatusBadRequest, http.StatusNotFound},
	Security: []map[string][]string{
		middleware.BearerAuthSecurity("extension:read"),
	},
}

type SearchExtensionsInput struct {
	Text             string `query:"text" example:"monokai" doc:"The text to search for"`
	EditorBackground string `query:"editorBackground" example:"#000000" doc:"The editor background color to search for"`
	Language         string `query:"language" default:"js" example:"js" doc:"The language to return themes for"`
	PageNumber       int    `query:"pageNumber" default:"1" example:"1" doc:"The page number to return"`
	PageSize         int    `query:"pageSize" default:"36" example:"10" doc:"The number of results to return"`
	ColorDistance    int    `query:"colorDistance" default:"10" example:"100" doc:"The maximum color distance to search for"`
	SortBy           string `query:"sortBy" default:"relevance" example:"relevance" doc:"The sort order for results. Set to 'relevance', 'installs', 'trendingDaily', 'trendingWeekly', 'trendingMonthly', 'rating', or 'updatedAt'."`
}

type SearchExtensionsOutput struct {
	Body struct {
		Total      int         `json:"total"`
		Extensions []Extension `json:"extensions"`
	}
}

func (h Handler) SearchExtensions(ctx context.Context, input *SearchExtensionsInput) (*SearchExtensionsOutput, error) {
	queries := db.New(h.DBPool)

	var editorBackground string
	var err error
	if input.EditorBackground != "" {
		editorBackground, err = colors.HexToLabString(input.EditorBackground)
	}
	if err != nil {
		return nil, huma.Error400BadRequest("Invalid editor_background")
	}

	result, err := queries.SearchExtensions(ctx, db.SearchExtensionsParams{
		Text:             input.Text,
		EditorBackground: editorBackground,
		Language:         input.Language,
		PageNumber:       input.PageNumber,
		PageSize:         input.PageSize,
		ColorDistance:    input.ColorDistance,
		SortBy:           input.SortBy,
	})

	if err != nil {
		return nil, fmt.Errorf("failed to search extensions: %w", err)
	}

	resp := &SearchExtensionsOutput{}
	resp.Body.Extensions = make([]Extension, len(result))

	for index, row := range result {
		if index == 0 {
			resp.Body.Total = row.Total
		}

		extension := Extension{
			Name:                 row.Name,
			DisplayName:          row.DisplayName,
			PublisherName:        row.PublisherName,
			PublisherDisplayName: row.PublisherDisplayName,
		}

		if row.ShortDescription.Valid {
			extension.ShortDescription = &row.ShortDescription.String
		}

		for _, theme := range row.Themes {
			editorBackground, err := colors.LabStringToHex(theme.EditorBackground)
			if err != nil {
				h.logThemeError(ctx, "failed to convert editor_background to hex", row, theme, err)
				continue
			}

			extension.Themes = append(extension.Themes, Theme{
				Name:             theme.Name,
				DisplayName:      theme.DisplayName,
				EditorBackground: editorBackground,
				URL:              theme.URL,
			})
		}

		resp.Body.Extensions[index] = extension
	}

	return resp, nil
}

func (h Handler) logThemeError(ctx context.Context, msg string, extension db.SearchExtensionsRow, theme db.SearchExtensionsTheme, err error) {
	h.Logger.LogAttrs(ctx, slog.LevelError, msg,
		slog.String("theme_name", theme.Name),
		slog.String("extension_name", extension.Name),
		slog.String("publisher_name", extension.PublisherName),
		slog.String("err", err.Error()),
	)
}
