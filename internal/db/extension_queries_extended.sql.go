package db

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type SearchExtensionsParams struct {
	Text                 string
	EditorBackground     string
	Language             string
	SortBy               string
	ColorDistance        int
	PublisherName        string
	ExtensionName        string
	ThemeName            string
	ExtensionsPageNumber int
	ExtensionsPageSize   int
	ThemesPageNumber     int
	ThemesPageSize       int
}

type SearchExtensionsRow struct {
	Total                int                            `db:"total"`
	Name                 string                         `db:"name"`
	DisplayName          string                         `db:"display_name"`
	PublisherName        string                         `db:"publisher_name"`
	PublisherDisplayName string                         `db:"publisher_display_name"`
	ShortDescription     pgtype.Text                    `db:"short_description"`
	Themes               []SearchExtensionsThemePartial `db:"themes"`
	TotalThemes          int                            `db:"total_themes"`
	Theme                *SearchExtensionsTheme         `db:"theme"`
}

type SearchExtensionsThemePartial struct {
	Name                       string `json:"name"`
	URL                        string `json:"url"`
	DisplayName                string `json:"display_name"`
	EditorBackground           string `json:"editor_background"`
	ActivityBarBadgeBackground string `json:"activity_bar_badge_background"`
}

type SearchExtensionsTheme struct {
	Name                          string  `json:"name"`
	URL                           string  `json:"url"`
	DisplayName                   string  `json:"display_name"`
	EditorBackground              string  `json:"editor_background"`
	EditorForeground              string  `json:"editor_foreground"`
	ActivityBarBackground         string  `json:"activity_bar_background"`
	ActivityBarForeground         string  `json:"activity_bar_foreground"`
	ActivityBarInActiveForeground string  `json:"activity_bar_in_active_foreground"`
	ActivityBarBorder             *string `json:"activity_bar_border"`
	ActivityBarActiveBorder       string  `json:"activity_bar_active_border"`
	ActivityBarActiveBackground   *string `json:"activity_bar_active_background"`
	ActivityBarBadgeBackground    string  `json:"activity_bar_badge_background"`
	ActivityBarBadgeForeground    string  `json:"activity_bar_badge_foreground"`
	TabsContainerBackground       *string `json:"tabs_container_background"`
	TabsContainerBorder           *string `json:"tabs_container_border"`
	StatusBarBackground           *string `json:"status_bar_background"`
	StatusBarForeground           string  `json:"status_bar_foreground"`
	StatusBarBorder               *string `json:"status_bar_border"`
	TabActiveBackground           *string `json:"tab_active_background"`
	TabInactiveBackground         *string `json:"tab_inactive_background"`
	TabActiveForeground           string  `json:"tab_active_foreground"`
	TabBorder                     string  `json:"tab_border"`
	TabActiveBorder               *string `json:"tab_active_border"`
	TabActiveBorderTop            *string `json:"tab_active_border_top"`
	TitleBarActiveBackground      string  `json:"title_bar_active_background"`
	TitleBarActiveForeground      string  `json:"title_bar_active_foreground"`
	TitleBarBorder                *string `json:"title_bar_border"`
}

func (q *Queries) SearchExtensions(ctx context.Context, arg SearchExtensionsParams) ([]SearchExtensionsRow, error) {
	orderBy := "installs DESC"
	if arg.SortBy == "relevance" {
		orderBy = "text_rank DESC, color_distance ASC, installs DESC"
	} else if arg.SortBy == "trendingDaily" {
		orderBy = "trending_daily DESC"
	} else if arg.SortBy == "trendingWeekly" {
		orderBy = "trending_weekly DESC"
	} else if arg.SortBy == "trendingMonthly" {
		orderBy = "trending_monthly DESC"
	} else if arg.SortBy == "rating" {
		orderBy = "weightedRating DESC"
	} else if arg.SortBy == "updatedAt" {
		orderBy = "updated_at DESC"
	}

	var searchExtensions = fmt.Sprintf(`
	SELECT
		r.total, 
		e.name,
		e.display_name,
		e.short_description,
		e.publisher_name,
		e.publisher_display_name,
		CASE
			WHEN @theme_name = '' THEN COALESCE(max(t2.total), 0)
			ELSE COALESCE(max(t2.total), 0) + 1 END AS total_themes,
		COALESCE(jsonb_agg(to_jsonb(t2.*) - 'extension_id' - 'id' - 'total') FILTER (WHERE t2.id IS NOT NULL), '[]') AS themes,
		jsonb_agg(to_jsonb(t3.*) - 'extension_id' - 'id' - 'tsv' - 'path' - 'created_at' - 'updated_at') -> 0 AS theme
	FROM extensions e
	JOIN (
		WITH results AS (
			SELECT 
				CASE 
					WHEN @text = '' THEN 0 
					ELSE TS_RANK_CD(t.tsv, query, 32) END AS text_rank,
				CASE 
					WHEN @editor_background = '' THEN 0 
					ELSE (@editor_background::cube <-> t.editor_background) END AS color_distance,
				ROW_NUMBER() OVER(
					PARTITION BY t.extension_id 
					ORDER BY
						CASE 
							WHEN @text = '' THEN 0 
							ELSE TS_RANK_CD(t.tsv, query, 32) END DESC,
						CASE
							WHEN @editor_background = '' THEN 0
							ELSE (@editor_background::cube <-> t.editor_background) END ASC,
						t.name ASC
				) AS row_number,
				t.id,
				t.extension_id,
				e.installs,
				e.trending_daily,
				e.trending_weekly,
				e.trending_monthly,
				e.weighted_rating,
				e.updated_at
			FROM themes t
			LEFT JOIN extensions e ON e.id = t.extension_id, WEBSEARCH_TO_TSQUERY(@text) query
			WHERE
				CASE WHEN @publisher_name = '' then true
				ELSE e.publisher_name = @publisher_name END
			AND 
				CASE WHEN @extension_name = '' then true
				ELSE e.name = @extension_name END
			AND 
				CASE 
					WHEN @text = '' THEN true 
					ELSE query @@ t.tsv END
			AND 
				CASE 
					WHEN @editor_background = '' THEN true 
					ELSE @editor_background::cube <-> t.editor_background <= @color_distance END
		)
		SELECT 
			COUNT(*) OVER() total, 
			ROW_NUMBER() OVER(ORDER BY %s) AS row_number, 
			extension_id, 
			color_distance
		FROM results
		WHERE row_number = 1
		OFFSET @extensions_offset
		LIMIT @extensions_limit
	) r ON r.extension_id = e.id
	LEFT JOIN LATERAL (
		SELECT 
			COUNT(*) OVER() total,
			t.extension_id,
			t.id,
			t.name,
			t.display_name,
			t.editor_background,
			t.activity_bar_badge_background,
			i.url
		FROM themes t
		JOIN images i ON i.theme_id = t.id AND i.language = @language AND i.type = 'preview' AND i.format = 'svg'
		WHERE e.id = t.extension_id
		AND
			CASE WHEN @theme_name = '' then true
			ELSE t.name != @theme_name END
		ORDER BY
			CASE
				WHEN @editor_background = '' THEN 0
				ELSE (@editor_background::cube <-> t.editor_background) END ASC,
			t.name ASC
		OFFSET @themes_offset
		LIMIT @themes_limit
	) t2 ON t2.extension_id = e.id
	LEFT JOIN LATERAL (
		SELECT 
			t.extension_id,
			t.name,
			t.display_name,
			t.editor_background,
			t.editor_foreground,
			t.activity_bar_background,
			t.activity_bar_foreground,
			t.activity_bar_in_active_foreground,
			t.activity_bar_border,
			t.activity_bar_active_border,
			t.activity_bar_active_background,
			t.activity_bar_badge_background,
			t.activity_bar_badge_foreground,
			t.tabs_container_background,
			t.tabs_container_border,
			t.status_bar_background,
			t.status_bar_foreground,
			t.status_bar_border,
			t.tab_active_background,
			t.tab_inactive_background,
			t.tab_active_foreground,
			t.tab_border,
			t.tab_active_border,
			t.tab_active_border_top,
			t.title_bar_active_background,
			t.title_bar_active_foreground,
			t.title_bar_border,
			i.url
		FROM themes t
		JOIN images i ON i.theme_id = t.id AND i.language = @language AND i.type = 'preview' AND i.format = 'svg'
		WHERE e.id = t.extension_id
		AND t.name = @theme_name
		OFFSET 0
		LIMIT 1
	) t3 ON t3.extension_id = e.id AND @theme_name != ''
	GROUP BY r.total, r.row_number, e.id
	ORDER BY r.row_number ASC
	`, orderBy)

	// Page numbers start at 1.
	extensionsOffset := (arg.ExtensionsPageNumber - 1) * arg.ExtensionsPageSize
	themesOffset := (arg.ThemesPageNumber - 1) * arg.ThemesPageSize

	rows, err := q.db.Query(ctx, searchExtensions, pgx.NamedArgs{
		"text":              arg.Text,
		"editor_background": arg.EditorBackground,
		"language":          arg.Language,
		"color_distance":    arg.ColorDistance,
		"publisher_name":    arg.PublisherName,
		"extension_name":    arg.ExtensionName,
		"theme_name":        arg.ThemeName,
		"extensions_offset": extensionsOffset,
		"extensions_limit":  arg.ExtensionsPageSize,
		"themes_offset":     themesOffset,
		"themes_limit":      arg.ThemesPageSize,
	})
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items, err := pgx.CollectRows(rows, pgx.RowToStructByName[SearchExtensionsRow])
	if err != nil {
		return nil, err
	}

	return items, nil
}
