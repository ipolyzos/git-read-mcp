# git-read-mcp

[![CI](https://github.com/ipolyzos/git-read-mcp/actions/workflows/ci.yml/badge.svg)](https://github.com/ipolyzos/git-read-mcp/actions/workflows/ci.yml)
[![Docs](https://github.com/ipolyzos/git-read-mcp/actions/workflows/docs.yml/badge.svg)](https://github.com/ipolyzos/git-read-mcp/actions/workflows/docs.yml)

A **read-only** [Model Context Protocol](https://modelcontextprotocol.io/) (MCP) server over **stdio** for inspecting Git history and diffs safely from AI tooling.

## What it does

The server exposes tools to **read** repository state only, for example:

- **Compare revisions** — view diffs between commits, branches, or working tree (scope to be defined per tool).
- **Inspect changes** — summarize or fetch patch text suitable for review workflows.

No writes, no branch manipulation, and no destructive Git operations are in scope.

## Safety constraints

This project is designed to minimize blast radius when embedded in agents:

| Constraint | Intent |
|------------|--------|
| **No writes** | No `git commit`, `push`, `reset --hard`, index/worktree mutations, or hooks that alter state. |
| **No network** | Git subprocesses must not fetch, push, or otherwise use remotes; allowlists exclude networked commands. |
| **Allowlisted args** | Only an explicit set of `git` subcommands and flags is permitted; unknown argv is rejected. |
| **Timeouts** | Every subprocess has a bounded runtime so a hung `git` cannot block the host indefinitely. |
| **Output caps** | Diff and stdout/stderr are truncated or rejected when exceeding configured byte/line limits. |

Details will live in `docs/security.md` as behavior is implemented.

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

## License

Apache-2.0 — see [LICENSE](LICENSE).
