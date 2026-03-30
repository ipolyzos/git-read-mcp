package git

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/ipolyzos/git-read-mcp/internal/config"
	gerr "github.com/ipolyzos/git-read-mcp/internal/errors"
	"github.com/ipolyzos/git-read-mcp/internal/execx"
)

const (
	detectGitTimeout  = 5 * time.Second
	detectGitMaxBytes = int64(64 * 1024)
)

// DetectGit runs a short `git --version` probe and returns the first line of output.
func DetectGit(ctx context.Context) (string, error) {
	cfg := config.Get()
	timeout := cfg.GitProbeTimeout
	if timeout <= 0 {
		timeout = detectGitTimeout
	}
	maxBytes := cfg.GitProbeMaxBytes
	if maxBytes <= 0 {
		maxBytes = detectGitMaxBytes
	}
	res, err := execx.Run(ctx, "git", []string{"--version"}, EnvGitReadOnly(), "", timeout, maxBytes)
	if err != nil {
		return "", err
	}
	if res.Truncated {
		return "", fmt.Errorf("git: version output truncated: %w", gerr.ErrExecTruncated)
	}
	if res.ExitCode != 0 {
		return "", gerr.NonZeroExit(res.ExitCode, res.Stderr)
	}
	out := strings.TrimSpace(string(res.Stdout))
	if out == "" {
		return "", fmt.Errorf("git: empty --version output")
	}
	if i := strings.IndexByte(out, '\n'); i >= 0 {
		out = strings.TrimSpace(out[:i])
	}
	return out, nil
}
