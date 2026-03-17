package server

import (
	"context"
	"fmt"
	"time"

	"github.com/mark3labs/mcp-go/server"
	"github.com/sdkim96/rag-api/config"
	"github.com/sdkim96/rag-api/internal/indexing"
	"github.com/sdkim96/rag-api/internal/infra/db"
)

type Server struct {
	mcp *server.MCPServer
	db  *db.Engine
}

func NewServer(cfg *config.Config) *Server {

	fmt.Println("[INIT] Initializing database connection...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	dbEngine, err := db.NewEngine(
		ctx,
		cfg.DB.DSN(),
		db.WithPing(ctx),
		db.WithMigrate(ctx),
	)
	if err != nil {
		panic(err)
	}

	fmt.Println("[INIT] Initializing MCP server and registering tools...")
	mcp := server.NewMCPServer(cfg.Project.Name, cfg.Project.Version)

	s := &Server{mcp: mcp, db: dbEngine}
	indexing.Register(s.mcp, dbEngine)

	return s

}

func (s *Server) RunStdio() error {
	return server.ServeStdio(s.mcp)
}

func (s *Server) RunHTTPStateless(addr string) error {
	httpServer := server.NewStreamableHTTPServer(s.mcp,
		server.WithEndpointPath("/api"),
		server.WithStateLess(true),
	)
	return httpServer.Start(addr)
}

func (s *Server) RunHTTPStateful(addr string) error {
	httpServer := server.NewStreamableHTTPServer(s.mcp,
		server.WithEndpointPath("/api"),
		server.WithStateful(true),
	)
	return httpServer.Start(addr)
}

func (s *Server) Close() error {
	if s.db == nil {
		return nil
	}
	return s.db.Conn().Close()
}
