// Package logging writes structured diagnostics to stderr only (never stdout).
package logging

import (
	"encoding/json"
	"os"

	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"

	gerr "github.com/ipolyzos/git-read-mcp/internal/errors"
)

// ToolFailure logs a tool execution failure with stable fields. Safe for stdio MCP: writes only to stderr.
func ToolFailure(tool string, req *mcpsdk.CallToolRequest, mapped *gerr.Error) {
	if mapped == nil {
		return
	}
	rec := map[string]any{
		"level":   "error",
		"msg":     "tool_failure",
		"tool":    tool,
		"code":    string(mapped.Code),
		"message": mapped.Message,
	}
	if req != nil && req.Session != nil {
		if id := req.Session.ID(); id != "" {
			rec["session_id"] = id
		}
	}
	// Params.Name is redundant with tool but helps when req.Session is nil in tests.
	if req != nil && req.Params != nil && req.Params.Name != "" {
		rec["request_tool"] = req.Params.Name
	}
	b, err := json.Marshal(rec)
	if err != nil {
		_, _ = os.Stderr.WriteString(`{"level":"error","msg":"tool_failure_log_marshal_failed"}` + "\n")
		return
	}
	_, _ = os.Stderr.Write(append(b, '\n'))
}
