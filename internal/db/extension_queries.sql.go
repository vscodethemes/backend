// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0
// source: extension_queries.sql

package db

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
)

const getExtension = `-- name: GetExtension :one
SELECT 
	e.name,
	e.display_name,
	e.publisher_name,
	e.publisher_display_name,
	e.short_description,
	jsonb_agg(json_build_object(
		'name', t.name,
		'display_name', t.display_name,
		'url', i.url
	)) AS themes
FROM extensions e
LEFT JOIN themes t ON t.extension_id = e.id
LEFT JOIN images i ON i.theme_id = t.id
WHERE 
	e.name = $1
	AND e.publisher_name = $2
	AND i.language = $3
GROUP BY e.id
`

type GetExtensionParams struct {
	ExtensionName string
	PublisherName string
	Language      string
}

type GetExtensionRow struct {
	Name                 string
	DisplayName          string
	PublisherName        string
	PublisherDisplayName string
	ShortDescription     pgtype.Text
	Themes               []byte
}

func (q *Queries) GetExtension(ctx context.Context, arg GetExtensionParams) (GetExtensionRow, error) {
	row := q.db.QueryRow(ctx, getExtension, arg.ExtensionName, arg.PublisherName, arg.Language)
	var i GetExtensionRow
	err := row.Scan(
		&i.Name,
		&i.DisplayName,
		&i.PublisherName,
		&i.PublisherDisplayName,
		&i.ShortDescription,
		&i.Themes,
	)
	return i, err
}
