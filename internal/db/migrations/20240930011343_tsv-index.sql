-- migrate:up

CREATE INDEX "themes_tsv_idx" ON "themes" USING gist ("tsv" tsvector_ops);

-- migrate:down

DROP INDEX "themes_tsv_idx";