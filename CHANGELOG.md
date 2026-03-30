# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- New MCP tool `summarize_diff` with structured diff summaries (files, insertions, deletions, hunks) from either inline unified diff input or computed file-path diffs.
- Full built-in registration for `compare_text` in server tool installation.
- Release workflow for tag-based multi-platform binaries and checksum publication.
- CI docs workflow with markdown link validation.
- Architecture, security, troubleshooting, release-process, and expanded tools documentation.

### Changed
- CI quality gates now include `gofmt` checks, `go vet`, `go test`, and `govulncheck`.
- Build defaults aligned for reproducibility with `CGO_ENABLED=0`, `-trimpath`, and `-buildvcs=false`.
- `README.md` expanded with MCP usage examples, sample responses, limitations, and CI/docs badges.

### Fixed
- Standardized structured error mapping across handlers with stable error codes: `INVALID_ARGUMENT`, `NOT_FOUND`, `TIMEOUT`, `TOO_LARGE`, `GIT_ERROR`, and `INTERNAL`.
- Logging hardened to structured `stderr` output so protocol traffic remains on `stdout`.
- Diff/output truncation surfaced explicitly in handler responses and tests.
