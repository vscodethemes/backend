// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0
// source: theme_queries.sql

package db

import (
	"context"
)

const getColorCounts = `-- name: GetColorCounts :many

SELECT
	t.editor_background as color,
	count(*) as count
FROM themes t
GROUP BY color
ORDER BY count DESC
`

type GetColorCountsRow struct {
	Color string
	Count int64
}

func (q *Queries) GetColorCounts(ctx context.Context) ([]GetColorCountsRow, error) {
	rows, err := q.db.Query(ctx, getColorCounts)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GetColorCountsRow
	for rows.Next() {
		var i GetColorCountsRow
		if err := rows.Scan(&i.Color, &i.Count); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}