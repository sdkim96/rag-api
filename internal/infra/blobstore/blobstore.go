package blobstore

import (
	"context"
	"io"
)

type Blob interface {
	URI() string
	MimeType() string
	Size() int64
}

type BlobStore interface {
	Get(ctx context.Context, key string) (Blob, error)
	Upload(ctx context.Context, key, mimeType string, r io.Reader) (err error)
	Download(ctx context.Context, key string) (io.ReadCloser, error)
	Delete(ctx context.Context, key string) error
}
