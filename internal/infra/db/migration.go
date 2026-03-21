package db

const System = "system"

const ddlSchema = `
CREATE EXTENSION IF NOT EXISTS vector;
CREATE EXTENSION IF NOT EXISTS pg_trgm;

-- 1. sources
CREATE TABLE IF NOT EXISTS sources (
    id          VARCHAR(255)  PRIMARY KEY,
    owner_id    VARCHAR(255)  NOT NULL,
    uri         VARCHAR(1024) NOT NULL,
    mime_type   VARCHAR(255)  NOT NULL,
    name        VARCHAR(1024) NULL,
    size        BIGINT        NULL,
    origin      JSONB         NOT NULL,
    created_at  TIMESTAMPTZ   DEFAULT NOW(),
    deleted_at  TIMESTAMPTZ   DEFAULT NULL
);

CREATE INDEX IF NOT EXISTS idx_sources_owner_id
    ON sources(owner_id);
CREATE INDEX IF NOT EXISTS idx_sources_uri
    ON sources(uri);
CREATE INDEX IF NOT EXISTS idx_sources_mime_type
    ON sources(mime_type);
CREATE INDEX IF NOT EXISTS idx_sources_active
    ON sources(owner_id, created_at DESC)
    WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_sources_name
    ON sources USING gin(name gin_trgm_ops)
    WHERE name IS NOT NULL;


-- 2. indexing
CREATE TABLE IF NOT EXISTS indexing (
    id          BIGSERIAL     PRIMARY KEY,
    source_id   VARCHAR(255)  NOT NULL REFERENCES sources(id),
    status      VARCHAR(50)   NOT NULL,
    error       TEXT          NULL,
    created_at  TIMESTAMPTZ   DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_indexing_source_id_created_at
    ON indexing(source_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_indexing_status
    ON indexing(status);


-- 3. parts
CREATE TABLE IF NOT EXISTS parts (
    source_id   VARCHAR(255)  NOT NULL REFERENCES sources(id),
    idx         INT           NOT NULL,
    raw         JSONB         NOT NULL,
    created_at  TIMESTAMPTZ   DEFAULT NOW(),

    PRIMARY KEY (source_id, idx)
);


-- 4. search
CREATE TABLE IF NOT EXISTS search (
    id          BIGSERIAL     PRIMARY KEY,
    source_id   VARCHAR(255)  NOT NULL REFERENCES sources(id),
    chunk_idx   INT           NOT NULL,
    part_idxs   INT[]         NOT NULL,
    vector      VECTOR(1536)  NOT NULL,
    topic       TEXT          NOT NULL,
    summary     TEXT          NOT NULL,
    keywords    TEXT[]        NOT NULL,
    created_at  TIMESTAMPTZ   DEFAULT NOW(),

    UNIQUE (source_id, chunk_idx)
);

CREATE INDEX IF NOT EXISTS idx_search_source_id_chunk_idx
    ON search(source_id, chunk_idx);
CREATE INDEX IF NOT EXISTS idx_search_vector
    ON search USING ivfflat(vector vector_cosine_ops);
CREATE INDEX IF NOT EXISTS idx_search_keywords
    ON search USING gin(keywords);
CREATE INDEX IF NOT EXISTS idx_search_topic
    ON search USING gin(topic gin_trgm_ops);
CREATE INDEX IF NOT EXISTS idx_search_summary
    ON search USING gin(summary gin_trgm_ops);


-- 5. cache
CREATE TABLE IF NOT EXISTS cache (
    key         VARCHAR(255)  PRIMARY KEY,
    value       BYTEA         NOT NULL,
    expires_at  TIMESTAMPTZ   NULL,
    created_at  TIMESTAMPTZ   DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_cache_expires_at
    ON cache(expires_at)
    WHERE expires_at IS NOT NULL;
`
