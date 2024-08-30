-- migrate:up

CREATE extension cube;

CREATE TABLE extensions (
  "id" bigserial PRIMARY KEY,
  "vsc_extension_id" text NOT NULL,
  "name" text NOT NULL,
  "display_name" text NOT NULL,
  "short_description" text,
  "publisher_id" text NOT NULL,
  "publisher_name" text NOT NULL,
  "publisher_display_name" text NOT NULL,
  "installs" integer NOT NULL,
  "trending_daily" numeric NOT NULL,
  "trending_weekly" numeric NOT NULL,
  "trending_monthly" numeric NOT NULL,
  "weighted_rating" numeric NOT NULL,
  "published_at" timestamp NOT NULL,
  "released_at" timestamp NOT NULL,
  "created_at" timestamp NOT NULL DEFAULT NOW(),
  "updated_at" timestamp NOT NULL DEFAULT NOW(),
  UNIQUE ("vsc_extension_id"),
  UNIQUE ("name", "publisher_name")
);

CREATE TABLE themes (
  "id" bigserial PRIMARY KEY,
  "extension_id" bigint NOT NULL REFERENCES extensions("id") ON DELETE CASCADE,
  "path" text NOT NULL,
  "name" text NOT NULL,
  "display_name" text NOT NULL,
  "editor_background" cube NOT NULL,
  "editor_foreground" cube NOT NULL,
  "activity_bar_background" cube NOT NULL,
  "activity_bar_foreground" cube NOT NULL,
  "activity_bar_in_active_foreground" cube NOT NULL,
  "activity_bar_border" cube,
  "activity_bar_active_border" cube NOT NULL,
  "activity_bar_active_background" cube,
  "activity_bar_badge_background" cube NOT NULL,
  "activity_bar_badge_foreground" cube NOT NULL,
  "tabs_container_background" cube,
  "tabs_container_border" cube,
  "status_bar_background" cube,
  "status_bar_foreground" cube NOT NULL,
  "status_bar_border" cube,
  "tab_active_background" cube,
  "tab_inactive_background" cube,
  "tab_active_foreground" cube NOT NULL,
  "tab_border" cube NOT NULL,
  "tab_active_border" cube,
  "tab_active_border_top" cube,
  "title_bar_active_background" cube NOT NULL,
  "title_bar_active_foreground" cube NOT NULL,
  "title_bar_border" cube,
  "created_at" timestamp NOT NULL DEFAULT NOW(),
  "updated_at" timestamp NOT NULL DEFAULT NOW(),
   UNIQUE ("extension_id", "path")
);

CREATE TABLE images (
  "id" bigserial PRIMARY KEY,
  "theme_id" bigint NOT NULL REFERENCES themes("id") ON DELETE CASCADE,
  "language" text NOT NULL,
  "type" text NOT NULL,
  "format" text NOT NULL,
  "url" text NOT NULL,
  "created_at" timestamp NOT NULL DEFAULT NOW(),
  "updated_at" timestamp NOT NULL DEFAULT NOW(),
  UNIQUE ("theme_id", "language", "type", "format")
);


-- migrate:down

DROP TABLE images;
DROP TABLE themes;
DROP TABLE extensions;
DROP extension cube;