# Release process

This document describes how maintainers cut releases for **git-read-mcp**. For user-facing change history, see [CHANGELOG.md](../CHANGELOG.md) at the repository root.

## Branch flow: `develop` → `main`

1. **Integration work** lands on **`develop`** (pull requests, CI green).
2. When you are ready to ship a stable line, merge **`develop` into `main`** (via PR or merge commit, per team preference).  
   - `main` should always reflect what you consider **release-ready**.  
   - Resolve conflicts on the integration PR; run `go test ./...` locally if needed.
3. **Release tags** are created from **`main`** (see below), not from unmerged feature branches.

Hotfixes can branch from `main`, merge back to both `main` and `develop`.

## Version bump

1. Decide the next version using [Semantic Versioning](https://semver.org/) (`MAJOR.MINOR.PATCH`, optional pre-release suffixes like `-rc.1`).
2. Update **[CHANGELOG.md](../CHANGELOG.md)**:
   - Move items under **`[Unreleased]`** into a new section **`[x.y.z] - YYYY-MM-DD`**.
   - Leave a fresh **`[Unreleased]`** section at the top for future work.
3. **Embedded binary version** (optional for dev builds):
   - Release **GitHub Actions** builds set the version at link time (`releaseVersion` in `cmd/server`; see the release workflow).
   - For ad-hoc builds you can set **`GIT_READ_MCP_VERSION`** in the environment instead of changing code.

There is no separate `VERSION` file; the tag and changelog are the source of truth.

## Tagging

1. On **`main`**, with changelog and any version-sensitive edits committed:
   ```bash
   git checkout main
   git pull
   git tag -a v0.1.0 -m "Release v0.1.0"
   git push origin v0.1.0
   ```
2. Use **annotated tags** (`-a`) so tooling and humans see message + author/date.
3. Tag names must match **`v*`** with a normal semver after `v` (e.g. `v1.2.3`, `v2.0.0-rc.1`).

Pushing the tag triggers the **Release** workflow (`.github/workflows/release.yml`), which builds cross-platform archives, computes checksums, and creates a GitHub Release with attached assets.

## Release notes

1. **CHANGELOG.md** is the canonical list of user-visible changes per version.
2. The GitHub **Release** title can mirror the tag (`v0.1.0`). The body should briefly summarize highlights and point readers to the matching **CHANGELOG** section.
3. Attachments are produced by CI: per-OS archives plus **`SHA256SUMS`**. Verifiers can run `sha256sum -c SHA256SUMS` (or `shasum -a 256 -c SHA256SUMS` on macOS) after downloading.

## Reproducible builds

Release binaries are built with:

- **`CGO_ENABLED=0`** — static, portable executables without a C toolchain.
- **`-trimpath`** — strips local filesystem paths from the binary.
- **`-buildvcs=false`** — avoids embedding VCS stamps that can vary between checkouts.
- **`-ldflags="-s -w"`** — smaller binaries (no separate symbol/table requirement for end users).

The same flags (except optional `-s -w`) are recommended in the [Makefile](../Makefile) for local release-like builds. Identical inputs (same commit, Go version, modules) should yield identical outputs; CI uses a fixed Go version from the workflow file.
