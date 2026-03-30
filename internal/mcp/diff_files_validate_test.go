package mcp

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/ipolyzos/git-read-mcp/internal/validate"
)

func TestParseIgnoreWhitespace(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		in string
		ok bool
	}{
		{"", true},
		{"none", true},
		{"NONE", true},
		{" all ", true},
		{"change", true},
		{"nope", false},
	} {
		_, err := ParseIgnoreWhitespace(tt.in)
		if tt.ok && err != nil {
			t.Fatalf("%q: %v", tt.in, err)
		}
		if !tt.ok && err == nil {
			t.Fatalf("%q: want error", tt.in)
		}
	}
}

func TestValidateAndResolveDiffFiles_contextLines(t *testing.T) {
	t.Parallel()
	repo := t.TempDir()
	writeFile(t, filepath.Join(repo, "a.txt"), "a")
	writeFile(t, filepath.Join(repo, "b.txt"), "b")

	_, _, _, err := ValidateAndResolveDiffFiles(DiffFilesInput{
		RepoPath:     repo,
		LeftPath:     "a.txt",
		RightPath:    "b.txt",
		ContextLines: 51,
	})
	if !errors.Is(err, validate.ErrIntOutOfRange) {
		t.Fatalf("got %v", err)
	}
}

func TestValidateAndResolveDiffFiles_outsideRepo(t *testing.T) {
	t.Parallel()
	repo := t.TempDir()
	outside := t.TempDir()
	writeFile(t, filepath.Join(repo, "a.txt"), "a")
	extFile := filepath.Join(outside, "b.txt")
	writeFile(t, extFile, "b")

	_, _, _, err := ValidateAndResolveDiffFiles(DiffFilesInput{
		RepoPath:     repo,
		LeftPath:     "a.txt",
		RightPath:    extFile,
		ContextLines: 3,
	})
	if !errors.Is(err, validate.ErrOutsideRepo) {
		t.Fatalf("got %v", err)
	}
}

func TestValidateAndResolveDiffFiles_okRelative(t *testing.T) {
	t.Parallel()
	repo := t.TempDir()
	writeFile(t, filepath.Join(repo, "a.txt"), "hello\n")
	writeFile(t, filepath.Join(repo, "b.txt"), "hallo\n")

	_, left, right, err := ValidateAndResolveDiffFiles(DiffFilesInput{
		RepoPath:     repo,
		LeftPath:     "a.txt",
		RightPath:    "b.txt",
		ContextLines: 0,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !filepath.IsAbs(left) || !filepath.IsAbs(right) {
		t.Fatalf("paths not absolute: %s %s", left, right)
	}
}

func TestDiffLineStatsFromUnifiedDiff(t *testing.T) {
	t.Parallel()
	diff := "--- a.txt\n+++ b.txt\n@@ -1 +1 @@\n-old\n+new\n"
	st := DiffLineStatsFromUnifiedDiff(diff)
	if st.Insertions != 1 || st.Deletions != 1 {
		t.Fatalf("got %+v", st)
	}
}

func writeFile(t *testing.T, path, body string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
}
