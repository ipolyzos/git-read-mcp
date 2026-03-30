# Tools and MCP wiring

## Protocol stack

Transport and JSON-RPC session lifecycle (including `initialize`, capability negotiation, and keep-alives such as `ping` when supported by the client and spec version) are implemented by the official SDK:

[`github.com/modelcontextprotocol/go-sdk/mcp`](https://github.com/modelcontextprotocol/go-sdk)

git-read-mcp adds a thin layer in `internal/mcp` for **tool registration**, **shared result shaping**, and **stdio wiring** (`RunStdio`). The server process must treat **stdout as the MCP wire** and **stderr as logs only** (see `cmd/server/main.go`).

## Tool registry

`internal/mcp.Registry` is the in-process catalog of tools:

1. Construct with `NewRegistry()`.
2. For each tool, call `Register(ToolMeta, mcpsdk.ToolHandler)` with a non-nil handler before serving.
3. `List()` returns stable metadata sorted by tool name (useful for tests and future introspection).
4. `Lookup(name)` fetches metadata plus the handler.
5. `InstallTools(mcpsdk.Server, *Registry)` calls `Server.AddTool` for every entry. It fails if any handler is still `nil`. An empty registry is valid and yields a server that advertises no tools.

Handlers use the SDK’s low-level type `mcpsdk.ToolHandler`:

```go
func(context.Context, *mcpsdk.CallToolRequest) (*mcpsdk.CallToolResult, error)
```

For typed inputs/outputs and automatic schema inference, the SDK’s generic `mcpsdk.AddTool` can be adopted later; the registry today wraps the explicit `ToolHandler` shape so git-specific tools can share one registration path.

## Result envelope helpers

- `ToolSuccess(text, structured any)` — successful tool execution. `structured` must JSON-marshal to an **object** when set (MCP requirement for `structuredContent`).
- `ToolError(message)` — tool-level failure with `isError=true` and text content; prefer this over a Go `error` return when the model should see the failure and recover.

Protocol-level faults (bad JSON-RPC framing, unsupported methods) remain the SDK’s responsibility and surface as transport or handler errors.

## Serving on stdio

`RunStdio(ctx, impl, registry)` constructs `mcpsdk.NewServer`, installs tools, and blocks on `StdioTransport` until the context is canceled (SIGINT/SIGTERM in `main`) or the client drops the session.

## Current status

The registry ships **empty**: no Git-backed tools are registered yet. Adding a tool is: implement `ToolHandler`, `Register`, ensure `RunStdio` receives that registry.

## Validation rules (planned tools)

All tools that touch the filesystem should call `NormalizeRepoPath` once per request for `repo_path`. Additional rules depend on the tool:

| Tool / mode | Path policy |
|-------------|-------------|
| **`diff_files`**, **`summarize_diff` (file mode)** | Resolve user file paths with `ResolvePath(repo, userPath)`, then **`EnsureWithinRepo(repo, resolved)`**. Require files to exist when the diff requires readable blobs. |
| **`diff_directories`** | Paths may target arbitrary directories, but **`EnsurePathExists(dir, true)`** must succeed after resolution. Default: resolve relative paths with `ResolvePath`; absolute paths are accepted as-is after cleaning, **without** `EnsureWithinRepo`, so callers can compare directories outside the repo when explicitly specified. |
| **`compare_text`** | Operates on temporary/buffer-backed inputs supplied by the host; **no repo containment** on those payloads beyond verifying `repo_path` exists (for context) via `NormalizeRepoPath`. Do not accept user filesystem paths for compare blobs. |

Cross-cutting limits (description length, integer bounds, booleans) should use `internal/validate` helpers so MCP schemas and runtime checks stay aligned.
