# Security model

This document summarizes **threats, assumptions, and mitigations** for git-read-mcp. For system structure and data flow, see [architecture.md](architecture.md). For operators (env tuning, stderr vs stdout), see [troubleshooting.md](troubleshooting.md).

## Threat model

### Trust boundaries

| Trusted | Untrusted |
|---------|-----------|
| MCP **host** (IDE, runner, OS) that starts the process and sets env | **Tool arguments** JSON from the client: paths, text blobs, numeric options, inline diffs |
| Local filesystem layout the host exposes to the process | Content read **through** paths the tools open (file/dir contents, git output) |
| `git` binary resolved on `PATH` | Network: server does not implement Git networking; subprocess env discourages prompts |

We assume the **client can invoke tools with arbitrary arguments** within schema bounds. Design goals:

1. **No shell injection** — user strings never become shell grammar.
2. **No flag injection** — user strings never become `git` flags; paths are operands after a literal `--`.
3. **Containment where promised** — repo-scoped tools cannot open paths outside `repo_path` (with documented symlink caveats).
4. **Bounded resources** — subprocess runtime and captured output are capped; parsers and large strings have limits.

The server is **read-only**: it does not commit, push, fetch, or mutate repos. Risk is primarily **disclosure** (read arbitrary allowed paths), **DoS** (CPU, memory, subprocess output), and **host interaction** via `git` behavior on pathological inputs.

## Deny shell interpolation

All external commands run through **`internal/execx`**:

- **`exec.CommandContext`** with **`Cmd` + `Args []string`** — **no** `/bin/sh -c`, pipes, or string concatenation into a shell.
- The first argument is the binary name (`git`); remaining entries are discrete argv words.
- **Metacharacters in paths** are not interpreted by a shell; they are passed as literal path strings to `git` and the kernel.

This removes an entire class of “`; rm -rf /`” style injection **at the spawn layer**. Remaining risk is **`git` itself** parsing odd paths or exotic repo features—mitigated by allowlisted argv and path validation.

## Allowlisted `git` arguments (per code path)

**Every** current invocation uses a fixed template; user input appears **only** as path operands **after** `--`, or not at all for `git --version`.

### `git diff --no-index` (all diff/compare tools)

Used by **`diff_files`**, **`diff_directories`**, **`summarize_diff`** (path mode), and **`compare_text`**. Constructed in **`internal/git`** via `buildDiffNoIndexArgs`:

| Position | Value | User-controlled? |
|----------|-------|------------------|
| `argv[0]` | `git` (implicit via `execx.Run("git", args, ...)`) | No |
| `argv[1..]` | `diff`, `--no-index` | No |
| next | `--unified=<n>` where `n` ∈ [0, 200] from tool `context_lines` / defaults | Bounded integer only |
| optional | `--ignore-all-space` or `--ignore-space-change` | From allowlisted enum (`none` / `all` / `change`), not raw strings |
| separator | `--` | Literal |
| operands | `left`, `right` absolute paths | **Yes** — must pass `validatePathArg`: non-empty, must not start with `-` |

So **no** user field selects subcommands, refspecs, config keys, or `-c` overrides.

### `git --version` (probe)

**`DetectGit`**: `git` with single argument `--version`. No user paths. Short timeout and small output cap.

### Tools that never invoke `git`

- **`summarize_diff`** with a non-empty inline **`diff`** string: parsing only (**`internal/mcp/diffparse`**), subject to byte/file/hunk caps (see below).

## Repository path constraints and traversal prevention

### `repo_path`

- Must exist as a **directory** after `filepath.Abs` / `Clean` (and optional symlink canonicalization for the **root** via `filepath.EvalSymlinks` when readable).
- Empty or missing roots are rejected before operands are resolved.

### Relative resolution

- **`ResolvePath(repoAbs, userPath)`**: relative paths join under **`repoAbs`** after `Clean`; absolute paths are only **`Clean`ed**—**containment is not automatic** until a tool calls **`EnsureWithinRepo`**.

### Containment policy by tool

| Tool | `EnsureWithinRepo` | Notes |
|------|-------------------|--------|
| **`diff_files`**, **`summarize_diff`** (path mode) | **Yes** for both file operands | Files must exist; traversal via `..` in relative segments is rejected. |
| **`diff_directories`** | **No** | **By design**: absolute `left_dir` / `right_dir` may point outside `repo_path`. Host must only connect clients it trusts with that capability. |
| **`compare_text`** | N/A for operands | Only **`NormalizeRepoPath`** on `repo_path` (context); blobs go to **temp** paths (see below). |

### Symlinks and containment

**`EnsureWithinRepo`** compares **cleaned absolute** strings and `filepath.Rel`; it does **not** recursively resolve symlinks **inside** the repo for every check. A symlink **inside** the repo that targets an **outside** path could still be opened when `git` or the OS follows it. Treat **repo trees heavy with symlinks** as higher exposure until future canonicalization is added.

## Output truncation and DoS considerations

### Subprocess (`execx` + `git`)

- **Wall-clock timeout** (per run): from **`internal/config`** / defaults (see **`GIT_READ_MCP_DIFF_TIMEOUT`** in [troubleshooting.md](troubleshooting.md)). Stops runaway `git` or huge trees.
- **Combined stdout+stderr byte cap**: shared budget; when exceeded, **`Truncated`** is set and extra bytes are dropped—**memory** use for capture stays bounded.
- **Exit semantics**: `git diff` exit `1` means “diff present,” not failure. Other non-zero exits surface as **`NonZeroExit`** with stderr capped in client-facing error mapping.

Large **responses** still consume **MCP client/host** memory when structured content is huge; truncation flags indicate partial diff text.

### Parsing and inline diffs (`summarize_diff`)

- **Inline `diff`**: length rejected if over effective **`max_diff_bytes`** before parse; otherwise input may be **sliced** with **`input_truncated`**.
- **Parser work**: **`max_files`** and **`max_hunks`** stop scanning early with **`parse_truncated`**.
- **`strings.Split` on lines**: bounded by **`MaxDiffBytes`** slice first—pathological line counts are mitigated but CPU scales with input size within cap.

### Other caps

- **`compare_text`**: each of **`left_text`** / **`right_text`** capped (**200 KiB** per operand) before write.
- **Context lines**: **0–50**; unified width in git argv bounded (**≤ 200** in **`internal/git`**).
- **Tool errors**: messages trimmed in **`internal/errors.Map`** for stable JSON responses (see [troubleshooting.md](troubleshooting.md)).

### Residual DoS

- Extremely **large allowed files** can still make `git diff` CPU-heavy within the timeout.
- **`diff_directories`** can span **arbitrary** directory trees if absolute paths are used—operational trust matters.

## Temporary files (`compare_text`)

The **`compare_text`** handler (**`internal/mcp/handlers/compare_text.go`**) materializes operands on disk for `git diff --no-index`. It is registered in **`InstallBuiltins`**; the same properties apply:

| Topic | Mitigation |
|-------|------------|
| **Location** | **`os.MkdirTemp`** with prefix `git-read-mcp-compare-*` under the default temp dir (or a **test-only** parent). |
| **Permissions** | Directory created by the OS defaults (e.g. `0700` on Unix for `MkdirTemp`). |
| **Path injection** | Only fixed names **`left.txt`** and **`right.txt`** under that directory; payload is **file contents**, not path segments. |
| **Repo writes** | **Nothing** is written under **`repo_path`**; repo is validated only for caller context. |
| **Size** | **`MaxCompareTextBytes`** per operand before filesystem write. |
| **Cleanup** | **`os.RemoveAll`** in **`defer`** after the diff (success or most failures). Process kill may leave garbage in temp until the OS reclaims it. |

## Git child environment (read-only posture)

**`EnvGitReadOnly`** merges host env with fixed overrides: **`GIT_PAGER`/`PAGER=cat`**, **`GIT_TERMINAL_PROMPT=0`**, **`LC_ALL=C`**. This reduces interactive prompts and stabilizes output; it is **not** a substitute for argv allowlisting.

## Related documentation

- [architecture.md](architecture.md) — layering, registry, validation, execx, git adapter.
- [tools.md](tools.md) — per-tool behavior and safety notes.

## Future hardening (non-exhaustive)

- Optional **symlink-expanded** containment for high-assurance deployments.
- Stronger **redaction** of stderr in **`GIT_ERROR`** responses when git emits long or sensitive diagnostics.
- Narrower **`diff_directories`** path policy if a stricter default becomes desirable.
