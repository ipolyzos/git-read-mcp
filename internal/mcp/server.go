package mcp

import (
	"context"
	"fmt"

	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

// RunStdio builds an MCP server for the given implementation, installs tools from
// reg, and serves on stdin/stdout until context cancellation or transport error.
func RunStdio(ctx context.Context, impl *mcpsdk.Implementation, reg *Registry) error {
	if impl == nil {
		return fmt.Errorf("implementation is required")
	}
	if reg == nil {
		return fmt.Errorf("registry is required")
	}
	srv := mcpsdk.NewServer(impl, nil)
	if err := InstallTools(srv, reg); err != nil {
		return err
	}
	return srv.Run(ctx, &mcpsdk.StdioTransport{})
}
