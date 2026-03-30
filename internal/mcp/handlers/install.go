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
		Name:        apimcp.ToolNameCompareText,
		Description: "Compare two text blobs using a private temp directory and git diff --no-index; repo_path is not written.",
	}, HandleCompareText)
	return nil
}
