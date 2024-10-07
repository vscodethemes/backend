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
	Total                int                     `db:"total"`
	Name                 string                  `db:"name"`
	DisplayName          string                  `db:"display_name"`
	PublisherName        string                  `db:"publisher_name"`
	PublisherDisplayName string                  `db:"publisher_display_name"`
	ShortDescription     pgtype.Text             `db:"short_description"`
	Themes               []SearchExtensionsTheme `db:"themes"`
	TotalThemes          int                     `db:"total_themes"`
}

type SearchExtensionsTheme struct {
	Name                       string `json:"name"`
	URL                        string `json:"url"`
	DisplayName                string `json:"display_name"`
	EditorBackground           string `json:"editor_background"`
	ActivityBarBadgeBackground string `json:"activity_bar_badge_background"`
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
		jsonb_agg(to_jsonb(t2.*) - 'extension_id' - 'id' - 'total') AS themes,
		max(t2.total) as total_themes
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
		ORDER BY
			CASE
				WHEN @theme_name = '' THEN true
				ELSE t.name = @theme_name END DESC,
			CASE
				WHEN @editor_background = '' THEN 0
				ELSE (@editor_background::cube <-> t.editor_background) END ASC,
			t.name ASC
		OFFSET @themes_offset
		LIMIT @themes_limit
	) t2 ON t2.extension_id = e.id
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
