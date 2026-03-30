package handlers

import (
	"context"
	"errors"
	"strings"

	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"

	gerr "github.com/ipolyzos/git-read-mcp/internal/errors"
	"github.com/ipolyzos/git-read-mcp/internal/git"
	apimcp "github.com/ipolyzos/git-read-mcp/internal/mcp"
	"github.com/ipolyzos/git-read-mcp/internal/mcp/diffparse"
)

// HandleSummarizeDiff parses a unified diff or computes one like diff_files, then returns structured stats.
func HandleSummarizeDiff(ctx context.Context, req *mcpsdk.CallToolRequest, in apimcp.SummarizeDiffInput) (*mcpsdk.CallToolResult, apimcp.SummarizeDiffOutput, error) {
	if err := apimcp.ValidateSummarizeDiffInput(in); err != nil {
		return nil, apimcp.SummarizeDiffOutput{}, err
	}

	opts := apimcp.SummarizeDiffOptions(in)

	if strings.TrimSpace(in.Diff) != "" {
		parsed := diffparse.ParseUnified(in.Diff, opts)
		out := apimcp.BuildSummarizeOutput(parsed, "inline", false)
		return nil, out, nil
	}

	_, leftAbs, rightAbs, err := apimcp.ValidateAndResolveDiffFiles(apimcp.DiffFilesInput{
		RepoPath:         in.RepoPath,
		LeftPath:         in.LeftPath,
		RightPath:        in.RightPath,
		ContextLines:     in.ContextLines,
		IgnoreWhitespace: in.IgnoreWhitespace,
	})
	if err != nil {
		return nil, apimcp.SummarizeDiffOutput{}, err
	}

	ignore, err := apimcp.ParseIgnoreWhitespace(in.IgnoreWhitespace)
	if err != nil {
		return nil, apimcp.SummarizeDiffOutput{}, err
	}

	ctxN := in.ContextLines
	gitRes, err := git.DiffNoIndex(ctx, leftAbs, rightAbs, git.DiffNoIndexOptions{
		Unified: &ctxN,
		Ignore:  ignore,
	})

	gitTrunc := false
	diffText := ""
	if gitRes != nil {
		diffText = gitRes.Diff
		gitTrunc = gitRes.Truncated
	}
	if err != nil {
		if errors.Is(err, gerr.ErrExecTruncated) {
			gitTrunc = true
		} else {
			return nil, apimcp.SummarizeDiffOutput{}, err
		}
	}

	parsed := diffparse.ParseUnified(diffText, opts)
	out := apimcp.BuildSummarizeOutput(parsed, "computed", gitTrunc)
	return nil, out, nil
}
