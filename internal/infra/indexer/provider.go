package indexer

import (
	"context"
	"encoding/json"
	"io"

	"github.com/sdkim96/indexing/input"
	"github.com/sdkim96/indexing/mime"
	"github.com/sdkim96/rag-api/internal/infra/blobstore"
	"github.com/sdkim96/rag-api/internal/infra/db"
	"github.com/sdkim96/rag-api/internal/source"
)

type BlobStoreProvider struct {
	db *db.Engine
	bs blobstore.BlobStore
}

func NewBlobStoreProvider(db *db.Engine, bs blobstore.BlobStore) *BlobStoreProvider {
	return &BlobStoreProvider{db: db, bs: bs}
}

type BlobInput struct {
	io.ReadCloser
	b blobstore.Blob
}

func (inp *BlobInput) MimeType() mime.Type {
	return mime.Type(inp.b.MimeType())
}
func (inp *BlobInput) Meta() map[string]any {

	return map[string]any{
		"uri":  inp.b.URI(),
		"size": inp.b.Size(),
	}
}
func (inp *BlobInput) Read(p []byte) (int, error) {
	return inp.ReadCloser.Read(p)

}
func (inp *BlobInput) Close() error {
	return inp.ReadCloser.Close()
}

func (p *BlobStoreProvider) Provide(ctx context.Context, sourcID string) (input.Input, error) {

	srcByte, err := p.db.SelectSource(ctx, sourcID)
	if err != nil {
		return nil, err
	}

	var src source.Source
	err = json.Unmarshal(srcByte, &src)
	if err != nil {
		return nil, err
	}

	b, err := p.bs.Get(ctx, src.ID)
	if err != nil {
		return nil, err
	}
	rc, err := p.bs.Download(ctx, src.ID)
	if err != nil {
		return nil, err
	}
	return &BlobInput{ReadCloser: rc, b: b}, nil
}

var _ input.Provider = (*BlobStoreProvider)(nil)
var _ input.Input = (*BlobInput)(nil)
