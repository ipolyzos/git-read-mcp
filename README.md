# git-read-mcp

[![CI](https://github.com/ipolyzos/git-read-mcp/actions/workflows/ci.yml/badge.svg)](https://github.com/ipolyzos/git-read-mcp/actions/workflows/ci.yml)
[![Docs](https://github.com/ipolyzos/git-read-mcp/actions/workflows/docs.yml/badge.svg)](https://github.com/ipolyzos/git-read-mcp/actions/workflows/docs.yml)

A **read-only** [Model Context Protocol](https://modelcontextprotocol.io/) (MCP) server over **stdio** for inspecting Git history and diffs safely from AI tooling.

## What it does

The server exposes **read-only** MCP tools over stdio, backed by allowlisted `git diff --no-index` (no fetch/push/commit):

| Tool | Use |
|------|-----|
| **`diff_files`** | Unified diff between two **files** under a repo root (paths must stay inside `repo_path`). |
| **`diff_directories`** | Unified diff between two **directory trees** (absolute dirs may lie outside `repo_path`). |
| **`summarize_diff`** | Short summary + per-file stats from an inline unified diff **or** the same file pair as `diff_files`. |
| **`compare_text`** | Unified diff between two **strings** (`left_text` / `right_text`, capped); `repo_path` is context-only—blobs use a temp dir, not the repo tree. |

See **[docs/tools.md](docs/tools.md)** for parameters, limits, and security notes.

No writes, no branch manipulation, and no destructive Git operations are in scope.

## Usage examples (MCP)

Clients call **`tools/call`** with a **`name`** (tool id) and **`arguments`** (JSON object). The wire is JSON-RPC; only the **`params`** shape is shown below.

**Request** (`diff_files` — arguments only):

```json
{
  "name": "diff_files",
  "arguments": {
    "repo_path": "/home/you/project",
    "left_path": "README.md",
    "right_path": "docs/CHANGELOG.md",
    "context_lines": 3,
    "ignore_whitespace": "none"
  }
}
```

**Typical success** (trimmed): the SDK returns **structured content** derived from the tool output. You should read **`structuredContent`** (or the JSON **text** `content` entry, depending on client). Example shape:

```json
{
  "diff": "diff --git a/README.md b/README.md\n--- a/README.md\n+++ b/README.md\n@@ -1,3 +1,3 @@\n ...\n",
  "truncated": false,
  "had_diff": true,
  "stats": { "insertions": 2, "deletions": 1 }
}
```

If **`truncated`** is `true`, **`diff`** is partial and **`stats`** may be omitted—retry with smaller inputs, fewer context lines, or a higher `GIT_READ_MCP_MAX_OUTPUT_BYTES` (see [docs/troubleshooting.md](docs/troubleshooting.md)).

**Request** (`summarize_diff` — inline diff):

```json
{
  "name": "summarize_diff",
  "arguments": {
    "diff": "diff --git a/x b/x\n--- a/x\n+++ b/x\n@@ -1 +1 @@\n-old\n+new\n"
  }
}
```

**Typical success** (trimmed):

```json
{
  "summary": "1 file(s), +1 / -1 lines, 1 hunks (approx). Source: inline diff.",
  "files": [
    {
      "path": "x",
      "insertions": 1,
      "deletions": 1,
      "hunks": 1
    }
  ],
  "totals": { "files": 1, "insertions": 1, "deletions": 1, "hunks": 1 },
  "diff_source": "inline",
  "input_truncated": false,
  "parse_truncated": false
}
```

**Typical tool error** (validation): errors are JSON text with **`code`** and **`message`** (and `isError: true` on the result). Example:

```json
{
  "code": "INVALID_ARGUMENT",
  "message": "validate: repo_path is empty"
}
```

## Limitations and best practices

- **Paths**: Prefer **repo-relative** paths for `diff_files` / `summarize_diff` (path mode) so containment is clear. `diff_directories` can compare any two dirs—only use absolute paths you trust.
- **Large diffs**: Watch **`truncated`** / **`git_diff_truncated`** / **`input_truncated`** / **`parse_truncated`**. Use **`summarize_diff`** when you only need counts and file lists, or lower **`context_lines`** to shrink output.
- **Stats**: `stats` and `summarize_diff` counts are **approximate** (line-based), not `git numstat`.
- **Stdout**: The MCP session uses **stdout**; only the client should read it. Logs and `tool_failure` lines go to **stderr** ([docs/troubleshooting.md](docs/troubleshooting.md)).
- **Security**: See [docs/security.md](docs/security.md) and [docs/architecture.md](docs/architecture.md).

## Safety constraints

This project is designed to minimize blast radius when embedded in agents:

| Constraint | Intent |
|------------|--------|
| **No writes** | No `git commit`, `push`, `reset --hard`, index/worktree mutations, or hooks that alter state. |
| **No network** | Git subprocesses must not fetch, push, or otherwise use remotes; allowlists exclude networked commands. |
| **Allowlisted args** | Only an explicit set of `git` subcommands and flags is permitted; unknown argv is rejected. |
| **Timeouts** | Every subprocess has a bounded runtime so a hung `git` cannot block the host indefinitely. |
| **Output caps** | Diff and stdout/stderr are truncated or rejected when exceeding configured byte/line limits. |

See **[docs/security.md](docs/security.md)** for the threat model, path rules, and subprocess limits.

## Development quickstart

Prerequisites: [Go](https://go.dev/dl/) 1.25+ (required by the official MCP Go SDK).

```bash
go test ./...
go run ./cmd/server
```

Formatting, tests, lint, and build:

```bash
make fmt
make test
make lint
make build
```

## Layout

| Path | Role |
|------|------|
| `cmd/server` | Stdio MCP server binary |
| `internal/mcp` | MCP transport and tool registration |
| `internal/git` | Read-only Git operations |
| `internal/execx` | Bounded subprocess execution |
| `internal/validate` | Input and path validation |
| `internal/errors` | Errors and mapping for clients |
| `internal/logging` | Logging |
| `docs/` | Architecture, security, tools, releases |

## Community

- Code of Conduct: [CODE_OF_CONDUCT.md](CODE_OF_CONDUCT.md)
- Contributing guide: [CONTRIBUTING.md](CONTRIBUTING.md)
- Security policy: [SECURITY.md](SECURITY.md)

## License

Apache-2.0 — see [LICENSE](LICENSE).
