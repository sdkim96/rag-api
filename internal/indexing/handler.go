package indexing

import (
	"context"

	mcpgo "github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/sdkim96/rag-api/internal/infra/db"
	"github.com/sdkim96/rag-api/internal/mcputil"
)

type AddRequest struct {
	Noop string `json:"noop"`
}
type AddResponse struct {
	Noop string `json:"noop"`
}

type DeleteRequest struct {
	Noop string `json:"noop"`
}
type DeleteResponse struct {
	Noop string `json:"noop"`
}

func Register(s *server.MCPServer, db *db.Engine) {
	mcputil.Register(s, mcputil.HandlerSpec[AddRequest, AddResponse]{
		Name:        "index_document",
		Description: "Index a document",
		Handler:     addHandler(db),
	})
	mcputil.Register(s, mcputil.HandlerSpec[DeleteRequest, DeleteResponse]{
		Name:        "delete_document",
		Description: "Delete a document",
		Handler:     deleteHandler(db),
	})
}

func addHandler(db *db.Engine) func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
	return func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		r, err := mcputil.Convert[AddRequest](req)
		if err != nil {
			return mcputil.Error(err), nil
		}
		// Business logic
		_ = r
		return mcputil.Text("indexed"), nil
	}
}

func deleteHandler(db *db.Engine) func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
	return func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		r, err := mcputil.Convert[DeleteRequest](req)
		if err != nil {
			return mcputil.Error(err), nil
		}
		// Business logic
		_ = r
		return mcputil.Text("deleted"), nil
	}
}
