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
5. `InstallTools(mcpsdk.Server, *Registry)` calls `Server.AddTool` for every entry. It fails if any handler is still `nil`. An empty registry is valid.

Handlers use the SDK’s low-level type `mcpsdk.ToolHandler`:

```go
func(context.Context, *mcpsdk.CallToolRequest) (*mcpsdk.CallToolResult, error)
```

Built-in tools (for example `diff_files`) are registered with the SDK’s generic `mcpsdk.AddTool` from `internal/mcp/handlers` and are passed into `RunStdio` via installer callbacks. Typed inputs/outputs live in `internal/mcp` (for example `DiffFilesInput` / `DiffFilesOutput`) so JSON schema is inferred by the SDK.

## Result envelope helpers

- `ToolSuccess(text, structured any)` — successful tool execution. `structured` must JSON-marshal to an **object** when set (MCP requirement for `structuredContent`).
- `ToolError(message)` — tool-level failure with `isError=true` and text content; prefer this over a returned Go `error` from low-level handlers when the model should see the failure and recover.

Protocol-level faults (bad JSON-RPC framing, unsupported methods) remain the SDK’s responsibility and surface as transport or handler errors.

## Serving on stdio

`RunStdio(ctx, impl, registry, installers...)` constructs `mcpsdk.NewServer`, applies `InstallTools` for the registry, runs optional `installers` (such as `handlers.InstallBuiltins`), then blocks on `StdioTransport` until the process is stopped or the client disconnects.

## Tool: `diff_files`

**Name:** `diff_files`  

**Purpose:** Read-only comparison of two **files** under a repository root using `git diff --no-index`.

### Inputs (JSON)

| Field | Type | Constraints |
|-------|------|-------------|
| `repo_path` | string | Must exist as a directory (validated with `NormalizeRepoPath`). |
| `left_path` | string | Relative to `repo_path` or absolute; must resolve inside the repo. |
| `right_path` | string | Same as `left_path`. |
| `context_lines` | int | `0`–`50`, passed to `--unified` (`0` means zero context lines). |
| `ignore_whitespace` | string (optional) | `none` (default), `all` → `--ignore-all-space`, `change` → `--ignore-space-change`. |

Relative paths are resolved with `ResolvePath`; both sides must pass `EnsureWithinRepo` and `EnsurePathExists` as files.

### Outputs (structured)

| Field | Meaning |
|-------|---------|
| `diff` | Unified diff text from git stdout. |
| `truncated` | True if the subprocess output hit the configured byte cap. |
| `had_diff` | True when git exited `1` (differences present); false when exit `0` (identical). |
| `stats` | When not truncated, approximate `insertions` / `deletions` by counting `+` / `-` lines (excluding `+++` / `---` headers). |

### Example arguments

```json
{
  "repo_path": "/home/you/project",
  "left_path": "pkg/foo.go",
  "right_path": "pkg/bar.go",
  "context_lines": 3,
  "ignore_whitespace": "none"
}
```

### Safety

- No shell: git is spawned via `execx` with a fixed argv; user strings are **never** interpreted as flags (paths are operands after `--`).
- Repo containment prevents reading arbitrary absolute paths outside `repo_path` when relative inputs are used; absolute inputs must still lie inside the repo after `ResolvePath` + `EnsureWithinRepo`.
- Git environment uses pager suppression and `GIT_TERMINAL_PROMPT=0` (see `internal/git` and `docs/architecture.md`).
- Truncation returns a tool success with `truncated: true` and partial `diff`; timeouts and unexpected git exits surface as errors.

## Validation rules (other planned tools)

| Tool / mode | Path policy |
|-------------|-------------|
| **`summarize_diff` (file mode)** | Same as `diff_files`: within repo, files must exist. |
| **`diff_directories`** | Paths may target arbitrary directories, but **`EnsurePathExists(dir, true)`** must succeed after resolution. Default: resolve relative paths with `ResolvePath`; absolute paths are accepted as-is after cleaning, **without** `EnsureWithinRepo`, so callers can compare directories outside the repo when explicitly specified. |
| **`compare_text`** | Operates on temporary/buffer-backed inputs supplied by the host; **no repo containment** on those payloads beyond verifying `repo_path` exists (for context) via `NormalizeRepoPath`. Do not accept user filesystem paths for compare blobs. |

Cross-cutting limits (description length, integer bounds, booleans) should use `internal/validate` helpers so MCP schemas and runtime checks stay aligned.
