package db

import (
	"context"
	"database/sql"
	"strings"

	"github.com/lib/pq"
	"github.com/pgvector/pgvector-go"
)

func (e *Engine) InsertSearch(
	ctx context.Context,
	sourceID string,
	chunkIdx int,
	partIdxs []int,
	vector []float32,
	topic string,
	summary string,
	keywords []string,
) error {
	return write(ctx, e.conn, `
INSERT INTO search (source_id, chunk_idx, part_idxs, vector, topic, summary, keywords)
VALUES ($1, $2, $3, $4, $5, $6, $7)
ON CONFLICT (source_id, chunk_idx) DO UPDATE
   SET part_idxs  = EXCLUDED.part_idxs,
       vector     = EXCLUDED.vector,
       topic      = EXCLUDED.topic,
       summary    = EXCLUDED.summary,
       keywords   = EXCLUDED.keywords,
       created_at = NOW()
    `,
		sourceID,
		chunkIdx,
		partIdxs,
		pgvector.NewVector(vector),
		topic,
		summary,
		keywords,
	)
}

func (e *Engine) HybridSearch(
	ctx context.Context,
	vector []float32,
	keywordQuery string,
	sourceIDs []string,
	limit int,
) (*sql.Rows, error) {
	keywords := strings.Fields(keywordQuery)

	return search(ctx, e.conn, `
WITH vector_search AS (
    SELECT id, source_id, part_idxs, topic, summary, keywords,
           ROW_NUMBER() OVER (ORDER BY vector <-> $1) AS rank
      FROM "search"
     WHERE source_id = ANY($2)
     LIMIT 20
),
keyword_search AS (
    SELECT id, source_id, part_idxs, topic, summary, keywords,
           ROW_NUMBER() OVER (ORDER BY id) AS rank  -- ORDER BY 1 → id
      FROM "search"
     WHERE source_id = ANY($2)
       AND (
           keywords  && $3::text[]
           OR topic   ILIKE '%' || $4 || '%'
           OR summary ILIKE '%' || $4 || '%'
       )
     LIMIT 20
)
SELECT
    COALESCE(v.id,        k.id)        AS id,
    COALESCE(v.source_id, k.source_id) AS source_id,
    COALESCE(v.part_idxs, k.part_idxs) AS part_idxs,
    COALESCE(v.topic,     k.topic)     AS topic,
    COALESCE(v.summary,   k.summary)   AS summary,
    COALESCE(v.keywords,  k.keywords)  AS keywords,
    COALESCE(1.0 / (60 + v.rank), 0) +
    COALESCE(1.0 / (60 + k.rank), 0)  AS score
  FROM vector_search v
  FULL OUTER JOIN keyword_search k ON v.id = k.id
 ORDER BY score DESC
 LIMIT $5
`,
		pgvector.NewVector(vector),
		pq.Array(sourceIDs), // $2
		pq.Array(keywords),  // $3
		keywordQuery,        // $4
		limit,               // $5
	)
}

// 벡터 검색
const vectorSearchQuery = `
SELECT id, source_id, part_idxs, topic, summary, keywords,
       vector <-> $1 AS score
  FROM search
 WHERE source_id = ANY($2)
 ORDER BY score
 LIMIT $3
`

// 키워드 검색
const keywordSearchQuery = `
SELECT id, source_id, part_idxs, topic, summary, keywords
  FROM search
 WHERE source_id = ANY($1)
   AND keywords && $2
`

// 전문 검색
const fullTextSearchQuery = `
SELECT id, source_id, part_idxs, topic, summary, keywords
  FROM search
 WHERE source_id = ANY($1)
   AND to_tsvector('english', topic || ' ' || summary) @@ to_tsquery('english', $2)
`
