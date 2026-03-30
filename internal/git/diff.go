package git

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/ipolyzos/git-read-mcp/internal/config"
	gerr "github.com/ipolyzos/git-read-mcp/internal/errors"
	"github.com/ipolyzos/git-read-mcp/internal/execx"
)

// IgnoreWhitespace selects an allowlisted git whitespace ignored mode.
type IgnoreWhitespace int

const (
	IgnoreWhitespaceNone IgnoreWhitespace = iota
	IgnoreWhitespaceAll
	IgnoreWhitespaceChange
)

// DiffNoIndexOptions configures git diff --no-index invocation. Only allowlisted flags
// are ever passed; user strings never become flags.
type DiffNoIndexOptions struct {
	// Unified sets --unified=n. Nil means 3 lines of context; explicit 0 is allowed.
	Unified *int
	Ignore  IgnoreWhitespace
	Timeout time.Duration
	// MaxOutputBytes caps captured stdout+stderr combined. Zero selects a default cap; negative means unlimited.
	MaxOutputBytes int64
}

// DiffNoIndexResult is the parsed outcome of a no-index diff.
type DiffNoIndexResult struct {
	Diff      string
	HadDiff   bool
	ExitCode  int
	Truncated bool
}

const (
	defaultDiffTimeout  = 30 * time.Second
	defaultDiffMaxBytes = 8 << 20
	maxUnifiedLines     = 200
)

// DiffNoIndex runs `git diff --no-index` with a fixed flag allowlist. Git exits with
// status 1 when a textual difference exists; that is treated as success with HadDiff set.
func DiffNoIndex(ctx context.Context, left, right string, opts DiffNoIndexOptions) (*DiffNoIndexResult, error) {
	if err := validatePathArg("left", left); err != nil {
		return nil, err
	}
	if err := validatePathArg("right", right); err != nil {
		return nil, err
	}

	args, err := buildDiffNoIndexArgs(left, right, opts)
	if err != nil {
		return nil, err
	}

	cfg := config.Get()
	timeout := opts.Timeout
	if timeout <= 0 {
		if cfg.DiffTimeout > 0 {
			timeout = cfg.DiffTimeout
		} else {
			timeout = defaultDiffTimeout
		}
	}
	maxBytes := opts.MaxOutputBytes
	if maxBytes == 0 {
		if cfg.MaxOutputBytes > 0 {
			maxBytes = cfg.MaxOutputBytes
		} else {
			maxBytes = defaultDiffMaxBytes
		}
	}

	res, err := execx.Run(ctx, "git", args, EnvGitReadOnly(), "", timeout, maxBytes)
	if err != nil {
		return nil, err
	}

	out := &DiffNoIndexResult{
		Truncated: res.Truncated,
		Diff:      string(res.Stdout),
	}

	switch res.ExitCode {
	case 0:
		out.ExitCode = 0
		out.HadDiff = false
	case 1:
		out.ExitCode = 1
		out.HadDiff = true
	default:
		return nil, gerr.NonZeroExit(res.ExitCode, res.Stderr)
	}

	if res.Truncated {
		return out, fmt.Errorf("git: diff output truncated: %w", gerr.ErrExecTruncated)
	}
	return out, nil
}

func validatePathArg(name, p string) error {
	if strings.TrimSpace(p) == "" {
		return fmt.Errorf("git: %s path is empty", name)
	}
	if strings.HasPrefix(p, "-") {
		return fmt.Errorf("git: %s path must not begin with '-'", name)
	}
	return nil
}

func unifiedValue(u *int) (int, error) {
	if u == nil {
		return 3, nil
	}
	if *u < 0 || *u > maxUnifiedLines {
		return 0, fmt.Errorf("git: unified context %d out of range [0,%d]", *u, maxUnifiedLines)
	}
	return *u, nil
}

func buildDiffNoIndexArgs(left, right string, opts DiffNoIndexOptions) ([]string, error) {
	n, err := unifiedValue(opts.Unified)
	if err != nil {
		return nil, err
	}

	args := []string{"diff", "--no-index", "--unified=" + strconv.Itoa(n)}
	switch opts.Ignore {
	case IgnoreWhitespaceNone:
	case IgnoreWhitespaceAll:
		args = append(args, "--ignore-all-space")
	case IgnoreWhitespaceChange:
		args = append(args, "--ignore-space-change")
	default:
		return nil, fmt.Errorf("git: unknown IgnoreWhitespace value")
	}
	args = append(args, "--", left, right)
	return args, nil
}
