package search

import (
	"context"
	"encoding/json"

	"github.com/lib/pq"
	"github.com/openai/openai-go"
	"github.com/sdkim96/rag-api/internal/infra/db"
	"github.com/sdkim96/rag-api/internal/infra/indexer"
)

func SelectSearchableSourceIDs(ctx context.Context, e *db.Engine, ownerID string) ([]string, error) {
	return e.SelectSearchableSourceIDs(ctx, ownerID)
}

func Embed(ctx context.Context, oaiClient openai.Client, query string) ([]float32, error) {
	resp, err := oaiClient.Embeddings.New(ctx, openai.EmbeddingNewParams{
		Input: openai.EmbeddingNewParamsInputUnion{
			OfString: openai.String(query),
		},
		Model: "text-embedding-3-small",
	})
	if err != nil {
		return nil, err
	}
	return indexer.ToFloat32(resp.Data[0].Embedding), nil
}

func HybridSearch(
	ctx context.Context,
	e *db.Engine,
	embedding []float32,
	keywordQuery string,
	sourceIDs []string,
	limit int,
) ([]SearchDoc, error) {
	rows, err := e.HybridSearch(ctx, embedding, keywordQuery, sourceIDs, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	docs := make([]SearchDoc, 0, limit)
	for rows.Next() {
		var d SearchDoc
		if err := rows.Scan(
			&d.ID,
			&d.SourceID,
			pq.Array(&d.PartIdxs), // ← 추가
			&d.Topic,
			&d.Summary,
			pq.Array(&d.Keywords), // ← 추가
			&d.Score,
		); err != nil {
			return nil, err
		}
		docs = append(docs, d)
	}
	return docs, rows.Err()
}

func SelectParts(ctx context.Context, e *db.Engine, sourceID string, idx []int32) ([]CUPart, error) {
	partBytes, err := e.SelectParts(ctx, sourceID, idx)
	if err != nil {
		return nil, err
	}

	var cuParts []CUPart
	for _, p := range partBytes {
		var cuPart CUPart
		if err := json.Unmarshal(p, &cuPart); err != nil {
			continue
		}
		cuParts = append(cuParts, cuPart)
	}
	return cuParts, nil
}
