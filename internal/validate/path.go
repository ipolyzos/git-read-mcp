package validate

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// NormalizeRepoPath makes repoPath absolute, ensures it exists, and is a directory.
// When the path is a symlink to a directory, the canonical target path is returned.
func NormalizeRepoPath(repoPath string) (string, error) {
	if strings.TrimSpace(repoPath) == "" {
		return "", fmt.Errorf("validate: repo_path is empty")
	}
	abs, err := filepath.Abs(repoPath)
	if err != nil {
		return "", fmt.Errorf("validate: repo_path: %w", err)
	}
	abs = filepath.Clean(abs)

	fi, err := os.Stat(abs)
	if err != nil {
		return "", fmt.Errorf("validate: repo_path: %w", err)
	}
	if !fi.IsDir() {
		return "", fmt.Errorf("validate: repo_path: %w", ErrNotDirectory)
	}

	if real, err := filepath.EvalSymlinks(abs); err == nil {
		if fi2, err := os.Stat(real); err == nil && fi2.IsDir() {
			return filepath.Clean(real), nil
		}
	}
	return abs, nil
}

// ResolvePath resolves p against repoPathAbs. Absolute inputs are cleaned and returned as-is.
// Relative inputs are joined to the repository root (repoPathAbs must already be absolute and clean).
func ResolvePath(repoPathAbs, p string) (string, error) {
	if strings.TrimSpace(repoPathAbs) == "" {
		return "", fmt.Errorf("validate: repo_path is empty")
	}
	if !filepath.IsAbs(repoPathAbs) {
		return "", ErrNotAbsolute
	}
	base := filepath.Clean(repoPathAbs)

	p = strings.TrimSpace(p)
	if p == "" {
		return "", fmt.Errorf("validate: path is empty")
	}

	var resolved string
	if filepath.IsAbs(p) {
		resolved = filepath.Clean(p)
	} else {
		resolved = filepath.Join(base, filepath.Clean(p))
	}
	return filepath.Clean(resolved), nil
}

// EnsureWithinRepo checks that absPath is repoPathAbs or contained within it.
// Both arguments must be absolute paths; filepath.Clean is applied.
func EnsureWithinRepo(repoPathAbs, absPath string) error {
	if !filepath.IsAbs(repoPathAbs) || !filepath.IsAbs(absPath) {
		return ErrNotAbsolute
	}
	base := filepath.Clean(repoPathAbs)
	target := filepath.Clean(absPath)

	rel, err := filepath.Rel(base, target)
	if err != nil {
		return fmt.Errorf("validate: %w", err)
	}
	if rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return ErrOutsideRepo
	}
	return nil
}

// EnsurePathExists verifies path exists; wantDir distinguishes file vs directory expectations.
func EnsurePathExists(path string, wantDir bool) error {
	fi, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("validate: path %q: %w", path, err)
	}
	switch {
	case wantDir && !fi.IsDir():
		return fmt.Errorf("validate: %q is not a directory", path)
	case !wantDir && fi.IsDir():
		return fmt.Errorf("validate: %q is a directory; expected file", path)
	default:
		return nil
	}
}
