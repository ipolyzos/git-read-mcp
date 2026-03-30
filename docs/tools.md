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

Built-in tools (`diff_files`, `diff_directories`, `summarize_diff`, …) are registered with the SDK’s generic `mcpsdk.AddTool` from `internal/mcp/handlers` and are passed into `RunStdio` via installer callbacks. Typed inputs/outputs live in `internal/mcp` so JSON schema is inferred by the SDK.

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

## Tool: `diff_directories`

**Name:** `diff_directories`  

**Purpose:** Read-only recursive comparison of two **directories** using `git diff --no-index`. Git walks both trees and emits a unified diff (requires a Git version that supports directory operands with `--no-index`, e.g. recent 2.x).

### Inputs (JSON)

| Field | Type | Constraints |
|-------|------|-------------|
| `repo_path` | string | Must exist as a directory; used to resolve **relative** `left_dir` / `right_dir`. |
| `left_dir` | string | Relative to `repo_path` or absolute path to a directory. |
| `right_dir` | string | Same as `left_dir`. |
| `context_lines` | int | `0`–`50` (`--unified`). |
| `ignore_whitespace` | string (optional) | Same semantics as `diff_files` (`none` / `all` / `change`). |

Relative paths are joined against `repo_path` via `ResolvePath`. **There is no `EnsureWithinRepo` check**: absolute paths may point outside the repository, but both operands must exist and be **directories** (`EnsurePathExists(..., true)`).

### Outputs (structured)

Same shape as `diff_files`: `diff`, `truncated`, `had_diff`, `stats` (stats omitted when output is truncated in the handler).

### Example: relative trees inside a repo

```json
{
  "repo_path": "/home/you/project",
  "left_dir": "baseline/pkg",
  "right_dir": "current/pkg",
  "context_lines": 3,
  "ignore_whitespace": "none"
}
```

### Example: absolute directories (e.g. two clones)

```json
{
  "repo_path": "/home/you/project",
  "left_dir": "/tmp/checkout-a",
  "right_dir": "/tmp/checkout-b",
  "context_lines": 3,
  "ignore_whitespace": "change"
}
```

Here `repo_path` still must be a valid directory (anchor for resolution); the compared trees live at the absolute paths.

### Safety

- Same argv discipline as `diff_files`: no shell; operands appear only after `--`.
- Wider filesystem reach than `diff_files` by design—only use absolute paths when you trust their source; relative paths stay anchored to `repo_path`.
- Very large trees can produce huge diffs; default subprocess output caps apply (`internal/git` / `execx`); `truncated` is set when the cap is hit.

## Tool: `summarize_diff`

**Name:** `summarize_diff`

**Purpose:** Produce a **short summary** plus **structured per-file stats** from a **unified diff**. Either pass raw diff text or the same path triple as `diff_files` so the server runs `git diff --no-index` and parses the result.

### Inputs (JSON)

| Field | Type | Constraints |
|-------|------|-------------|
| `diff` | string (optional) | Inline unified diff. If non-empty (after trim), **path fields are ignored**. |
| `repo_path` | string | Required when `diff` is empty: repository root (see `diff_files`). |
| `left_path` | string | Required with `repo_path` when computing: left file, repo-scoped. |
| `right_path` | string | Required with `repo_path` when computing: right file, repo-scoped. |
| `context_lines` | int | `0`–`50` when computing (same as `diff_files`). |
| `ignore_whitespace` | string (optional) | When computing: `none` / `all` / `change`. |
| `max_diff_bytes` | int (optional) | Max bytes fed to the parser (default 2 MiB, hard cap 8 MiB). Inline `diff` **longer** than the effective cap is rejected before parsing. |
| `max_files` | int (optional) | Stop collecting after this many files (default 500, hard cap 10 000). |
| `max_hunks` | int (optional) | Stop after this many hunks **total** across files (default 10 000, hard cap 100 000). |

### Outputs (structured)

| Field | Meaning |
|-------|---------|
| `summary` | One-line narrative: file count, approximate `+`/`-` lines, hunk count, source mode, truncation flags. |
| `files` | List of `{ path, old_path?, rename?, binary?, insertions, deletions, hunks }`. |
| `totals` | Rolled-up `files`, `insertions`, `deletions`, `hunks`. |
| `diff_source` | `"inline"` or `"computed"`. |
| `input_truncated` | Diff text was sliced to `max_diff_bytes` before parsing. |
| `parse_truncated` | Parsing stopped early because `max_files` or `max_hunks` was hit. |
| `git_diff_truncated` | Git subprocess output hit the output byte cap (same as `diff_files` `truncated`). Partial diff is still summarized when possible. |

### Examples

Inline diff:

```json
{
  "diff": "diff --git a/foo.go b/foo.go\n--- a/foo.go\n+++ b/foo.go\n@@ -1 +1 @@\n-old\n+new\n"
}
```

Computed (equivalent policy to `diff_files`):

```json
{
  "repo_path": "/home/you/project",
  "left_path": "pkg/old.go",
  "right_path": "pkg/new.go",
  "context_lines": 3,
  "max_hunks": 5000
}
```

### Limitations

- **Counts are approximate**: they come from a line scan over unified diff text (like counting `+`/`-` in hunks, excluding `+++` / `---` lines). They are not a full Git numstat.
- **Renames** are detected best-effort from `diff --git`, `rename from` / `rename to`, and `---` / `+++` headers; unusual formats may mis-label paths.
- **Binary files**: lines of the form `Binary files … differ` mark the file as binary with **no** line-level insertions/deletions.
- **Truncation**: `input_truncated` / `parse_truncated` / `git_diff_truncated` mean the summary may be **incomplete**; callers should not assume full coverage without checking those flags.

### Safety

- **Path mode** inherits the same containment and argv discipline as `diff_files` (no shell; operands only after `--`).
- Inline mode parses **only** the supplied string (capped by `max_diff_bytes` for parsing); oversized inline input is rejected up front.

## Validation rules (other planned tools)

| Tool / mode | Path policy |
|-------------|-------------|
| **`compare_text`** | Operates on temporary/buffer-backed inputs supplied by the host; **no repo containment** on those payloads beyond verifying `repo_path` exists (for context) via `NormalizeRepoPath`. Do not accept user filesystem paths for compare blobs. |

Cross-cutting limits (description length, integer bounds, booleans) should use `internal/validate` helpers so MCP schemas and runtime checks stay aligned.
