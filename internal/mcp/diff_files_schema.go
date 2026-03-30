package mcp

import (
	"fmt"
	"strings"

	"github.com/ipolyzos/git-read-mcp/internal/git"
	"github.com/ipolyzos/git-read-mcp/internal/validate"
)

// ToolNameDiffFiles is the MCP tool identifier for file diffs.
const ToolNameDiffFiles = "diff_files"

// DiffFilesInput is the JSON arguments for diff_files.
type DiffFilesInput struct {
	RepoPath         string `json:"repo_path" jsonschema:"required,Filesystem path to the repository root"`
	LeftPath         string `json:"left_path" jsonschema:"required,Left file path (relative to repo or absolute)"`
	RightPath        string `json:"right_path" jsonschema:"required,Right file path (relative to repo or absolute)"`
	ContextLines     int    `json:"context_lines" jsonschema:"Unified diff context lines (0-50)"`
	IgnoreWhitespace string `json:"ignore_whitespace,omitempty" jsonschema:"Optional: none, all, or change"`
}

// DiffFilesOutput is returned as structured content for diff_files.
type DiffFilesOutput struct {
	Diff      string         `json:"diff"`
	Truncated bool           `json:"truncated"`
	HadDiff   bool           `json:"had_diff"`
	Stats     *DiffLineStats `json:"stats,omitempty"`
}

// DiffLineStats summarizes +/- lines in unified diff text (excluding headers).
type DiffLineStats struct {
	Insertions int `json:"insertions"`
	Deletions  int `json:"deletions"`
}

// ParseIgnoreWhitespace maps tool input to git ignore mode.
func ParseIgnoreWhitespace(s string) (git.IgnoreWhitespace, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "", "none":
		return git.IgnoreWhitespaceNone, nil
	case "all":
		return git.IgnoreWhitespaceAll, nil
	case "change":
		return git.IgnoreWhitespaceChange, nil
	default:
		return 0, fmt.Errorf("ignore_whitespace: want none, all, or change, got %q", s)
	}
}

// ValidateAndResolveDiffFiles validates scalar inputs, resolves paths under repo_path,
// enforces repo containment, and checks both files exist.
func ValidateAndResolveDiffFiles(in DiffFilesInput) (repoAbs, leftAbs, rightAbs string, err error) {
	if err := validate.IntInRange("context_lines", in.ContextLines, 0, 50); err != nil {
		return "", "", "", err
	}
	repoAbs, err = validate.NormalizeRepoPath(in.RepoPath)
	if err != nil {
		return "", "", "", err
	}
	leftAbs, err = validate.ResolvePath(repoAbs, in.LeftPath)
	if err != nil {
		return "", "", "", err
	}
	if err := validate.EnsureWithinRepo(repoAbs, leftAbs); err != nil {
		return "", "", "", err
	}
	rightAbs, err = validate.ResolvePath(repoAbs, in.RightPath)
	if err != nil {
		return "", "", "", err
	}
	if err := validate.EnsureWithinRepo(repoAbs, rightAbs); err != nil {
		return "", "", "", err
	}
	if err := validate.EnsurePathExists(leftAbs, false); err != nil {
		return "", "", "", err
	}
	if err := validate.EnsurePathExists(rightAbs, false); err != nil {
		return "", "", "", err
	}
	return repoAbs, leftAbs, rightAbs, nil
}
