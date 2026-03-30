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

## Future work

- Path and argument allowlists for `git` subcommands and flags.
- Redaction or caps on error text returned to MCP clients.
