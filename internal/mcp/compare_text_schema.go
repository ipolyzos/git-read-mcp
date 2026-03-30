package mcp

import (
	"github.com/ipolyzos/git-read-mcp/internal/validate"
)

// ToolNameCompareText is the MCP tool identifier for in-memory text comparison.
const ToolNameCompareText = "compare_text"

// MaxCompareTextBytes is the maximum UTF-8 byte length allowed per text operand.
const MaxCompareTextBytes = 200 * 1024

// CompareTextInput is the JSON arguments for compare_text.
type CompareTextInput struct {
	RepoPath         string `json:"repo_path" jsonschema:"required,Existing directory; used for context only—no writes"`
	LeftText         string `json:"left_text" jsonschema:"required,Left text blob"`
	RightText        string `json:"right_text" jsonschema:"required,Right text blob"`
	ContextLines     int    `json:"context_lines" jsonschema:"Unified diff context lines (0-50)"`
	IgnoreWhitespace string `json:"ignore_whitespace,omitempty" jsonschema:"Optional: none, all, or change"`
}

// CompareTextOutput is returned as structured content for compare_text.
type CompareTextOutput struct {
	Diff      string         `json:"diff"`
	Truncated bool           `json:"truncated"`
	HadDiff   bool           `json:"had_diff"`
	Stats     *DiffLineStats `json:"stats,omitempty"`
}

// ValidateCompareTextInput checks repo_path, string size caps, context lines, and whitespace mode.
// It does not write to repo_path.
func ValidateCompareTextInput(in CompareTextInput) error {
	if err := validate.MaxStringBytes("left_text", in.LeftText, MaxCompareTextBytes); err != nil {
		return err
	}
	if err := validate.MaxStringBytes("right_text", in.RightText, MaxCompareTextBytes); err != nil {
		return err
	}
	if err := validate.IntInRange("context_lines", in.ContextLines, 0, 50); err != nil {
		return err
	}
	if _, err := ParseIgnoreWhitespace(in.IgnoreWhitespace); err != nil {
		return err
	}
	if _, err := validate.NormalizeRepoPath(in.RepoPath); err != nil {
		return err
	}
	return nil
}
