package handlers

import (
	"context"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	apimcp "github.com/ipolyzos/git-read-mcp/internal/mcp"
)

func TestHandleSummarizeDiff_inline(t *testing.T) {
	diff := `diff --git a/x b/x
--- a/x
+++ b/x
@@ -1 +1 @@
-old
+new
`
	_, out, err := HandleSummarizeDiff(context.Background(), nil, apimcp.SummarizeDiffInput{Diff: diff})
	if err != nil {
		t.Fatal(err)
	}
	if out.DiffSource != "inline" {
		t.Fatalf("source %q", out.DiffSource)
	}
	if len(out.Files) != 1 || out.Totals.Files != 1 {
		t.Fatalf("files: %+v", out)
	}
	if out.Totals.Insertions != 1 || out.Totals.Deletions != 1 {
		t.Fatalf("totals: %+v", out.Totals)
	}
	if !strings.Contains(out.Summary, "1 file") {
		t.Fatalf("summary: %s", out.Summary)
	}
}

func TestHandleSummarizeDiff_computed(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available on PATH")
	}
	repo := t.TempDir()
	writeFile(t, filepath.Join(repo, "a.txt"), "one\n")
	writeFile(t, filepath.Join(repo, "b.txt"), "two\n")

	_, out, err := HandleSummarizeDiff(context.Background(), nil, apimcp.SummarizeDiffInput{
		RepoPath:         repo,
		LeftPath:         "a.txt",
		RightPath:        "b.txt",
		ContextLines:     1,
		IgnoreWhitespace: "none",
	})
	if err != nil {
		t.Fatal(err)
	}
	if out.DiffSource != "computed" {
		t.Fatalf("source %q", out.DiffSource)
	}
	if len(out.Files) < 1 {
		t.Fatalf("expected files: %+v", out)
	}
	if out.Totals.Insertions+out.Totals.Deletions == 0 {
		t.Fatalf("expected line stats: %+v", out.Totals)
	}
}
