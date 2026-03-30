package handlers

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"sync"

	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"

	gerr "github.com/ipolyzos/git-read-mcp/internal/errors"
	"github.com/ipolyzos/git-read-mcp/internal/git"
	apimcp "github.com/ipolyzos/git-read-mcp/internal/mcp"
)

var (
	compareTextTempMu     sync.Mutex
	compareTextTempParent string // if non-empty, MkdirTemp uses this parent (tests only)
)

func mkdirTempForCompare() (string, error) {
	compareTextTempMu.Lock()
	parent := compareTextTempParent
	compareTextTempMu.Unlock()
	if parent != "" {
		return os.MkdirTemp(parent, "git-read-mcp-compare-*")
	}
	return os.MkdirTemp("", "git-read-mcp-compare-*")
}

// HandleCompareText materializes two blobs under a dedicated temp directory, runs
// git diff --no-index, and removes the directory (best-effort). The repository is
// never modified; repo_path is validated only.
func HandleCompareText(ctx context.Context, req *mcpsdk.CallToolRequest, in apimcp.CompareTextInput) (*mcpsdk.CallToolResult, apimcp.CompareTextOutput, error) {
	if err := apimcp.ValidateCompareTextInput(in); err != nil {
		return nil, apimcp.CompareTextOutput{}, mapToolError(apimcp.ToolNameCompareText, req, err)
	}

	ignore, err := apimcp.ParseIgnoreWhitespace(in.IgnoreWhitespace)
	if err != nil {
		return nil, apimcp.CompareTextOutput{}, mapToolError(apimcp.ToolNameCompareText, req, err)
	}

	ctxN := in.ContextLines
	opts := git.DiffNoIndexOptions{
		Unified: &ctxN,
		Ignore:  ignore,
	}

	dir, err := mkdirTempForCompare()
	if err != nil {
		return nil, apimcp.CompareTextOutput{}, mapToolError(apimcp.ToolNameCompareText, req, err)
	}
	defer func() { _ = os.RemoveAll(dir) }()

	leftPath := filepath.Join(dir, "left.txt")
	rightPath := filepath.Join(dir, "right.txt")
	if err := os.WriteFile(leftPath, []byte(in.LeftText), 0o600); err != nil {
		return nil, apimcp.CompareTextOutput{}, mapToolError(apimcp.ToolNameCompareText, req, err)
	}
	if err := os.WriteFile(rightPath, []byte(in.RightText), 0o600); err != nil {
		return nil, apimcp.CompareTextOutput{}, mapToolError(apimcp.ToolNameCompareText, req, err)
	}

	gitRes, err := git.DiffNoIndex(ctx, leftPath, rightPath, opts)
	out := apimcp.CompareTextOutput{}
	if gitRes != nil {
		out.Diff = gitRes.Diff
		out.Truncated = gitRes.Truncated
		out.HadDiff = gitRes.HadDiff
	}
	if err != nil {
		if errors.Is(err, gerr.ErrExecTruncated) {
			out.Truncated = true
			return nil, out, nil
		}
		return nil, apimcp.CompareTextOutput{}, mapToolError(apimcp.ToolNameCompareText, req, err)
	}
	if !out.Truncated {
		out.Stats = apimcp.DiffLineStatsFromUnifiedDiff(out.Diff)
	}
	return nil, out, nil
}
