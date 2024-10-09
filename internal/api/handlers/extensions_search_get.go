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
	Name                 string         `json:"name"`
	DisplayName          string         `json:"displayName"`
	PublisherName        string         `json:"publisherName"`
	PublisherDisplayName string         `json:"publisherDisplayName"`
	ShortDescription     *string        `json:"shortDescription"`
	Themes               []ThemePartial `json:"themes"`
	TotalThemes          int            `json:"totalThemes"`
	Theme                *Theme         `json:"theme"`
}

type ThemePartial struct {
	Name                       string `json:"name"`
	DisplayName                string `json:"displayName"`
	EditorBackground           string `json:"editorBackground"`
	ActivityBarBadgeBackground string `json:"activityBarBadgeBackground"`
	URL                        string `json:"url"`
}

type Theme struct {
	URL                           string  `json:"url"`
	Name                          string  `json:"name"`
	DisplayName                   string  `json:"displayName"`
	EditorBackground              string  `json:"editorBackground"`
	EditorForeground              string  `json:"editorForeground"`
	ActivityBarBackground         string  `json:"activityBarBackground"`
	ActivityBarForeground         string  `json:"activityBarForeground"`
	ActivityBarInActiveForeground string  `json:"activityBarInActiveForeground"`
	ActivityBarBorder             *string `json:"activityBarBorder"`
	ActivityBarActiveBorder       string  `json:"activityBarActiveBorder"`
	ActivityBarActiveBackground   *string `json:"activityBarActiveBackground"`
	ActivityBarBadgeBackground    string  `json:"activityBarBadgeBackground"`
	ActivityBarBadgeForeground    string  `json:"activityBarBadgeForeground"`
	TabsContainerBackground       *string `json:"tabsContainerBackground"`
	TabsContainerBorder           *string `json:"tabsContainerBorder"`
	StatusBarBackground           *string `json:"statusBarBackground"`
	StatusBarForeground           string  `json:"statusBarForeground"`
	StatusBarBorder               *string `json:"statusBarBorder"`
	TabActiveBackground           *string `json:"tabActiveBackground"`
	TabInactiveBackground         *string `json:"tabInactiveBackground"`
	TabActiveForeground           string  `json:"tabActiveForeground"`
	TabBorder                     string  `json:"tabBorder"`
	TabActiveBorder               *string `json:"tabActiveBorder"`
	TabActiveBorderTop            *string `json:"tabActiveBorderTop"`
	TitleBarActiveBackground      string  `json:"titleBarActiveBackground"`
	TitleBarActiveForeground      string  `json:"titleBarActiveForeground"`
	TitleBarBorder                *string `json:"titleBarBorder"`
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
			Themes:               []ThemePartial{},
		}

		if row.ShortDescription.Valid {
			extension.ShortDescription = &row.ShortDescription.String
		}

		if row.Theme != nil {
			editorBackground, err := colors.LabStringToHex(row.Theme.EditorBackground)
			if err != nil {
				h.logThemeError(ctx, "failed to convert editor_background to hex", row, row.Theme.Name, err)
				continue
			}

			editorForeground, err := colors.LabStringToHex(row.Theme.EditorForeground)
			if err != nil {
				h.logThemeError(ctx, "failed to convert editor_foreground to hex", row, row.Theme.Name, err)
				continue
			}

			activityBarBackground, err := colors.LabStringToHex(row.Theme.ActivityBarBackground)
			if err != nil {
				h.logThemeError(ctx, "failed to convert activity_bar_background to hex", row, row.Theme.Name, err)
				continue
			}

			activityBarForeground, err := colors.LabStringToHex(row.Theme.ActivityBarForeground)
			if err != nil {
				h.logThemeError(ctx, "failed to convert activity_bar_foreground to hex", row, row.Theme.Name, err)
				continue
			}

			activityBarInActiveForeground, err := colors.LabStringToHex(row.Theme.ActivityBarInActiveForeground)
			if err != nil {
				h.logThemeError(ctx, "failed to convert activity_bar_in_active_foreground to hex", row, row.Theme.Name, err)
				continue
			}

			var activityBarBorder *string
			if row.Theme.ActivityBarBorder != nil {
				v, err := colors.LabStringToHex(*row.Theme.ActivityBarBorder)
				if err != nil {
					h.logThemeError(ctx, "failed to convert activity_bar_border to hex", row, row.Theme.Name, err)
					continue
				}
				activityBarBorder = &v
			}

			activityBarActiveBorder, err := colors.LabStringToHex(row.Theme.ActivityBarActiveBorder)
			if err != nil {
				h.logThemeError(ctx, "failed to convert activity_bar_active_border to hex", row, row.Theme.Name, err)
				continue
			}

			var activityBarActiveBackground *string
			if row.Theme.ActivityBarActiveBackground != nil {
				v, err := colors.LabStringToHex(*row.Theme.ActivityBarActiveBackground)
				if err != nil {
					h.logThemeError(ctx, "failed to convert activity_bar_active_background to hex", row, row.Theme.Name, err)
					continue
				}
				activityBarActiveBackground = &v
			}

			activityBarBadgeBackground, err := colors.LabStringToHex(row.Theme.ActivityBarBadgeBackground)
			if err != nil {
				h.logThemeError(ctx, "failed to convert activity_bar_badge_background to hex", row, row.Theme.Name, err)
				continue
			}

			activityBarBadgeForeground, err := colors.LabStringToHex(row.Theme.ActivityBarBadgeForeground)
			if err != nil {
				h.logThemeError(ctx, "failed to convert activity_bar_badge_foreground to hex", row, row.Theme.Name, err)
				continue
			}

			var tabsContainerBackground *string
			if row.Theme.TabsContainerBackground != nil {
				v, err := colors.LabStringToHex(*row.Theme.TabsContainerBackground)
				if err != nil {
					h.logThemeError(ctx, "failed to convert tabs_container_background to hex", row, row.Theme.Name, err)
					continue
				}
				tabsContainerBackground = &v
			}

			var tabsContainerBorder *string
			if row.Theme.TabsContainerBorder != nil {
				v, err := colors.LabStringToHex(*row.Theme.TabsContainerBorder)
				if err != nil {
					h.logThemeError(ctx, "failed to convert tabs_container_border to hex", row, row.Theme.Name, err)
					continue
				}
				tabsContainerBorder = &v
			}

			var statusBarBackground *string
			if row.Theme.StatusBarBackground != nil {
				v, err := colors.LabStringToHex(*row.Theme.StatusBarBackground)
				if err != nil {
					h.logThemeError(ctx, "failed to convert status_bar_background to hex", row, row.Theme.Name, err)
					continue
				}
				statusBarBackground = &v
			}

			statusBarForeground, err := colors.LabStringToHex(row.Theme.StatusBarForeground)
			if err != nil {
				h.logThemeError(ctx, "failed to convert status_bar_foreground to hex", row, row.Theme.Name, err)
				continue
			}

			var statusBarBorder *string
			if row.Theme.StatusBarBorder != nil {
				v, err := colors.LabStringToHex(*row.Theme.StatusBarBorder)
				if err != nil {
					h.logThemeError(ctx, "failed to convert status_bar_border to hex", row, row.Theme.Name, err)
					continue
				}
				statusBarBorder = &v
			}

			var tabActiveBackground *string
			if row.Theme.TabActiveBackground != nil {
				v, err := colors.LabStringToHex(*row.Theme.TabActiveBackground)
				if err != nil {
					h.logThemeError(ctx, "failed to convert status_bar_border to hex", row, row.Theme.Name, err)
					continue
				}
				tabActiveBackground = &v
			}

			var tabInactiveBackground *string
			if row.Theme.TabInactiveBackground != nil {
				v, err := colors.LabStringToHex(*row.Theme.TabInactiveBackground)
				if err != nil {
					h.logThemeError(ctx, "failed to convert tab_inactive_background to hex", row, row.Theme.Name, err)
					continue
				}
				tabInactiveBackground = &v
			}

			tabActiveForeground, err := colors.LabStringToHex(row.Theme.TabActiveForeground)
			if err != nil {
				h.logThemeError(ctx, "failed to convert tab_active_foreground to hex", row, row.Theme.Name, err)
				continue
			}

			tabBorder, err := colors.LabStringToHex(row.Theme.TabBorder)
			if err != nil {
				h.logThemeError(ctx, "failed to convert tab_border to hex", row, row.Theme.Name, err)
				continue
			}

			var tabActiveBorder *string
			if row.Theme.TabActiveBorder != nil {
				v, err := colors.LabStringToHex(*row.Theme.TabActiveBorder)
				if err != nil {
					h.logThemeError(ctx, "failed to convert tab_active_border to hex", row, row.Theme.Name, err)
					continue
				}
				tabActiveBorder = &v
			}

			var tabActiveBorderTop *string
			if row.Theme.TabActiveBorderTop != nil {
				v, err := colors.LabStringToHex(*row.Theme.TabActiveBorderTop)
				if err != nil {
					h.logThemeError(ctx, "failed to convert tab_active_border_top to hex", row, row.Theme.Name, err)
					continue
				}
				tabActiveBorderTop = &v
			}

			titleBarActiveBackground, err := colors.LabStringToHex(row.Theme.TitleBarActiveBackground)
			if err != nil {
				h.logThemeError(ctx, "failed to convert title_bar_active_background to hex", row, row.Theme.Name, err)
				continue
			}

			titleBarActiveForeground, err := colors.LabStringToHex(row.Theme.TitleBarActiveForeground)
			if err != nil {
				h.logThemeError(ctx, "failed to convert title_bar_active_foreground to hex", row, row.Theme.Name, err)
				continue
			}

			var titleBarBorder *string
			if row.Theme.TitleBarBorder != nil {
				v, err := colors.LabStringToHex(*row.Theme.TitleBarBorder)
				if err != nil {
					h.logThemeError(ctx, "failed to convert title_bar_border to hex", row, row.Theme.Name, err)
					continue
				}
				titleBarBorder = &v
			}

			extension.Theme = &Theme{
				URL:                           row.Theme.URL,
				Name:                          row.Theme.Name,
				DisplayName:                   row.Theme.DisplayName,
				EditorBackground:              editorBackground,
				EditorForeground:              editorForeground,
				ActivityBarBackground:         activityBarBackground,
				ActivityBarForeground:         activityBarForeground,
				ActivityBarInActiveForeground: activityBarInActiveForeground,
				ActivityBarBorder:             activityBarBorder,
				ActivityBarActiveBorder:       activityBarActiveBorder,
				ActivityBarActiveBackground:   activityBarActiveBackground,
				ActivityBarBadgeBackground:    activityBarBadgeBackground,
				ActivityBarBadgeForeground:    activityBarBadgeForeground,
				TabsContainerBackground:       tabsContainerBackground,
				TabsContainerBorder:           tabsContainerBorder,
				StatusBarBackground:           statusBarBackground,
				StatusBarForeground:           statusBarForeground,
				StatusBarBorder:               statusBarBorder,
				TabActiveBackground:           tabActiveBackground,
				TabInactiveBackground:         tabInactiveBackground,
				TabActiveForeground:           tabActiveForeground,
				TabBorder:                     tabBorder,
				TabActiveBorder:               tabActiveBorder,
				TabActiveBorderTop:            tabActiveBorderTop,
				TitleBarActiveBackground:      titleBarActiveBackground,
				TitleBarActiveForeground:      titleBarActiveForeground,
				TitleBarBorder:                titleBarBorder,
			}
		}

		for _, theme := range row.Themes {
			editorBackground, err := colors.LabStringToHex(theme.EditorBackground)
			if err != nil {
				h.logThemeError(ctx, "failed to convert editor_background to hex", row, theme.Name, err)
				continue
			}

			activityBarBadgeBackground, err := colors.LabStringToHex(theme.ActivityBarBadgeBackground)
			if err != nil {
				h.logThemeError(ctx, "failed to convert activity_bar_badge_background to hex", row, theme.Name, err)
				continue
			}

			extension.Themes = append(extension.Themes, ThemePartial{
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

func (h Handler) logThemeError(ctx context.Context, msg string, extension db.SearchExtensionsRow, themeName string, err error) {
	h.Logger.LogAttrs(ctx, slog.LevelError, msg,
		slog.String("theme_name", themeName),
		slog.String("extension_name", extension.Name),
		slog.String("publisher_name", extension.PublisherName),
		slog.String("err", err.Error()),
	)
}
