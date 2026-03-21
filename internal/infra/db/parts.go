package db

import (
	"context"

	"github.com/lib/pq"
)

func (e *Engine) UpsertParts(ctx context.Context, sourceID string, idx int, raw []byte) error {
	return write(ctx, e.conn, `
INSERT INTO parts (source_id, idx, raw)
VALUES ($1, $2, $3)
ON CONFLICT (source_id, idx) DO UPDATE
    SET raw        = EXCLUDED.raw,
        created_at = NOW()
`, sourceID, idx, raw)
}

func (e *Engine) SelectParts(ctx context.Context, sourceID string, idxs []int32) ([][]byte, error) {
	return readAll(ctx, e.conn, `
SELECT p.raw
  FROM parts p
 WHERE source_id = $1
   AND idx = ANY($2) 
 ORDER BY idx
`, sourceID, pq.Array(idxs))
}
