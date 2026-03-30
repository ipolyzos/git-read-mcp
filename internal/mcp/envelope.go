package mcp

import (
	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

// ToolError returns a tool-level failure result (IsError=true). Use this when the
// tool ran but could not satisfy the request; protocol-level problems should
// surface as a returned Go error from the handler instead.
func ToolError(message string) *mcpsdk.CallToolResult {
	return &mcpsdk.CallToolResult{
		IsError: true,
		Content: []mcpsdk.Content{&mcpsdk.TextContent{Text: message}},
	}
}

// ToolSuccess builds a successful tool result with optional plain text and structured
// payload. StructuredContent must JSON-marshal to an object when non-nil.
func ToolSuccess(text string, structured any) *mcpsdk.CallToolResult {
	res := &mcpsdk.CallToolResult{}
	if text != "" {
		res.Content = []mcpsdk.Content{&mcpsdk.TextContent{Text: text}}
	}
	if structured != nil {
		res.StructuredContent = structured
	}
	return res
}
