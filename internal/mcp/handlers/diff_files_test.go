package handlers

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	apimcp "github.com/ipolyzos/git-read-mcp/internal/mcp"
)

func TestHandleDiffFiles_gitDiff(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available on PATH")
	}
	repo := t.TempDir()
	writeFile(t, filepath.Join(repo, "a.txt"), "hello\n")
	writeFile(t, filepath.Join(repo, "b.txt"), "hallo\n")

	_, out, err := HandleDiffFiles(context.Background(), nil, apimcp.DiffFilesInput{
		RepoPath:         repo,
		LeftPath:         "a.txt",
		RightPath:        "b.txt",
		ContextLines:     2,
		IgnoreWhitespace: "none",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !out.HadDiff {
		t.Fatal("expected differences")
	}
	if out.Diff == "" {
		t.Fatal("expected non-empty diff")
	}
	if out.Truncated {
		t.Fatal("unexpected truncation")
	}
	if out.Stats == nil {
		t.Fatal("expected stats")
	}
	if out.Stats.Insertions+out.Stats.Deletions == 0 {
		t.Fatalf("expected nonzero stats, got %+v", out.Stats)
	}
}

func TestHandleDiffFiles_invalidIgnore(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available on PATH")
	}
	repo := t.TempDir()
	writeFile(t, filepath.Join(repo, "a.txt"), "a")
	writeFile(t, filepath.Join(repo, "b.txt"), "b")

	_, _, err := HandleDiffFiles(context.Background(), nil, apimcp.DiffFilesInput{
		RepoPath:         repo,
		LeftPath:         "a.txt",
		RightPath:        "b.txt",
		ContextLines:     1,
		IgnoreWhitespace: "absolutely-not",
	})
	if err == nil {
		t.Fatal("expected error")
	}
}

func writeFile(t *testing.T, path, body string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
}
