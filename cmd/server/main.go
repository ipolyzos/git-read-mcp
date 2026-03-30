// Package main is the stdio MCP server entrypoint for git-read-mcp.
package main

import (
	"context"
	"errors"
	"log"
	"os"
	"os/signal"
	"syscall"

	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"

	gitreadmcp "github.com/ipolyzos/git-read-mcp/internal/mcp"
	"github.com/ipolyzos/git-read-mcp/internal/mcp/handlers"
)

func main() {
	log.SetOutput(os.Stderr)
	log.SetPrefix("git-read-mcp: ")
	log.SetFlags(log.LstdFlags)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	reg := gitreadmcp.NewRegistry()
	impl := &mcpsdk.Implementation{Name: "git-read-mcp", Version: version()}
	if err := gitreadmcp.RunStdio(ctx, impl, reg, handlers.InstallBuiltins); err != nil && !errors.Is(err, context.Canceled) {
		log.Fatalf("server stopped: %v", err)
	}
}

func version() string {
	if v := os.Getenv("GIT_READ_MCP_VERSION"); v != "" {
		return v
	}
	return "0.0.0-dev"
}
