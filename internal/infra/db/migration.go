package db

const ddlSchema = `
CREATE TABLE IF NOT EXISTS store (
    id         SERIAL PRIMARY KEY,
    namespace  VARCHAR(255) NOT NULL,
    key        VARCHAR(255) NOT NULL,
    value      JSONB        NOT NULL,
    metadata   JSONB        NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ  DEFAULT NOW(),
    deleted_at TIMESTAMPTZ  DEFAULT NULL,
    UNIQUE (namespace, key)
);
CREATE INDEX IF NOT EXISTS idx_store_namespace ON store (namespace);
CREATE INDEX IF NOT EXISTS idx_store_metadata ON store USING GIN (metadata);
`
