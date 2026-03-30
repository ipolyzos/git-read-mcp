package handlers

import (
	"context"
	"os"
	"os/exec"
	"strings"
	"testing"

	apimcp "github.com/ipolyzos/git-read-mcp/internal/mcp"
)

func TestHandleCompareText_simpleDiff(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available on PATH")
	}
	repo := t.TempDir()
	_, out, err := HandleCompareText(context.Background(), nil, apimcp.CompareTextInput{
		RepoPath:         repo,
		LeftText:         "hello\n",
		RightText:        "hallo\n",
		ContextLines:     2,
		IgnoreWhitespace: "none",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !out.HadDiff || out.Diff == "" {
		t.Fatalf("expected diff, got %+v", out)
	}
	if out.Stats == nil {
		t.Fatal("expected stats")
	}
}

func TestHandleCompareText_identical(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available on PATH")
	}
	repo := t.TempDir()
	_, out, err := HandleCompareText(context.Background(), nil, apimcp.CompareTextInput{
		RepoPath:     repo,
		LeftText:     "same\n",
		RightText:    "same\n",
		ContextLines: 3,
	})
	if err != nil {
		t.Fatal(err)
	}
	if out.HadDiff {
		t.Fatal("expected no diff")
	}
}

func TestHandleCompareText_cleansTempDir(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available on PATH")
	}
	parent := t.TempDir()
	stop := setCompareTextTempParentForTest(parent)
	defer stop()

	repo := t.TempDir()
	_, _, err := HandleCompareText(context.Background(), nil, apimcp.CompareTextInput{
		RepoPath:     repo,
		LeftText:     "one\n",
		RightText:    "two\n",
		ContextLines: 1,
	})
	if err != nil {
		t.Fatal(err)
	}

	entries, err := os.ReadDir(parent)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 0 {
		t.Fatalf("expected temp dir removed under parent, found: %v", entries)
	}
}

func TestHandleCompareText_rejectsOversizedInHandlerPath(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available on PATH")
	}
	repo := t.TempDir()
	huge := strings.Repeat("z", apimcp.MaxCompareTextBytes+1)
	_, _, err := HandleCompareText(context.Background(), nil, apimcp.CompareTextInput{
		RepoPath:     repo,
		LeftText:     huge,
		RightText:    "a",
		ContextLines: 1,
	})
	if err == nil {
		t.Fatal("expected validation error")
	}
}
