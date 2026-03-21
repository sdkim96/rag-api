package indexing

import (
	"context"
	"time"

	mcpgo "github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/sdkim96/rag-api/internal/infra/db"
	"github.com/sdkim96/rag-api/internal/infra/indexer"
	"github.com/sdkim96/rag-api/internal/mcputil"
)

type DoIndexingReq struct {
	ID string `json:"id"`
}
type DoIndexingResp struct {
	StartedAt time.Time `json:"started_at"`
}

func Register(s *server.MCPServer, e *db.Engine, index *indexer.Indexer) {
	mcputil.RegisterWithSchema(s, mcputil.HandlerSpecWithSchema[DoIndexingReq, DoIndexingResp]{
		Name: "index_source",
		Description: `Start a source indexing process.
A indexing process represents the workflow of:
- fetching the content.
- analyzing the content.
- splitting the content.
- enriching the content.
- storing the content.

This endpoint asynchronously starts the indexing process for a source. It does not wait for the indexing to complete, nor does it return the result of the indexing process. 
It only triggers the process to start.
If you want to check the status or result of the indexing process,
you need to check 'read_source' endpoint with the source ID after some time.,
`,
		Handler: doHandler(e, index),
	})
}

func doHandler(e *db.Engine, index *indexer.Indexer) func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
	return func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		r, err := mcputil.Convert[DoIndexingReq](req)
		if err != nil {
			return mcputil.Error(err), nil
		}
		go DoIndexing(context.Background(), e, index, r.ID)

		resp := DoIndexingResp{StartedAt: time.Now()}
		return mcputil.JSON(resp)
	}
}
