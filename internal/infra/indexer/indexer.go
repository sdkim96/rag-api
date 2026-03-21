package indexer

import (
	"fmt"

	"github.com/sdkim96/rag-api/config"
	"github.com/sdkim96/rag-api/internal/infra/blobstore"
	"github.com/sdkim96/rag-api/internal/infra/db"

	"github.com/sdkim96/indexing/runner"
)

type Indexer struct {
	runner *runner.Runner
}

func New(
	cfg *config.Config,
	e *db.Engine,
	bs blobstore.BlobStore,
) (*Indexer, error) {

	provider := NewBlobStoreProvider(e, bs)
	analyzer := NewCUAnalyzer(
		cfg.AzureCU.Endpoint,
		cfg.AzureCU.APIKey,
		bs,
	)
	partWriter := NewDBPartWriter(e)
	enricher := NewOpenAIEnricher(cfg.OpenAI.APIKey)
	searchWriter := NewPGVectorSearchWriter(e)
	cache := NewDBCache(e)

	r, err := runner.New(
		runner.WithProvider(provider),
		runner.WithAnalyzer(analyzer),
		runner.WithPartWriter(partWriter),
		runner.WithEnricher(enricher),
		runner.WithSearchWriter(searchWriter),
		runner.WithCache(cache),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create runner: %w", err)
	}

	return &Indexer{runner: r}, nil
}

func (i *Indexer) Runner() *runner.Runner {
	return i.runner
}
