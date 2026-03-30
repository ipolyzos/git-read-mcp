// Package diffparse implements a line-oriented parser for unified diffs produced by Git.
package diffparse

import (
	"strings"
)

// Options bounds parsing work and input size.
type Options struct {
	MaxDiffBytes int
	MaxFiles     int
	MaxHunks     int
}

// DefaultOptions returns conservative defaults.
func DefaultOptions() Options {
	return Options{
		MaxDiffBytes: 2 << 20,
		MaxFiles:     500,
		MaxHunks:     10000,
	}
}

// FileChange aggregates per-file statistics.
type FileChange struct {
	Path        string `json:"path"`
	OldPath     string `json:"old_path,omitempty"`
	Rename      bool   `json:"rename,omitempty"`
	Binary      bool   `json:"binary,omitempty"`
	Insertions  int    `json:"insertions"`
	Deletions   int    `json:"deletions"`
	Hunks       int    `json:"hunks"`
}

// Result is the outcome of ParseUnified.
type Result struct {
	Files           []FileChange `json:"files"`
	TotalInsertions int          `json:"total_insertions"`
	TotalDeletions  int          `json:"total_deletions"`
	TotalHunks      int          `json:"total_hunks"`
	InputTruncated  bool         `json:"input_truncated"`
	ParseTruncated  bool         `json:"parse_truncated"`
}

type fileAcc struct {
	path, oldPath string
	rename         bool
	binary         bool
	ins, del       int
	hunks          int
}

func (f *fileAcc) toChange() FileChange {
	return FileChange{
		Path:       f.path,
		OldPath:    f.oldPath,
		Rename:     f.rename,
		Binary:     f.binary,
		Insertions: f.ins,
		Deletions:  f.del,
		Hunks:      f.hunks,
	}
}

// ParseUnified scans unified diff text and accumulates approximate line/hunk counts.
// It tolerates missing `diff --git` headers (uses --- / +++), skips `+++`/`---` file
// headers for counting, ignores `\ No newline...`, and treats a binary hunk as having
// no line-level insertions/deletions.
func ParseUnified(diff string, o Options) Result {
	if o.MaxDiffBytes <= 0 {
		o.MaxDiffBytes = DefaultOptions().MaxDiffBytes
	}
	if o.MaxFiles <= 0 {
		o.MaxFiles = DefaultOptions().MaxFiles
	}
	if o.MaxHunks <= 0 {
		o.MaxHunks = DefaultOptions().MaxHunks
	}

	res := Result{}
	if len(diff) > o.MaxDiffBytes {
		diff = diff[:o.MaxDiffBytes]
		res.InputTruncated = true
	}

	lines := strings.Split(strings.ReplaceAll(diff, "\r\n", "\n"), "\n")

	var cur *fileAcc
	inHunk := false
	globalHunks := 0

	flush := func() {
		if cur == nil {
			return
		}
		if len(res.Files) >= o.MaxFiles {
			res.ParseTruncated = true
			cur = nil
			inHunk = false
			return
		}
		fc := cur.toChange()
		res.Files = append(res.Files, fc)
		res.TotalInsertions += fc.Insertions
		res.TotalDeletions += fc.Deletions
		res.TotalHunks += fc.Hunks
		cur = nil
		inHunk = false
	}

	for _, line := range lines {
		if res.ParseTruncated {
			break
		}

		switch {
		case strings.HasPrefix(line, "diff --git "):
			flush()
			if len(res.Files) >= o.MaxFiles {
				res.ParseTruncated = true
				break
			}
			cur = parseDiffGitLine(line)
			inHunk = false

		case strings.HasPrefix(line, "rename from "):
			if cur == nil {
				cur = &fileAcc{}
			}
			cur.rename = true
			cur.oldPath = stripGitPathPrefix(strings.TrimPrefix(line, "rename from "))

		case strings.HasPrefix(line, "rename to "):
			if cur == nil {
				cur = &fileAcc{}
			}
			cur.rename = true
			cur.path = stripGitPathPrefix(strings.TrimPrefix(line, "rename to "))

		case strings.HasPrefix(line, "--- "):
			path := minusPlusPath(line, "--- ")
			if path == "/dev/null" {
				inHunk = false
				continue
			}
			if cur == nil {
				flush()
				if len(res.Files) >= o.MaxFiles {
					res.ParseTruncated = true
					break
				}
				cur = &fileAcc{oldPath: path, path: path}
			} else if cur.oldPath == "" || cur.path == cur.oldPath {
				cur.oldPath = path
				if cur.path == "" {
					cur.path = path
				}
			}
			inHunk = false

		case strings.HasPrefix(line, "+++ "):
			path := minusPlusPath(line, "+++ ")
			if path == "/dev/null" {
				inHunk = false
				continue
			}
			if cur == nil {
				if len(res.Files) >= o.MaxFiles {
					res.ParseTruncated = true
					break
				}
				cur = &fileAcc{path: path, oldPath: path}
			} else {
				cur.path = path
			}
			inHunk = false

		case strings.HasPrefix(line, "Binary files ") && strings.HasSuffix(line, " differ"):
			if cur == nil {
				cur = &fileAcc{}
			}
			cur.binary = true
			inHunk = false

		case strings.HasPrefix(line, "@@"):
			if cur == nil {
				continue
			}
			if globalHunks >= o.MaxHunks {
				res.ParseTruncated = true
				break
			}
			globalHunks++
			cur.hunks++
			inHunk = true

		case line == `\ No newline at end of file`:
			continue

		default:
			if !inHunk || cur == nil || cur.binary {
				continue
			}
			if len(line) == 0 {
				continue
			}
			switch line[0] {
			case '+':
				if !strings.HasPrefix(line, "+++") {
					cur.ins++
				}
			case '-':
				if !strings.HasPrefix(line, "---") {
					cur.del++
				}
			}
		}
	}

	flush()
	return res
}

func parseDiffGitLine(line string) *fileAcc {
	rest := strings.TrimPrefix(line, "diff --git ")
	if tabIdx := strings.IndexByte(rest, '\t'); tabIdx >= 0 {
		rest = rest[:tabIdx]
	}
	idx := strings.Index(rest, " b/")
	if idx < 0 {
		return &fileAcc{path: rest}
	}
	oldRaw := rest[:idx]
	newRaw := rest[idx+len(" b/"):]
	oldPath := stripGitPathPrefix(strings.TrimPrefix(oldRaw, "a/"))
	newPath := stripGitPathPrefix(newRaw)
	return &fileAcc{path: newPath, oldPath: oldPath}
}

func minusPlusPath(line, prefix string) string {
	p := strings.TrimPrefix(line, prefix)
	p = strings.TrimPrefix(p, "a/")
	p = strings.TrimPrefix(p, "b/")
	if i := strings.IndexByte(p, '\t'); i >= 0 {
		p = p[:i]
	}
	return p
}

func stripGitPathPrefix(p string) string {
	p = strings.TrimPrefix(p, `"`)
	p = strings.TrimSuffix(p, `"`)
	p = strings.ReplaceAll(p, `\t`, "\t")
	p = strings.ReplaceAll(p, `\\`, `\`)
	return p
}
