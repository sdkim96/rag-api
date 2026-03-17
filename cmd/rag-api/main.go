package main

import (
	"flag"
	"fmt"

	"github.com/sdkim96/rag-api/config"
	"github.com/sdkim96/rag-api/internal/server"
)

func main() {
	mode := flag.String("mode", "stdio", "transport mode: stdio, http, or http-stateful")
	addr := flag.String("addr", ":8080", "http listen address")
	flag.Parse()

	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("failed to load config: %v\n", err)
		return
	}

	switch *mode {
	case "stdio":
		err = server.NewServer(cfg).RunStdio()
	case "http":
		fmt.Printf("listening on %s (stateless)\n", *addr)
		err = server.NewServer(cfg).RunHTTPStateless(*addr)
	case "http-stateful":
		fmt.Printf("listening on %s (stateful)\n", *addr)
		err = server.NewServer(cfg).RunHTTPStateful(*addr)
	default:
		err = fmt.Errorf("unknown mode: %s", *mode)
	}

	if err != nil {
		fmt.Printf("server error: %v\n", err)
	}
}
