package diffparse

import (
	"strings"
	"testing"
)

func TestParseUnified_multiFile(t *testing.T) {
	diff := `diff --git a/foo.go b/foo.go
--- a/foo.go
+++ b/foo.go
@@ -1,2 +1,3 @@
 package foo
+// x
 const y = 1
diff --git a/bar.go b/bar.go
--- a/bar.go
+++ b/bar.go
@@ -1 +1 @@
-old
+new
`
	o := Options{MaxDiffBytes: 1 << 20, MaxFiles: 100, MaxHunks: 100}
	r := ParseUnified(diff, o)
	if len(r.Files) != 2 {
		t.Fatalf("files: got %d want 2: %#v", len(r.Files), r.Files)
	}
	if r.Files[0].Path != "foo.go" || r.Files[0].Hunks != 1 || r.Files[0].Insertions != 1 {
		t.Errorf("file0: %#v", r.Files[0])
	}
	if r.Files[1].Path != "bar.go" || r.Files[1].Insertions != 1 || r.Files[1].Deletions != 1 {
		t.Errorf("file1: %#v", r.Files[1])
	}
	if r.TotalHunks != 2 || r.TotalInsertions != 2 || r.TotalDeletions != 1 {
		t.Errorf("totals: %+v", r)
	}
}

func TestParseUnified_rename(t *testing.T) {
	diff := `diff --git a/oldname.txt b/newname.txt
similarity index 100%
rename from oldname.txt
rename to newname.txt
`
	r := ParseUnified(diff, Options{MaxDiffBytes: 4096, MaxFiles: 10, MaxHunks: 10})
	if len(r.Files) != 1 {
		t.Fatalf("got %d files", len(r.Files))
	}
	f := r.Files[0]
	if !f.Rename || f.Path != "newname.txt" || f.OldPath != "oldname.txt" {
		t.Errorf("rename file: %#v", f)
	}
}

func TestParseUnified_binary(t *testing.T) {
	diff := `diff --git a/photo.png b/photo.png
index 1111111..2222222 100644
--- a/photo.png
+++ b/photo.png
Binary files a/photo.png and b/photo.png differ
`
	r := ParseUnified(diff, DefaultOptions())
	if len(r.Files) != 1 {
		t.Fatalf("got %d files", len(r.Files))
	}
	f := r.Files[0]
	if !f.Binary || f.Path != "photo.png" {
		t.Errorf("binary file: %#v", f)
	}
	if f.Insertions != 0 || f.Deletions != 0 || f.Hunks != 0 {
		t.Errorf("binary should not count hunk lines: %#v", f)
	}
}

func TestParseUnified_skipsHeadersInHunk(t *testing.T) {
	diff := `diff --git a/x b/x
--- a/x
+++ b/x
@@ -1,3 +1,3 @@
 a
-b
+--- not a header
+c
`
	r := ParseUnified(diff, DefaultOptions())
	if len(r.Files) != 1 {
		t.Fatal(r.Files)
	}
	if r.Files[0].Deletions != 1 || r.Files[0].Insertions != 2 {
		t.Errorf("got ins=%d del=%d", r.Files[0].Insertions, r.Files[0].Deletions)
	}
}

func TestParseUnified_maxBytesTruncates(t *testing.T) {
	longLine := strings.Repeat("x", 100) + "\n"
	var sb strings.Builder
	sb.WriteString("diff --git a/z b/z\n--- a/z\n+++ b/z\n@@ -1 +1 @@\n")
	sb.WriteString("-" + longLine)
	sb.WriteString("+" + longLine)
	full := sb.String()
	o := Options{MaxDiffBytes: 80, MaxFiles: 10, MaxHunks: 10}
	r := ParseUnified(full, o)
	if !r.InputTruncated {
		t.Error("expected InputTruncated")
	}
	if len(r.Files) == 0 && r.TotalInsertions == 0 {
		t.Log("truncated before complete parse is ok")
	}
}

func TestParseUnified_maxFiles(t *testing.T) {
	var sb strings.Builder
	for i := range 5 {
		sb.WriteString("diff --git a/f")
		sb.WriteString(string(rune('0'+i)))
		sb.WriteString(".go b/f")
		sb.WriteString(string(rune('0'+i)))
		sb.WriteString(".go\n--- a/x\n+++ b/x\n@@ -1 +1 @@\n-a\n+b\n")
	}
	o := Options{MaxDiffBytes: 1 << 20, MaxFiles: 2, MaxHunks: 100}
	r := ParseUnified(sb.String(), o)
	if len(r.Files) != 2 {
		t.Fatalf("want 2 files, got %d", len(r.Files))
	}
	if !r.ParseTruncated {
		t.Error("ParseTruncated")
	}
}

func TestParseUnified_maxHunks(t *testing.T) {
	var sb strings.Builder
	sb.WriteString("diff --git a/x b/x\n--- a/x\n+++ b/x\n")
	for range 4 {
		sb.WriteString("@@ -1 +1 @@\n-a\n+b\n")
	}
	o := Options{MaxDiffBytes: 1 << 20, MaxFiles: 10, MaxHunks: 2}
	r := ParseUnified(sb.String(), o)
	if r.TotalHunks != 2 || !r.ParseTruncated {
		t.Fatalf("hunks=%d ParseTruncated=%v", r.TotalHunks, r.ParseTruncated)
	}
}
