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

// diffFilesMaxOutputOverride is nonzero only in tests (see export_test.go).
var diffFilesMaxOutputOverride int64

// HandleDiffFiles runs git diff --no-index between two repo-scoped files.
func HandleDiffFiles(ctx context.Context, req *mcpsdk.CallToolRequest, in apimcp.DiffFilesInput) (*mcpsdk.CallToolResult, apimcp.DiffFilesOutput, error) {
	_, leftAbs, rightAbs, err := apimcp.ValidateAndResolveDiffFiles(in)
	if err != nil {
		return nil, apimcp.DiffFilesOutput{}, mapToolError(apimcp.ToolNameDiffFiles, req, err)
	}

	ignore, err := apimcp.ParseIgnoreWhitespace(in.IgnoreWhitespace)
	if err != nil {
		return nil, apimcp.DiffFilesOutput{}, mapToolError(apimcp.ToolNameDiffFiles, req, err)
	}

	ctxN := in.ContextLines
	opts := git.DiffNoIndexOptions{
		Unified: &ctxN,
		Ignore:  ignore,
	}
	if max := atomic.LoadInt64(&diffFilesMaxOutputOverride); max != 0 {
		opts.MaxOutputBytes = max
	}

	gitRes, err := git.DiffNoIndex(ctx, leftAbs, rightAbs, opts)
	out := apimcp.DiffFilesOutput{}
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
		return nil, apimcp.DiffFilesOutput{}, mapToolError(apimcp.ToolNameDiffFiles, req, err)
	}
	if !out.Truncated {
		out.Stats = apimcp.DiffLineStatsFromUnifiedDiff(out.Diff)
	}
	return nil, out, nil
}
