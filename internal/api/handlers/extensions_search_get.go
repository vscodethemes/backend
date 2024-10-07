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
	Text                 string `query:"text" example:"monokai" doc:"The text to search for"`
	EditorBackground     string `query:"editorBackground" example:"#000000" doc:"The editor background color to search for"`
	Language             string `query:"language" default:"js" example:"js" doc:"The language to return themes for"`
	SortBy               string `query:"sortBy" default:"relevance" example:"relevance" doc:"The sort order for results. Set to 'relevance', 'installs', 'trendingDaily', 'trendingWeekly', 'trendingMonthly', 'rating', or 'updatedAt'."`
	ColorDistance        int    `query:"colorDistance" default:"10" example:"100" doc:"The maximum color distance to search for"`
	PublisherName        string `query:"publisherName" example:"sdras" doc:"The publisher name to filter by"`
	ExtensionName        string `query:"extensionName" example:"night-owl" doc:"The extension name to filter by"`
	ThemeName            string `query:"themeName" example:"night-owl" doc:"The theme name to filter by"`
	ExtensionsPageNumber int    `query:"extensionsPageNumber" default:"1" example:"1" doc:"The page number for extensions"`
	ExtensionsPageSize   int    `query:"extensionsPageSize" default:"10" example:"10" doc:"The page size for extensions"`
	ThemesPageNumber     int    `query:"themesPageNumber" default:"1" example:"1" doc:"The page number for themes"`
	ThemesPageSize       int    `query:"themesPageSize" default:"10" example:"10" doc:"The page size for themes"`
}

type SearchExtensionsOutput struct {
	Body struct {
		Total      int         `json:"total"`
		Extensions []Extension `json:"extensions"`
	}
}

type Extension struct {
	Name                 string  `json:"name"`
	DisplayName          string  `json:"displayName"`
	PublisherName        string  `json:"publisherName"`
	PublisherDisplayName string  `json:"publisherDisplayName"`
	ShortDescription     *string `json:"shortDescription"`
	Themes               []Theme `json:"themes"`
	TotalThemes          int     `json:"totalThemes"`
}

type Theme struct {
	Name                       string `json:"name"`
	DisplayName                string `json:"displayName"`
	EditorBackground           string `json:"editorBackground"`
	ActivityBarBadgeBackground string `json:"activityBarBadgeBackground"`
	URL                        string `json:"url"`
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
		Text:                 input.Text,
		Language:             input.Language,
		EditorBackground:     editorBackground,
		SortBy:               input.SortBy,
		ColorDistance:        input.ColorDistance,
		PublisherName:        input.PublisherName,
		ExtensionName:        input.ExtensionName,
		ThemeName:            input.ThemeName,
		ExtensionsPageNumber: input.ExtensionsPageNumber,
		ExtensionsPageSize:   input.ExtensionsPageSize,
		ThemesPageNumber:     input.ThemesPageNumber,
		ThemesPageSize:       input.ThemesPageSize,
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
			TotalThemes:          row.TotalThemes,
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

			activityBarBadgeBackground, err := colors.LabStringToHex(theme.ActivityBarBadgeBackground)
			if err != nil {
				h.logThemeError(ctx, "failed to convert activity_bar_badge_background to hex", row, theme, err)
				continue
			}

			extension.Themes = append(extension.Themes, Theme{
				Name:                       theme.Name,
				DisplayName:                theme.DisplayName,
				EditorBackground:           editorBackground,
				ActivityBarBadgeBackground: activityBarBadgeBackground,
				URL:                        theme.URL,
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
