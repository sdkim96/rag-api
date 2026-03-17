package mcputil

import (
	"context"
	"encoding/json"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

type HandlerSpec[In, Out any] struct {
	Name        string
	Description string
	Handler     func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error)
}

// Register registers a tool with the MCP server.
func Register[In, Out any](s *server.MCPServer, spec HandlerSpec[In, Out]) {
	s.AddTool(
		mcp.NewTool(
			spec.Name,
			mcp.WithDescription(spec.Description),
			mcp.WithInputSchema[In](),
			mcp.WithOutputSchema[Out](),
		),
		spec.Handler,
	)
}

// Convert deserializes MCP request arguments into T.
func Convert[T any](req mcp.CallToolRequest) (T, error) {
	var v T
	b, err := json.Marshal(req.GetArguments())
	if err != nil {
		return v, err
	}
	if err := json.Unmarshal(b, &v); err != nil {
		return v, err
	}
	return v, nil
}

// Error returns a tool result error.
func Error(err error) *mcp.CallToolResult {
	return mcp.NewToolResultError(err.Error())
}

// Text returns a tool result text.
func Text(text string) *mcp.CallToolResult {
	return mcp.NewToolResultText(text)
}

func JSON[T any](v T) (*mcp.CallToolResult, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	var m map[string]any
	if err := json.Unmarshal(b, &m); err != nil {
		return nil, err
	}
	return mcp.NewToolResultJSON(m)
}
