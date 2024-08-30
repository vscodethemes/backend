-- name: UpsertExtension :one
insert into "extensions" (
  "vsc_extension_id", 
  "name", 
  "display_name", 
  "short_description", 
  "publisher_id",
  "publisher_name",
  "publisher_display_name",
  "installs", 
  "trending_daily",
  "trending_weekly",
  "trending_monthly",
  "weighted_rating",
  "published_at",
  "released_at"
)
values (
  @vsc_extension_id, 
  @name, 
  @display_name, 
  @short_description, 
  @publisher_id,
  @publisher_name,
  @publisher_display_name,
  @installs, 
  @trending_daily,
  @trending_weekly,
  @trending_monthly,
  @weighted_rating,
  @published_at,
  @released_at
)
on conflict("vsc_extension_id") do update set
  "name" = excluded."name",
  "display_name" = excluded."display_name",
  "short_description" = excluded."short_description",
  "publisher_id" = excluded."publisher_id",
  "publisher_name" = excluded."publisher_name",
  "publisher_display_name" = excluded."publisher_display_name",
  "installs" = excluded."installs",
  "trending_daily" = excluded."trending_daily",
  "trending_weekly" = excluded."trending_weekly",
  "trending_monthly" = excluded."trending_monthly",
  "weighted_rating" = excluded."weighted_rating",
  "published_at" = excluded."published_at",
  "released_at" = excluded."released_at",
  "updated_at" = now()
returning *;