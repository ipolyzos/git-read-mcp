package handlers

import (
	"context"
	"errors"
	"sync/atomic"

	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"

	gerr "github.com/ipolyzos/git-read-mcp/internal/errors"
	"github.com/ipolyzos/git-read-mcp/internal/git"
	apimcp "github.com/ipolyzos/git-read-mcp/internal/mcp"
)

// diffDirMaxOutputOverride is nonzero only in tests (see export_test.go).
var diffDirMaxOutputOverride int64

// HandleDiffDirectories runs git diff --no-index between two directory trees.
func HandleDiffDirectories(ctx context.Context, req *mcpsdk.CallToolRequest, in apimcp.DiffDirectoriesInput) (*mcpsdk.CallToolResult, apimcp.DiffDirectoriesOutput, error) {
	_, leftAbs, rightAbs, err := apimcp.ValidateAndResolveDiffDirectories(in)
	if err != nil {
		return nil, apimcp.DiffDirectoriesOutput{}, mapToolError(apimcp.ToolNameDiffDirectories, req, err)
	}

	ignore, err := apimcp.ParseIgnoreWhitespace(in.IgnoreWhitespace)
	if err != nil {
		return nil, apimcp.DiffDirectoriesOutput{}, mapToolError(apimcp.ToolNameDiffDirectories, req, err)
	}

	ctxN := in.ContextLines
	opts := git.DiffNoIndexOptions{
		Unified: &ctxN,
		Ignore:  ignore,
	}
	if max := atomic.LoadInt64(&diffDirMaxOutputOverride); max != 0 {
		opts.MaxOutputBytes = max
	}

	gitRes, err := git.DiffNoIndex(ctx, leftAbs, rightAbs, opts)
	out := apimcp.DiffDirectoriesOutput{}
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
		return nil, apimcp.DiffDirectoriesOutput{}, mapToolError(apimcp.ToolNameDiffDirectories, req, err)
	}
	if !out.Truncated {
		out.Stats = apimcp.DiffLineStatsFromUnifiedDiff(out.Diff)
	}
	return nil, out, nil
}
