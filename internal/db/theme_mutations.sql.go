// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0
// source: theme_mutations.sql

package db

import (
	"context"
)

const upsertTheme = `-- name: UpsertTheme :one
insert into "themes" (
  "extension_id",
  "path",
  "name", 
  "display_name",
  "editor_background",
  "editor_foreground",
  "activity_bar_background",
  "activity_bar_foreground",
  "activity_bar_in_active_foreground",
  "activity_bar_border",
  "activity_bar_active_border",
  "activity_bar_active_background",
  "activity_bar_badge_background",
  "activity_bar_badge_foreground",
  "tabs_container_background",
  "tabs_container_border",
  "status_bar_background",
  "status_bar_foreground",
  "status_bar_border",
  "tab_active_background",
  "tab_inactive_background",
  "tab_active_foreground",
  "tab_border",
  "tab_active_border",
  "tab_active_border_top",
  "title_bar_active_background",
  "title_bar_active_foreground",
  "title_bar_border"
)
values (
  $1, 
  $2, 
  $3, 
  $4,
  $5,
  $6,
  $7,
  $8,
  $9,
  $10,
  $11,
  $12,
  $13,
  $14,
  $15,
  $16,
  $17,
  $18,
  $19,
  $20,
  $21,
  $22,
  $23,
  $24,
  $25,
  $26,
  $27,
  $28
)
on conflict("extension_id", "path") do update set
  "name" = excluded."name",
  "display_name" = excluded."display_name",
  "editor_background" = excluded."editor_background",
  "editor_foreground" = excluded."editor_foreground",
  "activity_bar_background" = excluded."activity_bar_background",
  "activity_bar_foreground" = excluded."activity_bar_foreground",
  "activity_bar_in_active_foreground" = excluded."activity_bar_in_active_foreground",
  "activity_bar_border" = excluded."activity_bar_border",
  "activity_bar_active_border" = excluded."activity_bar_active_border",
  "activity_bar_active_background" = excluded."activity_bar_active_background",
  "activity_bar_badge_background" = excluded."activity_bar_badge_background",
  "activity_bar_badge_foreground" = excluded."activity_bar_badge_foreground",
  "tabs_container_background" = excluded."tabs_container_background",
  "tabs_container_border" = excluded."tabs_container_border",
  "status_bar_background" = excluded."status_bar_background",
  "status_bar_foreground" = excluded."status_bar_foreground",
  "status_bar_border" = excluded."status_bar_border",
  "tab_active_background" = excluded."tab_active_background",
  "tab_inactive_background" = excluded."tab_inactive_background",
  "tab_active_foreground" = excluded."tab_active_foreground",
  "tab_border" = excluded."tab_border",
  "tab_active_border" = excluded."tab_active_border",
  "tab_active_border_top" = excluded."tab_active_border_top",
  "title_bar_active_background" = excluded."title_bar_active_background",
  "title_bar_active_foreground" = excluded."title_bar_active_foreground",
  "title_bar_border" = excluded."title_bar_border",
  "updated_at" = now()
returning id, extension_id, path, name, display_name, editor_background, editor_foreground, activity_bar_background, activity_bar_foreground, activity_bar_in_active_foreground, activity_bar_border, activity_bar_active_border, activity_bar_active_background, activity_bar_badge_background, activity_bar_badge_foreground, tabs_container_background, tabs_container_border, status_bar_background, status_bar_foreground, status_bar_border, tab_active_background, tab_inactive_background, tab_active_foreground, tab_border, tab_active_border, tab_active_border_top, title_bar_active_background, title_bar_active_foreground, title_bar_border, created_at, updated_at, tsv
`

type UpsertThemeParams struct {
	ExtensionID                   int64
	Path                          string
	Name                          string
	DisplayName                   string
	EditorBackground              string
	EditorForeground              string
	ActivityBarBackground         string
	ActivityBarForeground         string
	ActivityBarInActiveForeground string
	ActivityBarBorder             *string
	ActivityBarActiveBorder       string
	ActivityBarActiveBackground   *string
	ActivityBarBadgeBackground    string
	ActivityBarBadgeForeground    string
	TabsContainerBackground       *string
	TabsContainerBorder           *string
	StatusBarBackground           *string
	StatusBarForeground           string
	StatusBarBorder               *string
	TabActiveBackground           *string
	TabInactiveBackground         *string
	TabActiveForeground           string
	TabBorder                     string
	TabActiveBorder               *string
	TabActiveBorderTop            *string
	TitleBarActiveBackground      string
	TitleBarActiveForeground      string
	TitleBarBorder                *string
}

func (q *Queries) UpsertTheme(ctx context.Context, arg UpsertThemeParams) (Theme, error) {
	row := q.db.QueryRow(ctx, upsertTheme,
		arg.ExtensionID,
		arg.Path,
		arg.Name,
		arg.DisplayName,
		arg.EditorBackground,
		arg.EditorForeground,
		arg.ActivityBarBackground,
		arg.ActivityBarForeground,
		arg.ActivityBarInActiveForeground,
		arg.ActivityBarBorder,
		arg.ActivityBarActiveBorder,
		arg.ActivityBarActiveBackground,
		arg.ActivityBarBadgeBackground,
		arg.ActivityBarBadgeForeground,
		arg.TabsContainerBackground,
		arg.TabsContainerBorder,
		arg.StatusBarBackground,
		arg.StatusBarForeground,
		arg.StatusBarBorder,
		arg.TabActiveBackground,
		arg.TabInactiveBackground,
		arg.TabActiveForeground,
		arg.TabBorder,
		arg.TabActiveBorder,
		arg.TabActiveBorderTop,
		arg.TitleBarActiveBackground,
		arg.TitleBarActiveForeground,
		arg.TitleBarBorder,
	)
	var i Theme
	err := row.Scan(
		&i.ID,
		&i.ExtensionID,
		&i.Path,
		&i.Name,
		&i.DisplayName,
		&i.EditorBackground,
		&i.EditorForeground,
		&i.ActivityBarBackground,
		&i.ActivityBarForeground,
		&i.ActivityBarInActiveForeground,
		&i.ActivityBarBorder,
		&i.ActivityBarActiveBorder,
		&i.ActivityBarActiveBackground,
		&i.ActivityBarBadgeBackground,
		&i.ActivityBarBadgeForeground,
		&i.TabsContainerBackground,
		&i.TabsContainerBorder,
		&i.StatusBarBackground,
		&i.StatusBarForeground,
		&i.StatusBarBorder,
		&i.TabActiveBackground,
		&i.TabInactiveBackground,
		&i.TabActiveForeground,
		&i.TabBorder,
		&i.TabActiveBorder,
		&i.TabActiveBorderTop,
		&i.TitleBarActiveBackground,
		&i.TitleBarActiveForeground,
		&i.TitleBarBorder,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.Tsv,
	)
	return i, err
}
