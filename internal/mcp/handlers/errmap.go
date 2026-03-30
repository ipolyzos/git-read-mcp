package handlers

import (
	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"

	gerr "github.com/ipolyzos/git-read-mcp/internal/errors"
	"github.com/ipolyzos/git-read-mcp/internal/logging"
)

// mapToolError logs a non-nil error to stderr and returns a stable *gerr.Error for the MCP tool result.
func mapToolError(tool string, req *mcpsdk.CallToolRequest, err error) error {
	if err == nil {
		return nil
	}
	mapped := gerr.Map(err)
	logging.ToolFailure(tool, req, mapped)
	return mapped
}
