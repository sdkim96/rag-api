package search

import (
	"context"

	mcpgo "github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/openai/openai-go"
	"github.com/sdkim96/rag-api/internal/infra/db"
	"github.com/sdkim96/rag-api/internal/mcputil"
)

type SearchSourceReq struct {
	Query    string   `json:"query"`
	Keywords []string `json:"keywords,omitempty"`
}

type SearchSourceResp struct {
	Content string `json:"content"`
}

func Register(s *server.MCPServer, e *db.Engine, oaiClient openai.Client) {
	mcputil.RegisterWithInSchema(s, mcputil.HandlerSpecWithInSchema[SearchSourceReq]{
		Name: "search_source",
		Description: `Search for relevant information based on a query.
This endpoint allows you to find relevant information across all indexed sources based on a natural language query.
The search results include a summary, topic, keywords, and the relevant parts of the source content.`,
		Handler: searchHandler(context.Background(), e, oaiClient),
	})
}

func searchHandler(ctx context.Context, e *db.Engine, oaiClient openai.Client) func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
	return func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		r, err := mcputil.Convert[SearchSourceReq](req)
		if err != nil {
			return mcputil.Error(err), nil
		}

		availableKeys, err := SelectSearchableSourceIDs(ctx, e, db.System)
		if err != nil {
			return mcputil.Error(err), nil
		}

		embedding, err := Embed(ctx, oaiClient, r.Query)
		if err != nil {
			return mcputil.Error(err), nil
		}

		docs, err := HybridSearch(ctx, e, embedding, r.Query, availableKeys, 10)
		if err != nil {
			return mcputil.Error(err), nil
		}
		if len(docs) == 0 {
			return mcputil.Text("No relevant information found."), nil
		}

		var content string
		for _, doc := range docs {
			content += "# " + doc.Topic + "\n"
			cuparts, err := SelectParts(ctx, e, doc.SourceID, doc.PartIdxs)
			if err != nil {
				return mcputil.Error(err), nil
			}
			for _, part := range cuparts {
				content += part.Text() + "\n"
			}
			content += "---\n"
		}
		return mcputil.Text(content), nil
	}
}
