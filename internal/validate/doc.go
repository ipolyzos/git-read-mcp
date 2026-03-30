// Package validate provides path and input guards for MCP tools.
//
// Path helpers assume callers normalize the repository root once via NormalizeRepoPath,
// then use ResolvePath/EnsureWithinRepo according to each tool's contract (see docs/tools.md).
package validate
