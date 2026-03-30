package validate

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestNormalizeRepoPath(t *testing.T) {
	t.Parallel()
	repo := t.TempDir()
	if err := os.WriteFile(filepath.Join(repo, "file"), []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}

	got, err := NormalizeRepoPath(repo)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if !filepath.IsAbs(got) {
		t.Fatalf("want absolute path: %s", got)
	}
	if _, err := NormalizeRepoPath(filepath.Join(repo, "file")); err == nil {
		t.Fatal("expected error for non-directory")
	}
	if _, err := NormalizeRepoPath(filepath.Join(repo, "missing")); err == nil {
		t.Fatal("expected error for missing path")
	}
}

func TestResolvePath_relativeAndAbsolute(t *testing.T) {
	t.Parallel()
	repo := t.TempDir()
	repoAbs, err := NormalizeRepoPath(repo)
	if err != nil {
		t.Fatal(err)
	}

	rel := filepath.Join("sub", "dir")
	want := filepath.Clean(filepath.Join(repoAbs, rel))
	got, err := ResolvePath(repoAbs, rel)
	if err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Fatalf("ResolvePath relative: got %s want %s", got, want)
	}

	absInput := filepath.Join(t.TempDir(), "elsewhere")
	got2, err := ResolvePath(repoAbs, absInput)
	if err != nil {
		t.Fatal(err)
	}
	if got2 != filepath.Clean(absInput) {
		t.Fatalf("ResolvePath absolute: got %s", got2)
	}
}

func TestEnsureWithinRepo_allowsInside(t *testing.T) {
	t.Parallel()
	repo := t.TempDir()
	child := filepath.Join(repo, "nested")
	if err := os.MkdirAll(child, 0o755); err != nil {
		t.Fatal(err)
	}
	repoAbs, err := NormalizeRepoPath(repo)
	if err != nil {
		t.Fatal(err)
	}
	target := filepath.Join(repoAbs, "nested")
	if err := EnsureWithinRepo(repoAbs, target); err != nil {
		t.Fatal(err)
	}
	if err := EnsureWithinRepo(repoAbs, repoAbs); err != nil {
		t.Fatalf("root should be allowed: %v", err)
	}
}

func TestEnsureWithinRepo_blocksTraversal(t *testing.T) {
	t.Parallel()
	repo := t.TempDir()
	repoAbs, err := NormalizeRepoPath(repo)
	if err != nil {
		t.Fatal(err)
	}

	outside, err := ResolvePath(repoAbs, filepath.Join("..", "..", "..", "etc", "passwd"))
	if err != nil {
		t.Fatal(err)
	}
	if err := EnsureWithinRepo(repoAbs, outside); !errors.Is(err, ErrOutsideRepo) {
		t.Fatalf("want ErrOutsideRepo, got %v", err)
	}
}

func TestEnsurePathExists(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	file := filepath.Join(dir, "f.txt")
	if err := os.WriteFile(file, []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := EnsurePathExists(dir, true); err != nil {
		t.Fatal(err)
	}
	if err := EnsurePathExists(file, false); err != nil {
		t.Fatal(err)
	}
	if err := EnsurePathExists(file, true); err == nil {
		t.Fatal("expected dir check failure for file")
	}
}

func TestResolvePath_requiresAbsoluteRepo(t *testing.T) {
	t.Parallel()
	if _, err := ResolvePath("relative/repo", "x"); !errors.Is(err, ErrNotAbsolute) {
		t.Fatalf("got %v", err)
	}
}
