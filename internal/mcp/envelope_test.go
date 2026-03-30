package mcp

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestToolError(t *testing.T) {
	t.Parallel()
	r := ToolError("boom")
	if !r.IsError {
		t.Fatal("expected IsError")
	}
	if len(r.Content) != 1 {
		t.Fatalf("content: %#v", r.Content)
	}
}

func TestToolSuccessStructuredMarshalsAsObject(t *testing.T) {
	t.Parallel()
	r := ToolSuccess("note", map[string]any{"ok": true})
	data, err := json.Marshal(r)
	if err != nil {
		t.Fatal(err)
	}
	s := string(data)
	if !strings.Contains(s, `"content"`) || !strings.Contains(s, "structuredContent") {
		t.Fatalf("unexpected JSON: %s", s)
	}
}

func TestToolSuccessTextOnly(t *testing.T) {
	t.Parallel()
	r := ToolSuccess("plain", nil)
	if len(r.Content) != 1 {
		t.Fatalf("content: %#v", r.Content)
	}
	if r.StructuredContent != nil {
		t.Fatalf("structured: %#v", r.StructuredContent)
	}
}
