package mcp

import (
	"errors"
	"strings"
	"testing"

	"github.com/ipolyzos/git-read-mcp/internal/validate"
)

func TestValidateCompareTextInput_rightTooLarge(t *testing.T) {
	t.Parallel()
	repo := t.TempDir()
	left := strings.Repeat("a", 100)
	right := strings.Repeat("b", MaxCompareTextBytes+1)
	err := ValidateCompareTextInput(CompareTextInput{
		RepoPath:     repo,
		LeftText:     left,
		RightText:    right,
		ContextLines: 3,
	})
	if !errors.Is(err, validate.ErrStringTooLong) {
		t.Fatalf("got %v", err)
	}
}

func TestValidateCompareTextInput_leftTooLarge(t *testing.T) {
	t.Parallel()
	repo := t.TempDir()
	err := ValidateCompareTextInput(CompareTextInput{
		RepoPath:     repo,
		LeftText:     strings.Repeat("x", MaxCompareTextBytes+1),
		RightText:    "y",
		ContextLines: 0,
	})
	if !errors.Is(err, validate.ErrStringTooLong) {
		t.Fatalf("got %v", err)
	}
}

func TestValidateCompareTextInput_ok(t *testing.T) {
	t.Parallel()
	repo := t.TempDir()
	if err := ValidateCompareTextInput(CompareTextInput{
		RepoPath:         repo,
		LeftText:         "a\n",
		RightText:        "b\n",
		ContextLines:     3,
		IgnoreWhitespace: "none",
	}); err != nil {
		t.Fatal(err)
	}
}

func TestValidateCompareTextInput_contextLines(t *testing.T) {
	t.Parallel()
	repo := t.TempDir()
	err := ValidateCompareTextInput(CompareTextInput{
		RepoPath:     repo,
		LeftText:     "a",
		RightText:    "b",
		ContextLines: 60,
	})
	if !errors.Is(err, validate.ErrIntOutOfRange) {
		t.Fatalf("got %v", err)
	}
}
