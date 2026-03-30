package mcp

import (
	"context"
	"testing"

	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestRegistry_RegisterLookupList(t *testing.T) {
	t.Parallel()
	r := NewRegistry()
	h := func(context.Context, *mcpsdk.CallToolRequest) (*mcpsdk.CallToolResult, error) {
		return ToolSuccess("ok", nil), nil
	}
	if err := r.Register(ToolMeta{Name: "a", Description: "first"}, h); err != nil {
		t.Fatal(err)
	}
	if err := r.Register(ToolMeta{Name: "b", Description: "second"}, h); err != nil {
		t.Fatal(err)
	}

	meta, got, ok := r.Lookup("a")
	if !ok || meta.Name != "a" || meta.Description != "first" || got == nil {
		t.Fatalf("lookup a: ok=%v meta=%+v handlerNil=%v", ok, meta, got == nil)
	}
	if _, _, ok := r.Lookup("missing"); ok {
		t.Fatal("expected miss")
	}

	list := r.List()
	if len(list) != 2 || list[0].Name != "a" || list[1].Name != "b" {
		t.Fatalf("list order/names: %+v", list)
	}
}

func TestRegistry_RegisterReplaces(t *testing.T) {
	t.Parallel()
	r := NewRegistry()
	h1 := func(context.Context, *mcpsdk.CallToolRequest) (*mcpsdk.CallToolResult, error) {
		return ToolSuccess("one", nil), nil
	}
	h2 := func(context.Context, *mcpsdk.CallToolRequest) (*mcpsdk.CallToolResult, error) {
		return ToolSuccess("two", nil), nil
	}
	if err := r.Register(ToolMeta{Name: "x", Description: "old"}, h1); err != nil {
		t.Fatal(err)
	}
	if err := r.Register(ToolMeta{Name: "x", Description: "new"}, h2); err != nil {
		t.Fatal(err)
	}
	meta, _, ok := r.Lookup("x")
	if !ok || meta.Description != "new" {
		t.Fatalf("want replacement meta, got %+v ok=%v", meta, ok)
	}
}

func TestRegistry_RegisterEmptyName(t *testing.T) {
	t.Parallel()
	r := NewRegistry()
	if err := r.Register(ToolMeta{Name: ""}, stubHandler()); err == nil {
		t.Fatal("expected error for empty name")
	}
}

func TestInstallToolsNilHandler(t *testing.T) {
	t.Parallel()
	r := NewRegistry()
	if err := r.Register(ToolMeta{Name: "x", Description: "y"}, nil); err != nil {
		t.Fatal(err)
	}
	srv := mcpsdk.NewServer(&mcpsdk.Implementation{Name: "t", Version: "v"}, nil)
	if err := InstallTools(srv, r); err == nil {
		t.Fatal("expected error for nil handler")
	}
}

func TestInstallToolsEmptyRegistry(t *testing.T) {
	t.Parallel()
	r := NewRegistry()
	srv := mcpsdk.NewServer(&mcpsdk.Implementation{Name: "t", Version: "v"}, nil)
	if err := InstallTools(srv, r); err != nil {
		t.Fatal(err)
	}
}

func stubHandler() mcpsdk.ToolHandler {
	return func(context.Context, *mcpsdk.CallToolRequest) (*mcpsdk.CallToolResult, error) {
		return ToolSuccess("", nil), nil
	}
}
