# Troubleshooting

This MCP server uses **stdout for the protocol only**. Diagnostics and structured errors from the server process go to **stderr**. If stdout is redirected or consumed by anything other than the MCP client, the session will break.

## Environment tuning (optional)

| Variable | Meaning | Default |
|----------|---------|---------|
| `GIT_READ_MCP_DIFF_TIMEOUT` | Max wall time for each `git diff --no-index` (and similar) invocation | `30s` |
| `GIT_READ_MCP_MAX_OUTPUT_BYTES` | Max combined bytes captured from git stdout+stderr per diff | `8388608` (8 MiB) |
| `GIT_READ_MCP_GIT_PROBE_TIMEOUT` | Timeout for `git --version` probes | `5s` |
| `GIT_READ_MCP_GIT_PROBE_MAX_BYTES` | Byte cap for probe output | `65536` |

Invalid values cause the process to **exit on startup** with `config: ...` on stderr. Durations use Go’s `time.ParseDuration` form (e.g. `45s`, `2m`).

## Git not found or failing to start

**Symptoms:** tool result with `code` `GIT_ERROR` and a message about the git executable; stderr may show `tool_failure` JSON lines with the same code.

**Checks:**

- `git` is on `PATH` for the MCP server process (IDE/agent environment may differ from your shell).
- Run `git --version` in the same environment as the server.

## Invalid repository or paths

**Symptoms:** `INVALID_ARGUMENT` or `NOT_FOUND` in the tool error JSON; messages often prefixed with `validate:`.

**Checks:**

- `repo_path` exists and is a **directory**.
- For `diff_files` / `summarize_diff` (path mode): `left_path` and `right_path` resolve **inside** the repo and point to **existing files**.
- Paths are not empty and do not start with `-` (rejected as operands).

## Truncated diff output

**Symptoms:** structured result has `truncated: true` (or `git_diff_truncated` for summaries); diff text may end abruptly.

**Meaning:** Output hit `GIT_READ_MCP_MAX_OUTPUT_BYTES` (or an internal test override). This is **not** a tool-level error: the call succeeded with partial output.

**Mitigation:** Increase `GIT_READ_MCP_MAX_OUTPUT_BYTES`, narrow the diff (smaller files/trees, fewer context lines), or use a more targeted tool.

## Timeouts

**Symptoms:** `TIMEOUT` in the tool error JSON; stderr may log `tool_failure` with code `TIMEOUT`.

**Meaning:** The git subprocess exceeded `GIT_READ_MCP_DIFF_TIMEOUT`, or the request context was canceled.

**Mitigation:** Increase `GIT_READ_MCP_DIFF_TIMEOUT`, or reduce work (smaller trees, less data).

## Tool error JSON shape

Failures returned from tools use a single JSON object in the text content, for example:

```json
{"code":"INVALID_ARGUMENT","message":"validate: ..."}
```

Stable `code` values include: `INVALID_ARGUMENT`, `NOT_FOUND`, `TIMEOUT`, `TOO_LARGE`, `GIT_ERROR`, `INTERNAL`.

## Structured stderr logs

On handler failures the server may emit one JSON line to stderr, e.g. `{"level":"error","msg":"tool_failure","tool":"diff_files","code":"NOT_FOUND","message":"..."}`. Use these for operators; clients should rely on the tool result payload above.
