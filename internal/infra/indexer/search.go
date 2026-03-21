package indexer

import (
	"context"
	"fmt"

	"github.com/sdkim96/indexing/search"
	"github.com/sdkim96/rag-api/internal/infra/db"
)

type PGVectorSearchWriter struct {
	engine *db.Engine
}

func NewPGVectorSearchWriter(engine *db.Engine) *PGVectorSearchWriter {
	return &PGVectorSearchWriter{engine: engine}
}

var _ search.SearchWriter = (*PGVectorSearchWriter)(nil)

func (w *PGVectorSearchWriter) Write(ctx context.Context, sourceID string, docs []search.SearchDoc) error {

	for chunkIdx, doc := range docs {
		fields := doc.Fields()

		title, ok := fields["topic"].(string)
		if !ok {
			return fmt.Errorf("missing or invalid title field")
		}
		summary, ok := fields["summary"].(string)
		if !ok {
			return fmt.Errorf("missing or invalid summary field")
		}
		idx, ok := fields["idxs"].([]int)
		if !ok {
			return fmt.Errorf("missing or invalid idxs field")
		}
		keywords, ok := fields["keywords"].([]string)
		if !ok {
			return fmt.Errorf("missing or invalid keywords field")
		}

		embedding, ok := fields["embedding"].([]float64)

		if err := w.engine.InsertSearch(
			ctx,
			sourceID,
			chunkIdx,
			idx,
			ToFloat32(embedding),
			title,
			summary,
			keywords,
		); err != nil {
			return err
		}
	}
	return nil
}

func ToFloat32(f64 []float64) []float32 {
	f32 := make([]float32, len(f64))
	for i, v := range f64 {
		f32[i] = float32(v)
	}
	return f32
}
