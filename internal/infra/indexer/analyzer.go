package indexer

import (
	"bytes"
	"context"
	"log"
	"net/http"

	"github.com/sdkim96/indexing/analyze/cu"
	"github.com/sdkim96/indexing/mime"
	"github.com/sdkim96/indexing/urio"
	"github.com/sdkim96/rag-api/internal/infra/blobstore"
)

type CUAnalyzer struct {
	*cu.CU
}

type AzureBlobStoreFigWriter struct {
	bs       blobstore.BlobStore
	path     string
	mimeType mime.Type
	buf      bytes.Buffer
	uri      urio.URI
}

func NewAzureBlobStoreFigWriter(bs blobstore.BlobStore, path string, mimeType mime.Type) *AzureBlobStoreFigWriter {
	return &AzureBlobStoreFigWriter{bs: bs, path: path, mimeType: mimeType}
}

var _ urio.WriteCloser = (*AzureBlobStoreFigWriter)(nil)

func (w *AzureBlobStoreFigWriter) Write(p []byte) (n int, err error) {
	return w.buf.Write(p)
}
func (w *AzureBlobStoreFigWriter) Close() error {
	if err := w.bs.Upload(context.Background(), w.path, string(w.mimeType), &w.buf); err != nil {
		return err
	}
	blob, err := w.bs.Get(context.Background(), w.path)
	if err != nil {
		return err
	}
	w.uri = urio.URI(blob.URI())
	return nil
}
func (w *AzureBlobStoreFigWriter) URI() urio.URI {
	return w.uri
}

func NewCUAnalyzer(endpoint, apiKey string, bs blobstore.BlobStore) *CUAnalyzer {
	return &CUAnalyzer{
		CU: cu.New(
			cu.NewClient(endpoint, apiKey, http.DefaultClient),
			func(ctx context.Context, name string, mimeType mime.Type) (urio.WriteCloser, error) {
				return NewAzureBlobStoreFigWriter(bs, name, mimeType), nil
			},
			cu.WithPollCallback(func(status cu.OperationStatus) {
				log.Printf("[CU] status: %s", status)
			}),
		),
	}
}
