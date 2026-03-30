package mcp

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestValidateSummarizeDiffInput_inlineTooLarge(t *testing.T) {
	in := SummarizeDiffInput{Diff: strings.Repeat("a", 100), MaxDiffBytes: 50}
	err := ValidateSummarizeDiffInput(in)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestValidateSummarizeDiffInput_neitherMode(t *testing.T) {
	err := ValidateSummarizeDiffInput(SummarizeDiffInput{})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestValidateSummarizeDiffInput_pathsOK(t *testing.T) {
	repo := t.TempDir()
	if err := os.WriteFile(filepath.Join(repo, "l.txt"), []byte("a"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(repo, "r.txt"), []byte("b"), 0o644); err != nil {
		t.Fatal(err)
	}
	err := ValidateSummarizeDiffInput(SummarizeDiffInput{
		RepoPath:  repo,
		LeftPath:  "l.txt",
		RightPath: "r.txt",
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestSummarizeDiffOptions_caps(t *testing.T) {
	o := SummarizeDiffOptions(SummarizeDiffInput{
		MaxDiffBytes: 999_999_999,
		MaxFiles:     999_999,
		MaxHunks:     999_999,
	})
	if o.MaxDiffBytes != hardSummarizeMaxDiffBytes {
		t.Errorf("diff bytes cap: %d", o.MaxDiffBytes)
	}
	if o.MaxFiles != hardSummarizeMaxFiles {
		t.Errorf("files cap: %d", o.MaxFiles)
	}
	if o.MaxHunks != hardSummarizeMaxHunks {
		t.Errorf("hunks cap: %d", o.MaxHunks)
	}
}
