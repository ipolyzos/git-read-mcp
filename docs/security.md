# Security model

## Subprocess execution

Git operations run via `internal/execx` with the following hardening:

| Rule | Rationale |
|------|-----------|
| **No shell** | Commands are spawned with `exec.CommandContext` using a resolved binary path and `[]string` args only—no `sh -c`, so injection through metacharacters is avoided. |
| **Caller-set timeout** | Each run uses a `context.Context` and an optional wall-clock bound. The context is passed to `CommandContext`, so hung children are killed when the deadline passes (`TimedOut` on the result). |
| **Combined output cap** | stdout and stderr share one byte budget. Once exceeded, additional bytes are discarded and `Truncated` is set so callers never unboundedly grow memory from hostile commands. |
| **Environment and cwd** | Callers pass explicit `env` and `cwd`. Passing `nil` env inherits the host process environment; an empty slice starts the child with no inherited variables (use intentionally when minimizing leakage). |
| **Error mapping** | Timeouts surface as `errors.ErrExecTimeout`; truncation is visible on the result and via `errors.FromExecResult`; non-zero exits are surfaced with `errors.NonZeroExit` / `errors.FromExecResult` for consistent logging and API responses. |

These rules apply to the future Git adapter: subprocess arguments must remain allowlisted at a higher layer (`internal/git` + `internal/validate`); `execx` enforces resource limits only.

## Input and path validation

`internal/validate` enforces repository scoping and basic input size/range checks before any Git subprocess runs:

| Control | Behavior |
|---------|----------|
| **Repository root** | `NormalizeRepoPath` requires `repo_path` to exist as a directory (symlink targets are canonicalized when readable). |
| **Resolution** | `ResolvePath` joins relative user paths to the normalized repo root; absolute inputs are accepted as-is after `filepath.Clean`, so callers must apply the right “within repo” policy afterward. |
| **Traversal** | `EnsureWithinRepo` rejects paths whose cleaned absolute form escapes the repo root (string prefix / `filepath.Rel` checks). |
| **Existence** | `EnsurePathExists` validates file vs directory expectations for paths that must be present on disk (e.g. directory diff inputs). |
| **Limits** | `MaxStringBytes`, `IntInRange`, and `BoolDefault` centralize caps and defaults for tool arguments. |

Path comparisons do not follow symlinks inside the repository when checking containment; a symlink fully inside the repo that points elsewhere may bypass naive prefix checks when tools open the path. Treat symlink-heavy trees as higher risk; tightening that belongs with future path canonicalization work.

## Future work

- Path and argument allowlists for `git` subcommands and flags.
- Optional symlink-expanded containment checks for high-risk tools.
- Redaction or caps on error text returned to MCP clients.
