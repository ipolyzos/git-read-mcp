package mcp

import (
	"fmt"
	"strings"

	"github.com/ipolyzos/git-read-mcp/internal/mcp/diffparse"
	"github.com/ipolyzos/git-read-mcp/internal/validate"
)

// ToolNameSummarizeDiff is the MCP tool name for diff summarization.
const ToolNameSummarizeDiff = "summarize_diff"

const (
	defaultSummarizeMaxDiffBytes = 2 << 20
	defaultSummarizeMaxFiles     = 500
	defaultSummarizeMaxHunks     = 10000
	hardSummarizeMaxDiffBytes    = 8 << 20
	hardSummarizeMaxFiles        = 10000
	hardSummarizeMaxHunks        = 100000
)

// SummarizeDiffInput accepts an inline unified diff or two file paths (same rules as diff_files).
type SummarizeDiffInput struct {
	Diff             string `json:"diff,omitempty" jsonschema:"Optional inline unified diff; if set, paths are ignored"`
	RepoPath         string `json:"repo_path,omitempty" jsonschema:"Repository root when computing diff from paths"`
	LeftPath         string `json:"left_path,omitempty" jsonschema:"Left file when computing diff from paths"`
	RightPath        string `json:"right_path,omitempty" jsonschema:"Right file when computing diff from paths"`
	ContextLines     int    `json:"context_lines,omitempty" jsonschema:"Unified context when computing diff (0-50)"`
	IgnoreWhitespace string `json:"ignore_whitespace,omitempty" jsonschema:"When computing diff: none, all, or change"`

	MaxDiffBytes int `json:"max_diff_bytes,omitempty" jsonschema:"Cap diff bytes fed to the parser (default 2MiB, max 8MiB)"`
	MaxFiles     int `json:"max_files,omitempty" jsonschema:"Stop after this many files (default 500)"`
	MaxHunks     int `json:"max_hunks,omitempty" jsonschema:"Stop after this many hunks total (default 10000)"`
}

// SummarizeDiffOutput contains a short narrative plus structured stats.
type SummarizeDiffOutput struct {
	Summary          string           `json:"summary"`
	Files            []SummarizedFile `json:"files"`
	Totals           SummarizeTotals  `json:"totals"`
	DiffSource       string           `json:"diff_source"`
	InputTruncated   bool             `json:"input_truncated"`
	ParseTruncated   bool             `json:"parse_truncated"`
	GitDiffTruncated bool             `json:"git_diff_truncated,omitempty"`
}

// SummarizedFile mirrors diffparse.FileChange for stable JSON output types.
type SummarizedFile struct {
	Path       string `json:"path"`
	OldPath    string `json:"old_path,omitempty"`
	Rename     bool   `json:"rename,omitempty"`
	Binary     bool   `json:"binary,omitempty"`
	Insertions int    `json:"insertions"`
	Deletions  int    `json:"deletions"`
	Hunks      int    `json:"hunks"`
}

// SummarizeTotals rolls up counts across files.
type SummarizeTotals struct {
	Files      int `json:"files"`
	Insertions int `json:"insertions"`
	Deletions  int `json:"deletions"`
	Hunks      int `json:"hunks"`
}

// SummarizeDiffOptions normalizes limit fields from tool input.
func SummarizeDiffOptions(in SummarizeDiffInput) diffparse.Options {
	mb := in.MaxDiffBytes
	if mb <= 0 {
		mb = defaultSummarizeMaxDiffBytes
	}
	if mb > hardSummarizeMaxDiffBytes {
		mb = hardSummarizeMaxDiffBytes
	}
	mf := in.MaxFiles
	if mf <= 0 {
		mf = defaultSummarizeMaxFiles
	}
	if mf > hardSummarizeMaxFiles {
		mf = hardSummarizeMaxFiles
	}
	mh := in.MaxHunks
	if mh <= 0 {
		mh = defaultSummarizeMaxHunks
	}
	if mh > hardSummarizeMaxHunks {
		mh = hardSummarizeMaxHunks
	}
	return diffparse.Options{MaxDiffBytes: mb, MaxFiles: mf, MaxHunks: mh}
}

// effectiveMaxDiffForInline returns the cap used to reject oversize inline diffs before parsing.
func effectiveMaxDiffForInline(in SummarizeDiffInput) int {
	mb := in.MaxDiffBytes
	if mb <= 0 {
		mb = defaultSummarizeMaxDiffBytes
	}
	if mb > hardSummarizeMaxDiffBytes {
		mb = hardSummarizeMaxDiffBytes
	}
	return mb
}

// ValidateSummarizeDiffInput verifies that exactly one mode is usable and bounds make sense.
func ValidateSummarizeDiffInput(in SummarizeDiffInput) error {
	hasDiff := strings.TrimSpace(in.Diff) != ""
	hasPaths := strings.TrimSpace(in.RepoPath) != "" && strings.TrimSpace(in.LeftPath) != "" && strings.TrimSpace(in.RightPath) != ""
	if !hasDiff && !hasPaths {
		return fmt.Errorf("summarize_diff: provide diff or repo_path+left_path+right_path")
	}
	if hasDiff {
		if lim := effectiveMaxDiffForInline(in); len(in.Diff) > lim {
			return fmt.Errorf("summarize_diff: diff length %d exceeds max_diff_bytes %d", len(in.Diff), lim)
		}
		return nil
	}

	if err := validate.IntInRange("context_lines", in.ContextLines, 0, 50); err != nil {
		return err
	}
	if _, err := ParseIgnoreWhitespace(in.IgnoreWhitespace); err != nil {
		return err
	}
	df := DiffFilesInput{
		RepoPath:         in.RepoPath,
		LeftPath:         in.LeftPath,
		RightPath:        in.RightPath,
		ContextLines:     in.ContextLines,
		IgnoreWhitespace: in.IgnoreWhitespace,
	}
	if _, _, _, err := ValidateAndResolveDiffFiles(df); err != nil {
		return err
	}
	return nil
}

// BuildSummarizeOutput maps a parse result into MCP output and narrative summary.
func BuildSummarizeOutput(parsed diffparse.Result, source string, gitTrunc bool) SummarizeDiffOutput {
	files := make([]SummarizedFile, 0, len(parsed.Files))
	for _, f := range parsed.Files {
		files = append(files, SummarizedFile{
			Path: f.Path, OldPath: f.OldPath, Rename: f.Rename, Binary: f.Binary,
			Insertions: f.Insertions, Deletions: f.Deletions, Hunks: f.Hunks,
		})
	}
	out := SummarizeDiffOutput{
		Summary:          formatSummarizeSummary(parsed, source, gitTrunc),
		Files:            files,
		DiffSource:       source,
		InputTruncated:   parsed.InputTruncated,
		ParseTruncated:   parsed.ParseTruncated,
		GitDiffTruncated: gitTrunc,
		Totals: SummarizeTotals{
			Files:      len(parsed.Files),
			Insertions: parsed.TotalInsertions,
			Deletions:  parsed.TotalDeletions,
			Hunks:      parsed.TotalHunks,
		},
	}
	return out
}

func formatSummarizeSummary(r diffparse.Result, source string, gitDiffTrunc bool) string {
	var b strings.Builder
	if len(r.Files) == 0 && r.TotalInsertions == 0 && r.TotalDeletions == 0 {
		b.WriteString("No file changes detected in the diff.")
	} else {
		b.WriteString(fmt.Sprintf("%d file(s), +%d / -%d lines, %d hunks (approx).",
			len(r.Files), r.TotalInsertions, r.TotalDeletions, r.TotalHunks))
	}
	if source == "computed" {
		b.WriteString(" Source: computed via git diff --no-index.")
	} else {
		b.WriteString(" Source: inline diff.")
	}
	if r.InputTruncated {
		b.WriteString(" Diff text was truncated to max_diff_bytes before parsing.")
	}
	if r.ParseTruncated {
		b.WriteString(" Parsing stopped early due to max_files or max_hunks.")
	}
	if gitDiffTrunc {
		b.WriteString(" Git subprocess output was truncated.")
	}
	return b.String()
}
