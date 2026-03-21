package indexer

import (
	"context"
	"encoding/json"

	"github.com/sdkim96/indexing/part"
	"github.com/sdkim96/rag-api/internal/infra/db"
)

type DBPartWriter struct {
	db *db.Engine
}

func NewDBPartWriter(db *db.Engine) *DBPartWriter {
	return &DBPartWriter{db: db}
}

func (w *DBPartWriter) Write(ctx context.Context, sourceID string, parts []part.Part) error {

	for idx, p := range parts {
		b, err := json.Marshal(p)
		if err != nil {
			return err
		}

		if err := w.db.UpsertParts(ctx, sourceID, idx, b); err != nil {
			return err
		}
	}
	return nil
}

var _ part.PartWriter = (*DBPartWriter)(nil)
