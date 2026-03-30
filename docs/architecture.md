# Architecture overview

This document describes how **git-read-mcp** is structured: stdio MCP transport, tool registration, validation, subprocess execution, and the read-only Git adapter. For deployment and operator issues, see [troubleshooting.md](troubleshooting.md) and [security.md](security.md).

## System context (stdio)

The server speaks **Model Context Protocol** over the process **stdin** and **stdout** pair. Nothing else may write to **stdout** or the MCP session corrupts.

```
                    ┌────────────────────────────────────────┐
  stdin (JSON-RPC)  │            git-read-mcp process         │  stdout (JSON-RPC)
 ──────────────────►│  ┌──────────────────────────────────┐  ├──────────────────►  MCP client
                    │  │ github.com/modelcontextprotocol/ │  │
                    │  │ go-sdk/mcp  +  StdioTransport    │  │
                    │  └──────────────────────────────────┘  │
                    │                    │                    │
                    │                    ▼                    │
                    │  tools / handlers / git / execx       │
                    │                    │                    │
                    │                    ▼                    │
                    │            os.Stderr only              ├──► logs (stdlib log,
                    └────────────────────────────────────────┘    structured JSON)
```

**Entrypoint** (`cmd/server/main.go`):

1. **stderr** — `log` package prefix `git-read-mcp:`; structured tool errors via `internal/logging` (JSON lines).
2. **config** — `config.LoadFromEnv()` for diff timeouts and output caps (see [troubleshooting.md](troubleshooting.md)).
3. **Lifecycle** — `signal.NotifyContext` for clean shutdown; `mcp.RunStdio` blocks until cancel or transport error.

`RunStdio` (**`internal/mcp/server.go`**) constructs `mcpsdk.NewServer`, installs tools from an **`Registry`** (optional extra tools), runs **`handlers.InstallBuiltins`** to attach built-in tools via the SDK’s generic **`AddTool`**, then **`srv.Run(ctx, &mcpsdk.StdioTransport{})`**.

## Tool registry and handlers

Two complementary mechanisms exist:

| Mechanism | Role |
|-----------|------|
| **`mcp.Registry`** | In-process map of **`ToolMeta` + `ToolHandler`**; **`InstallTools`** registers each with **`Server.AddTool`** (low-level handler). Used when you extend the server with custom tools before **`RunStdio`**. |
| **`handlers.InstallBuiltins`** | Calls **`mcpsdk.AddTool`** with typed **`ToolHandlerFor`** implementations: `diff_files`, `diff_directories`, `summarize_diff`, `compare_text`. Input schema, unmarshaling, and structured output are handled by the SDK. |

```
  Registry (optional)          Built-ins installer
  ┌─────────────────┐           ┌──────────────────────────┐
  │ Register(meta,  │           │ AddTool + HandleDiffFiles│
  │   handler)      │           │ AddTool + HandleDiff…  │
  └────────┬────────┘           └────────────┬─────────────┘
           │                                 │
           └──────────────┬──────────────────┘
                          ▼
                 mcpsdk.Server (tool dispatch)
                          │
                          ▼
              Handler: validate → git/execx → result
                          │
                     error path
                          ▼
              mapToolError → errors.Map + logging.ToolFailure
```

**Failure path:** handlers return Go **`error`**; the SDK turns that into a tool result with **`isError: true`**. **`mapToolError`** (**`internal/mcp/handlers/errmap.go`**) classifies via **`internal/errors.Map`** (stable JSON **`code` / `message`**) and emits one structured line to **stderr** so operators can correlate without polluting stdout.

**Success with partial output:** e.g. **`git diff`** output capped — **`Truncated: true`** on the structured payload and **no** tool-level error when the handler treats **`ErrExecTruncated`** as a soft condition (see diff handlers).

## Validation layer and path safety model

**Package:** **`internal/validate`**.

Typical flow for repo-scoped tools:

```
  repo_path  ──► NormalizeRepoPath  (exists, is dir, optional symlink canonicalization)
                    │
  user path  ──► ResolvePath(repoAbs, path)   (relative → under repo root join; absolute → cleaned as-is)
                    │
                    ├──► EnsureWithinRepo(repoAbs, resolved)   [tools that require containment]
                    └──► EnsurePathExists(resolved, file|dir)
```

| Concern | Behavior |
|---------|----------|
| **Repository root** | Must exist as a directory; errors surface as **`validate:`**-prefixed messages. |
| **Relative paths** | Joined to normalized **`repo_path`** via **`filepath.Join`** after **`filepath.Clean`**. |
| **Absolute paths** | Cleaned only; **containment is not automatic** — callers must invoke **`EnsureWithinRepo`** when policy requires it. |
| **Traversal** | **`EnsureWithinRepo`** uses **`filepath.Rel`**; **`..`** or prefix **`..`** + separator ⇒ **`ErrOutsideRepo`**. |
| **Symlinks** | Repo root may be canonicalized via **`EvalSymlinks`** when readable. Inner symlink targets are **not** fully canonicalized for containment checks (see **`security.md`**). |
| **Scalar limits** | **`IntInRange`**, **`MaxStringBytes`**, etc., back MCP schema ranges and oversized blobs. |

**Tool-specific policy (important):**

- **`diff_files`**, **`summarize_diff`** (path mode): both operands **`EnsureWithinRepo`** + files must exist.
- **`diff_directories`**: **`NormalizeRepoPath`** and **`ResolvePath`** for operands, **no** **`EnsureWithinRepo`** — absolute directory paths may lie outside the repo by design; both sides must exist as **directories**.
- **`compare_text`**: **`NormalizeRepoPath`** for context only; content is written under **`os.MkdirTemp`**, not under **`repo_path`**.

## Subprocess runner (`internal/execx`)

**`execx.Run(ctx, cmd, args, env, cwd, timeout, maxBytes)`**

- **No shell** — **`exec.CommandContext` with argv slice**; **`cmd`** is the binary name (e.g. **`git`**).
- **Timeout** — optional wall clock via **`context.WithTimeout`**; **`DeadlineExceeded`** ⇒ result **`TimedOut`**, error **`ErrExecTimeout`**.
- **Combined output cap** — **`maxBytes`** applies to **stdout + stderr together** (shared counter). Overrun sets **`Truncated`** and does not grow buffers without bound.
- **Exit codes** — **`0`** returned as success with **`ExitCode` set**; non-zero **`exec.ExitError`** ⇒ **`ExitCode`** populated, **`error` nil** unless timeout/truncation/copy failure (see package doc).

```
  execx.Run
       │
       ├─► context (parent + optional timeout)
       ├─► io.Copy from stdout/stderr into capped buffers
       ├─► wait → ExitCode, TimedOut, Truncated
       └─► git package maps to DiffNoIndexResult / NonZeroExit / ErrExecTruncated
```

Defaults for diff-sized work are tuned in **`internal/git`** and **`internal/config`** (**`GIT_READ_MCP_DIFF_TIMEOUT`**, **`GIT_READ_MCP_MAX_OUTPUT_BYTES`**); **`DetectGit`** uses shorter timeout and a small cap for **`git --version`**.

## Git adapter: allowlists and environment

**Package:** **`internal/git`**.

### Environment (`EnvGitReadOnly`)

Built from **`os.Environ()`** merged with fixed overrides (no shell, literals only):

| Variable | Value | Why |
|----------|-------|-----|
| **`GIT_PAGER`**, **`PAGER`** | `cat` | Avoid TTY pagers when stdio is the MCP wire. |
| **`GIT_TERMINAL_PROMPT`** | `0` | No interactive credential prompts. |
| **`LC_ALL`** | `C` | Predictable locale for diff text. |

### Allowlisted invocation (`DiffNoIndex`)

User-controlled strings never become **flag tokens**. Arguments are built only from:

1. Literal **`git`** subcommand **`diff`**, **`--no-index`**.
2. **`--unified=<n>`** with **`n`** from bounded integer handling (default **3** if unset in options; max **200**).
3. Optional **`--ignore-all-space`** or **`--ignore-space-change`** from enum **`IgnoreWhitespace`** (not raw strings).
4. Literal **`--`**, then **left** and **right** path operands.

Path operands are validated: non-empty, must not start with **`-`**.

```
  argv:  git  diff  --no-index  --unified=<n>  [ignore flags]  --  <left>  <right>
         │                    │                      │              └── only after '--'
         └────────────────────┴──────────────────────┘
                    Allowlisted prefixes only
```

### Exit semantics

| Exit code | Interpretation |
|-----------|------------------|
| **0** | No textual difference; **`HadDiff: false`**. |
| **1** | Differences present; **`HadDiff: true`** (normal for `diff`). |
| **Other** | Failure; **`NonZeroExit`** with stderr payload. |

Truncation of captured output yields **`Truncated: true`** and **`ErrExecTruncated`** wrapped for the caller to downgrade to a successful tool response when appropriate.

---

## Related packages (quick map)

| Package | Responsibility |
|---------|----------------|
| **`internal/mcp`** | **`RunStdio`**, **`Registry`**, **`InstallTools`**, tool I/O types, envelopes. |
| **`internal/mcp/handlers`** | Builtin tool bodies; **`mapToolError`**. |
| **`internal/validate`** | Paths, repo normalization, containment, existence, numeric/string caps. |
| **`internal/execx`** | Bounded subprocess runs. |
| **`internal/git`** | **`DiffNoIndex`**, **`DetectGit`**, **`EnvGitReadOnly`**. |
| **`internal/errors`** | **`Map`**, coded errors for MCP text JSON. |
| **`internal/config`** | Env-based timeouts and output caps consumed by **`git`**. |
| **`internal/logging`** | Structured stderr lines for tool failures. |

For subprocess and validation rules in prose aimed at reviewers, **`docs/security.md`** complements this file.
