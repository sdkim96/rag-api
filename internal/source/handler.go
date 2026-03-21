package source

import (
	"context"

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
	Title    string            `json:"title"`
	Headers  map[string]string `json:"headers,omitempty"`
	Meta     *SourceMeta       `json:"source_meta,omitempty"`
}

type AddSourceResp struct {
	ID      string `json:"id"`
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
}

type ReadSourcesReq struct {
	ID      []string `json:"id,omitempty"`
	Offset  int      `json:"offset"`
	Limit   int      `json:"limit"`
	Keyword string   `json:"keyword,omitempty"`
}

type ReadSourcesResp struct {
	Sources []SourceIndexing `json:"sources"`
	Offset  int              `json:"offset"`
	Limit   int              `json:"limit"`
	Total   int              `json:"total"`
}

type DeleteSourceReq struct {
	ID string `json:"id"`
}
type DeleteSourceResp struct {
	ID     string `json:"id"`
	Status string `json:"status"`
}

func Register(s *server.MCPServer, e *db.Engine, bs blobstore.BlobStore) {
	mcputil.RegisterWithSchema(s, mcputil.HandlerSpecWithSchema[AddSourceReq, AddSourceResp]{
		Name: "add_source",
		Description: `Register a source URI to be indexed.
Note that this endpoint does not index the content immediately. It only registers the source being ready for indexing. 

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
		Handler: addHandler(e, bs),
	})
	mcputil.RegisterWithSchema(s, mcputil.HandlerSpecWithSchema[ReadSourcesReq, ReadSourcesResp]{
		Name: "read_sources",
		Description: `List all registered sources and their indexing status.

	Returns all sources registered for indexing along with their current status.

	Status values:
	- registered  Source has been registered but not yet indexed
	- indexing    Currently being processed by the pipeline
	- indexed     Successfully indexed and searchable
	- error       Processing failed`,
		Handler: readHandler(e),
	})
	mcputil.RegisterWithSchema(s, mcputil.HandlerSpecWithSchema[DeleteSourceReq, DeleteSourceResp]{
		Name:        "delete_source",
		Description: "Delete a registered source and its indexed data",
		Handler:     deleteHandler(e),
	})
}
func addHandler(e *db.Engine, bs blobstore.BlobStore) func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
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
			OwnerID:   db.System,
			URI:       blob.URI,
			MimeType:  blob.MimeType,
			Name:      r.Title,
			Origin:    org,
			Size:      blob.Size,
			CreatedAt: blob.CreatedAt,
		}

		err = InsertSource(ctx, e, source)
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

		srcs, err := SelectSources(ctx, db, r)
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
		err = DeleteSource(ctx, db, r.ID)
		if err != nil {
			return mcputil.Error(err), nil
		}
		resp := DeleteSourceResp{
			ID:     r.ID,
			Status: "ok",
		}
		return mcputil.JSON(resp)
	}
}
