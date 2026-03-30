package mcp

import (
	"fmt"
	"sort"
	"sync"

	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

// ToolMeta is the stable description for a tool listing.
type ToolMeta struct {
	Name        string
	Description string
}

// Registry stores tool metadata and low-level MCP handlers in one place.
type Registry struct {
	mu    sync.RWMutex
	tools map[string]registeredTool
}

type registeredTool struct {
	meta    ToolMeta
	handler mcpsdk.ToolHandler
}

// NewRegistry returns an empty registry.
func NewRegistry() *Registry {
	return &Registry{tools: make(map[string]registeredTool)}
}

// Register adds or replaces a tool. Handler must be non-nil before InstallTools runs.
func (r *Registry) Register(meta ToolMeta, handler mcpsdk.ToolHandler) error {
	if meta.Name == "" {
		return fmt.Errorf("tool name is required")
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.tools[meta.Name] = registeredTool{meta: meta, handler: handler}
	return nil
}

// Lookup returns metadata and handler for a tool name.
func (r *Registry) Lookup(name string) (ToolMeta, mcpsdk.ToolHandler, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	t, ok := r.tools[name]
	if !ok {
		return ToolMeta{}, nil, false
	}
	return t.meta, t.handler, true
}

// List returns tool metadata sorted by name.
func (r *Registry) List() []ToolMeta {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]ToolMeta, 0, len(r.tools))
	for _, t := range r.tools {
		out = append(out, t.meta)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

// InstallTools attaches every registered tool to the MCP server.
func InstallTools(server *mcpsdk.Server, reg *Registry) error {
	reg.mu.RLock()
	defer reg.mu.RUnlock()
	for name, t := range reg.tools {
		if t.handler == nil {
			return fmt.Errorf("tool %q: handler is nil", name)
		}
		server.AddTool(&mcpsdk.Tool{Name: t.meta.Name, Description: t.meta.Description}, t.handler)
	}
	return nil
}
