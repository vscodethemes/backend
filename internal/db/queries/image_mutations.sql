-- name: UpsertImage :one
insert into "images" (
  "theme_id",
  "language", 
  "type",
  "format",
  "url"
)
values (
  @theme_id, 
  @language, 
  @type, 
  @format,
  @url
)
on conflict("theme_id", "language", "type",  "format") do update set
  "url" = excluded."url",
  "updated_at" = now()
returning *;