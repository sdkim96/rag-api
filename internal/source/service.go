package source

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/sdkim96/indexing/urio"
	"github.com/sdkim96/rag-api/internal/infra/blobstore"
	"github.com/sdkim96/rag-api/internal/infra/db"
)

func FetchSource(ctx context.Context, origin Origin) (io.Reader, error) {
	u := urio.URI(origin.URI)
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
		URI:       blob.URI(),
		Key:       key,
		MimeType:  blob.MimeType(),
		Size:      blob.Size(),
		CreatedAt: time.Now(),
	}, nil
}

func InsertSource(ctx context.Context, e *db.Engine, s Source) error {

	origin, err := json.Marshal(s.Origin)
	if err != nil {
		return err
	}

	return e.InsertSource(
		ctx,
		s.ID,
		db.System,
		s.URI,
		s.MimeType,
		s.Name,
		s.Size,
		origin,
	)
}

func SelectSources(ctx context.Context, e *db.Engine, r ReadSourcesReq) ([]SourceIndexing, error) {
	srcBytes, err := e.SelectSources(ctx, r.ID, r.Offset, r.Limit, r.Keyword)
	if err != nil {
		return nil, err
	}
	var srcs []SourceIndexing
	for _, b := range srcBytes {
		var srcI SourceIndexing
		err := json.Unmarshal(b, &srcI)
		if err != nil {
			return nil, err
		}
		if srcI.Status == "" {
			srcI.Status = StatusNotStarted
		}
		srcs = append(srcs, srcI)
	}
	return srcs, nil
}

func DeleteSource(ctx context.Context, e *db.Engine, sid string) error {
	return e.DeleteSource(ctx, sid)
}
