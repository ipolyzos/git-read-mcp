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
	mcpsdk.AddTool(srv, &mcpsdk.Tool{
		Name:        apimcp.ToolNameSummarizeDiff,
		Description: "Summarize a unified diff: file list, approximate +/- lines and hunk counts. Pass diff text or repo_path+left_path+right_path (same rules as diff_files) to compute via git diff --no-index.",
	}, HandleSummarizeDiff)
	mcpsdk.AddTool(srv, &mcpsdk.Tool{
		Name:        apimcp.ToolNameCompareText,
		Description: "Compare two text blobs in memory via git diff --no-index (read-only). repo_path is validated for context only; operands are written to a temp directory, not under the repo.",
	}, HandleCompareText)
	return nil
}
