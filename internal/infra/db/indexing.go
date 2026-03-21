package db

import "context"

func (e *Engine) InsertIndexing(ctx context.Context, sourceID, status string, errMsg *string) error {
	return write(ctx, e.conn, `
INSERT INTO indexing (source_id, status, error)
VALUES ($1, $2, $3)
    `, sourceID, status, errMsg)
}
