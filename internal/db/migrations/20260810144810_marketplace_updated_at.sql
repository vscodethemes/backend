-- migrate:up

ALTER TABLE "extensions" ADD COLUMN "marketplace_updated_at" timestamp without time zone;

-- Seed the column so we don't trigger a full sync of all extensions.
UPDATE "extensions" SET "marketplace_updated_at" = "updated_at";

-- migrate:down

ALTER TABLE "extensions" DROP COLUMN "marketplace_updated_at";
