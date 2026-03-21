package indexing

import (
	"context"
	"log"

	"github.com/sdkim96/rag-api/internal/infra/db"
	"github.com/sdkim96/rag-api/internal/infra/indexer"
	"github.com/sdkim96/rag-api/internal/source"
)

func DoIndexing(
	ctx context.Context,
	e *db.Engine,
	i *indexer.Indexer,
	key string,
) {
	e.InsertIndexing(ctx, key, string(source.StatusInProgress), nil)
	for event, err := range i.Runner().Run(ctx, key) {
		if err != nil {
			log.Printf("failed at %s: %v", event.Stage, err)
			errMsg := err.Error()
			e.InsertIndexing(ctx, key, string(source.StatusFailed), &errMsg)
			return
		}
		log.Printf("[%s] %s", event.Stage, event.Duration)
	}
	e.InsertIndexing(ctx, key, string(source.StatusCompleted), nil)
}
