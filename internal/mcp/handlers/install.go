package handlers

import (
	"fmt"

	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"

	apimcp "github.com/ipolyzos/git-read-mcp/internal/mcp"
)

// InstallBuiltins registers stock tools on the MCP server.
func InstallBuiltins(srv *mcpsdk.Server) error {
	if srv == nil {
		return fmt.Errorf("handlers: server is nil")
	}
	mcpsdk.AddTool(srv, &mcpsdk.Tool{
		Name:        apimcp.ToolNameDiffFiles,
		Description: "Compare two files inside a repository using git diff --no-index (read-only).",
	}, HandleDiffFiles)
	mcpsdk.AddTool(srv, &mcpsdk.Tool{
		Name:        apimcp.ToolNameDiffDirectories,
		Description: "Compare two directory trees using git diff --no-index (read-only). Relative paths resolve from repo_path; absolute paths may be outside the repo.",
	}, HandleDiffDirectories)
	return nil
}
