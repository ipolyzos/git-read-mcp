package handlers

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	apimcp "github.com/ipolyzos/git-read-mcp/internal/mcp"
)

func TestHandleDiffDirectories_treeDiff(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available on PATH")
	}
	repo := t.TempDir()
	left := filepath.Join(repo, "L")
	right := filepath.Join(repo, "R")
	if err := os.MkdirAll(filepath.Join(left, "sub"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(right, "sub"), 0o755); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(left, "sub", "f.txt"), "alpha\n")
	writeFile(t, filepath.Join(right, "sub", "f.txt"), "omega\n")

	_, out, err := HandleDiffDirectories(context.Background(), nil, apimcp.DiffDirectoriesInput{
		RepoPath:     repo,
		LeftDir:      "L",
		RightDir:     "R",
		ContextLines: 2,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !out.HadDiff {
		t.Fatal("expected differences between trees")
	}
	if out.Diff == "" || !strings.Contains(out.Diff, "f.txt") {
		t.Fatalf("unexpected diff: %q", out.Diff)
	}
	if out.Truncated {
		t.Fatal("unexpected truncation")
	}
	if out.Stats == nil {
		t.Fatal("expected stats")
	}
}

func TestHandleDiffDirectories_truncation(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available on PATH")
	}
	repo := t.TempDir()
	left := filepath.Join(repo, "L")
	right := filepath.Join(repo, "R")
	if err := os.MkdirAll(left, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(right, 0o755); err != nil {
		t.Fatal(err)
	}
	var b strings.Builder
	for i := 0; i < 200; i++ {
		b.WriteString("unchanged context line\n")
	}
	base := b.String()
	writeFile(t, filepath.Join(left, "big.txt"), base+"OLD\n")
	writeFile(t, filepath.Join(right, "big.txt"), base+"NEW\n")

	cleanup := setDiffDirMaxOutputForTest(400)
	defer cleanup()

	_, out, err := HandleDiffDirectories(context.Background(), nil, apimcp.DiffDirectoriesInput{
		RepoPath:     repo,
		LeftDir:      left,
		RightDir:     right,
		ContextLines: 2,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !out.Truncated {
		t.Fatal("expected truncated output with tiny byte cap")
	}
}
