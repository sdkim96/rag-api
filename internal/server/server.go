package server

import (
	"context"
	"fmt"
	"time"

	"github.com/mark3labs/mcp-go/server"
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"github.com/sdkim96/rag-api/config"
	"github.com/sdkim96/rag-api/internal/indexing"
	"github.com/sdkim96/rag-api/internal/infra/blobstore"
	"github.com/sdkim96/rag-api/internal/infra/db"
	"github.com/sdkim96/rag-api/internal/infra/indexer"
	"github.com/sdkim96/rag-api/internal/search"
	"github.com/sdkim96/rag-api/internal/source"
)

type Server struct {
	mcp *server.MCPServer
	db  *db.Engine
	bs  blobstore.BlobStore
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

	bs, err := blobstore.NewAzureBlobStore(
		cfg.AzureBlobStore.AccountName,
		cfg.AzureBlobStore.ConnString,
		cfg.AzureBlobStore.ContainerName,
	)
	if err != nil {
		panic(err)
	}
	oaiClient := openai.NewClient(
		option.WithAPIKey(cfg.OpenAI.APIKey),
	)

	index, err := indexer.New(cfg, dbEngine, bs)
	if err != nil {
		panic(err)
	}

	fmt.Println("[INIT] Initializing MCP server and registering tools...")
	mcp := server.NewMCPServer(cfg.Project.Name, cfg.Project.Version)

	s := &Server{mcp: mcp, db: dbEngine, bs: bs}

	source.Register(s.mcp, dbEngine, bs)
	indexing.Register(s.mcp, dbEngine, index)
	search.Register(s.mcp, dbEngine, oaiClient)

	fmt.Printf("[INIT] Server initialization completed. Endpoints: %v", s.mcp.ListTools())

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
