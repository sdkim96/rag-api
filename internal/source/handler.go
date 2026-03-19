package source

import (
	"context"
	"time"

	"github.com/google/uuid"
	mcpgo "github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/sdkim96/rag-api/internal/infra/blobstore"
	"github.com/sdkim96/rag-api/internal/infra/db"
	"github.com/sdkim96/rag-api/internal/mcputil"
)

type AddSourceReq struct {
	URI      string            `json:"uri"`
	MimeType string            `json:"mime_type"`
	Headers  map[string]string `json:"headers,omitempty"`
}
type AddSourceResp struct {
	ID      string `json:"id"`
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
}

type ReadSourcesReq struct {
	Offset  int    `json:"offset"`
	Limit   int    `json:"limit"`
	Keyword string `json:"keyword,omitempty"`
}

type ReadSourcesResp struct {
	Sources []Source `json:"sources"`
	Offset  int      `json:"offset"`
	Limit   int      `json:"limit"`
	Total   int      `json:"total"`
}

type DeleteSourceReq struct {
	ID string `json:"id"`
}
type DeleteSourceResp struct {
	ID     string `json:"id"`
	Status string `json:"status"`
}

func Register(s *server.MCPServer, db *db.Engine, bs blobstore.BlobStore) {
	mcputil.Register(s, mcputil.HandlerSpec[AddSourceReq, AddSourceResp]{
		Name: "add_source",
		Description: `Register a source URI to be indexed.

A source is an accessible endpoint that this system will fetch, analyze, and index.
You do not need to upload the file — just provide a URI and MimeType that this system can access.
The system will download and process the content automatically.

Headers can be included for authentication or other purposes when fetching the source.
Only the Headers specified in the request will be sent; default headers are not included.

Supported URI schemes:
- file://   Local filesystem path  (e.g. file:///data/report.pdf)
- http://   HTTP endpoint          (e.g. http://internal.server/doc.pdf)
- https://  HTTPS endpoint         (e.g. https://storage.example.com/report.pdf)
`,
		Handler: addHandler(db, bs),
	})
	mcputil.Register(s, mcputil.HandlerSpec[ReadSourcesReq, ReadSourcesResp]{
		Name: "read_sources",
		Description: `List all registered sources and their indexing status.

	Returns all sources registered for indexing along with their current status.

	Status values:
	- registered  Source has been registered but not yet indexed
	- indexing    Currently being processed by the pipeline
	- indexed     Successfully indexed and searchable
	- error       Processing failed`,
		Handler: readHandler(db),
	})
	mcputil.Register(s, mcputil.HandlerSpec[DeleteSourceReq, DeleteSourceResp]{
		Name:        "delete_source",
		Description: "Delete a registered source and its indexed data",
		Handler:     deleteHandler(db),
	})
}
func addHandler(db *db.Engine, bs blobstore.BlobStore) func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
	return func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		r, err := mcputil.Convert[AddSourceReq](req)
		if err != nil {
			return mcputil.Error(err), nil
		}

		sid, err := uuid.NewV7()
		if err != nil {
			return mcputil.Error(err), nil
		}
		org := Origin{URI: r.URI, MimeType: r.MimeType, Headers: r.Headers}

		blobUpload, err := FetchSource(ctx, org)
		if err != nil {
			return mcputil.Error(err), nil
		}

		blob, err := UploadBlob(ctx, bs, sid.String(), org.MimeType, blobUpload)
		if err != nil {
			return mcputil.Error(err), nil
		}

		source := Source{
			ID:        sid.String(),
			Origin:    org,
			Blob:      blob,
			CreatedAt: time.Now(),
		}

		err = InsertDB(ctx, db, sid.String(), source)
		if err != nil {
			return mcputil.Error(err), nil
		}

		resp := AddSourceResp{
			ID:     sid.String(),
			Status: "ok",
		}
		return mcputil.JSON(resp)
	}
}

func readHandler(db *db.Engine) func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
	return func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		r, err := mcputil.Convert[ReadSourcesReq](req)
		if err != nil {
			return mcputil.Error(err), nil
		}

		srcs, err := ReadDB(ctx, db, r.Offset, r.Limit, r.Keyword)
		if err != nil {
			return mcputil.Error(err), nil
		}

		resp := ReadSourcesResp{
			Sources: srcs,
			Offset:  r.Offset,
			Limit:   r.Limit,
			Total:   len(srcs),
		}
		return mcputil.JSON(resp)
	}
}

func deleteHandler(db *db.Engine) func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
	return func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		r, err := mcputil.Convert[DeleteSourceReq](req)
		if err != nil {
			return mcputil.Error(err), nil
		}
		// Business logic
		_ = r
		resp := DeleteSourceResp{
			ID:     r.ID,
			Status: "ok",
		}
		return mcputil.JSON(resp)
	}
}
