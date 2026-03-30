package mcp

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/ipolyzos/git-read-mcp/internal/validate"
)

func TestValidateAndResolveDiffDirectories_requiresDirectories(t *testing.T) {
	t.Parallel()
	repo := t.TempDir()
	f := filepath.Join(repo, "f.txt")
	writeFile(t, f, "x")
	d := filepath.Join(repo, "d")
	if err := os.Mkdir(d, 0o755); err != nil {
		t.Fatal(err)
	}

	_, _, _, err := ValidateAndResolveDiffDirectories(DiffDirectoriesInput{
		RepoPath:     repo,
		LeftDir:      f,
		RightDir:     "d",
		ContextLines: 1,
	})
	if err == nil {
		t.Fatal("expected error when left_dir is a file")
	}

	_, _, _, err = ValidateAndResolveDiffDirectories(DiffDirectoriesInput{
		RepoPath:     repo,
		LeftDir:      "d",
		RightDir:     f,
		ContextLines: 1,
	})
	if err == nil {
		t.Fatal("expected error when right_dir is a file")
	}
}

func TestValidateAndResolveDiffDirectories_absoluteOutsideRepo(t *testing.T) {
	t.Parallel()
	repo := t.TempDir()
	other := t.TempDir()
	d1 := filepath.Join(other, "tree1")
	d2 := filepath.Join(other, "tree2")
	if err := os.MkdirAll(d1, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(d2, 0o755); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(d1, "f.txt"), "a\n")
	writeFile(t, filepath.Join(d2, "f.txt"), "b\n")

	_, left, right, err := ValidateAndResolveDiffDirectories(DiffDirectoriesInput{
		RepoPath:     repo,
		LeftDir:      d1,
		RightDir:     d2,
		ContextLines: 2,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !filepath.IsAbs(left) || !filepath.IsAbs(right) {
		t.Fatalf("paths: %s %s", left, right)
	}
}

func TestValidateAndResolveDiffDirectories_contextLines(t *testing.T) {
	t.Parallel()
	repo := t.TempDir()
	d := filepath.Join(repo, "d")
	if err := os.MkdirAll(d, 0o755); err != nil {
		t.Fatal(err)
	}
	_, _, _, err := ValidateAndResolveDiffDirectories(DiffDirectoriesInput{
		RepoPath:     repo,
		LeftDir:      d,
		RightDir:     d,
		ContextLines: 51,
	})
	if err == nil || !errors.Is(err, validate.ErrIntOutOfRange) {
		t.Fatalf("got %v", err)
	}
}
