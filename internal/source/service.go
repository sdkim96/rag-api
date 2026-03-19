package source

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/sdkim96/indexing/uri"
	"github.com/sdkim96/rag-api/internal/infra/blobstore"
	"github.com/sdkim96/rag-api/internal/infra/db"
)

func FetchSource(ctx context.Context, origin Origin) (io.Reader, error) {
	u := uri.URI(origin.URI)
	err := u.Validate()
	if err != nil {
		return nil, err
	}
	switch u.Scheme() {
	case "http", "https":
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, origin.URI, http.NoBody)
		if err != nil {
			return nil, err
		}
		for k, v := range origin.Headers {
			req.Header.Set(k, v)
			req.Header.Set("Content-Type", origin.MimeType)
		}
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return nil, err
		}
		return resp.Body, nil
	case "file":
		f, err := os.OpenFile(u.Path(), os.O_RDONLY, 0644)
		if err != nil {
			return nil, err
		}
		return f, nil
	default:
		return nil, fmt.Errorf("Unsupported Scheme. %s", u.Scheme())

	}
}

func UploadBlob(ctx context.Context, s blobstore.BlobStore, key, mimeType string, r io.Reader) (Blob, error) {
	err := s.Upload(ctx, key, mimeType, r)
	if err != nil {
		return Blob{}, err
	}
	blob, err := s.Get(ctx, key)
	if err != nil {
		return Blob{}, err
	}
	return Blob{
		URI:      blob.URI(),
		Key:      key,
		MimeType: blob.MimeType(),
		Size:     blob.Size(),
	}, nil
}

func InsertDB(ctx context.Context, e *db.Engine, sid string, source Source) error {
	s, err := json.Marshal(source)
	if err != nil {
		return err
	}
	e.Write(ctx, db.NewSourceNS("rag-api"), sid, s)
	return nil
}

func ReadDB(ctx context.Context, e *db.Engine, offset, limit int, keyword string) ([]Source, error) {
	v, err := e.ReadAll(ctx, db.NewSourceNS("rag-api"),
		db.WithOffset(offset),
		db.WithLimit(limit),
		db.WithKeyword(keyword),
	)
	if err != nil {
		return nil, err
	}
	var srcs []Source
	for _, vv := range v {
		var s Source
		err = json.Unmarshal(vv, &s)
		if err != nil {
			return nil, err
		}
		srcs = append(srcs, s)
	}
	return srcs, nil
}
