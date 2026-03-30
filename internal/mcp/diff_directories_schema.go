package mcp

import (
	"fmt"
	"strings"

	"github.com/ipolyzos/git-read-mcp/internal/validate"
)

// ToolNameDiffDirectories is the MCP tool identifier for directory tree diffs.
const ToolNameDiffDirectories = "diff_directories"

// DiffDirectoriesInput is the JSON arguments for diff_directories.
type DiffDirectoriesInput struct {
	RepoPath         string `json:"repo_path" jsonschema:"required,Repository root used to resolve relative directory paths"`
	LeftDir          string `json:"left_dir" jsonschema:"required,Left directory path (relative to repo or absolute)"`
	RightDir         string `json:"right_dir" jsonschema:"required,Right directory path (relative to repo or absolute)"`
	ContextLines     int    `json:"context_lines" jsonschema:"Unified diff context lines (0-50)"`
	IgnoreWhitespace string `json:"ignore_whitespace,omitempty" jsonschema:"Optional: none, all, or change"`
}

// DiffDirectoriesOutput is returned as structured content for diff_directories.
type DiffDirectoriesOutput struct {
	Diff      string         `json:"diff"`
	Truncated bool           `json:"truncated"`
	HadDiff   bool           `json:"had_diff"`
	Stats     *DiffLineStats `json:"stats,omitempty"`
}

// ValidateAndResolveDiffDirectories validates inputs, normalizes repo_path, resolves directory
// operands with ResolvePath (relative paths use the repo root), does not enforce repo containment,
// and requires both sides to exist as directories.
func ValidateAndResolveDiffDirectories(in DiffDirectoriesInput) (repoAbs, leftAbs, rightAbs string, err error) {
	if err := validateOperandPath("left_dir", in.LeftDir); err != nil {
		return "", "", "", err
	}
	if err := validateOperandPath("right_dir", in.RightDir); err != nil {
		return "", "", "", err
	}
	if err := validate.IntInRange("context_lines", in.ContextLines, 0, 50); err != nil {
		return "", "", "", err
	}

	repoAbs, err = validate.NormalizeRepoPath(in.RepoPath)
	if err != nil {
		return "", "", "", err
	}

	leftAbs, err = validate.ResolvePath(repoAbs, in.LeftDir)
	if err != nil {
		return "", "", "", err
	}
	rightAbs, err = validate.ResolvePath(repoAbs, in.RightDir)
	if err != nil {
		return "", "", "", err
	}

	if err := validate.EnsurePathExists(leftAbs, true); err != nil {
		return "", "", "", err
	}
	if err := validate.EnsurePathExists(rightAbs, true); err != nil {
		return "", "", "", err
	}

	if _, err := ParseIgnoreWhitespace(in.IgnoreWhitespace); err != nil {
		return "", "", "", err
	}

	return repoAbs, leftAbs, rightAbs, nil
}

func validateOperandPath(field, p string) error {
	s := strings.TrimSpace(p)
	if s == "" {
		return fmt.Errorf("%s is empty", field)
	}
	if strings.HasPrefix(s, "-") {
		return fmt.Errorf("%s must not begin with '-'", field)
	}
	return nil
}
