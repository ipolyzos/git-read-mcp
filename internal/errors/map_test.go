package errors

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/ipolyzos/git-read-mcp/internal/validate"
)

func TestMap_table(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		err  error
		want Code
	}{
		{"nil", nil, ""},
		{"timeout_exec", ErrExecTimeout, CodeTimeout},
		{"truncated_exec", ErrExecTruncated, CodeTooLarge},
		{"ctx_deadline", context.DeadlineExceeded, CodeTimeout},
		{"ctx_canceled", context.Canceled, CodeTimeout},
		{"outside_repo", validate.ErrOutsideRepo, CodeInvalidArgument},
		{"not_exist", os.ErrNotExist, CodeNotFound},
		{"nonzero", NonZeroExit(2, []byte("remote: rejected")), CodeGitError},
		{"ignore_whitespace", errors.New("ignore_whitespace: want none"), CodeInvalidArgument},
		{"validate_prefix", errors.New("validate: repo_path is empty"), CodeInvalidArgument},
		{"git_arg", errors.New("git: left path is empty"), CodeInvalidArgument},
		{"unknown", errors.New("something weird"), CodeInternal},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := Map(tc.err)
			if tc.err == nil {
				if got != nil {
					t.Fatalf("want nil, got %+v", got)
				}
				return
			}
			if got == nil {
				t.Fatal("expected non-nil")
			}
			if got.Code != tc.want {
				t.Fatalf("code: got %q want %q (full %+v)", got.Code, tc.want, got)
			}
		})
	}
}

func TestError_JSONErrorString(t *testing.T) {
	t.Parallel()
	e := NewError(CodeInvalidArgument, `path "x" is bad`)
	raw := e.Error()
	var dec struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	}
	if err := json.Unmarshal([]byte(raw), &dec); err != nil {
		t.Fatalf("Error() not JSON: %v: %q", err, raw)
	}
	if dec.Code != string(CodeInvalidArgument) || !strings.Contains(dec.Message, "path") {
		t.Fatalf("decoded %+v from %q", dec, raw)
	}
}

func TestMap_idempotent(t *testing.T) {
	t.Parallel()
	orig := NewError(CodeTimeout, "x")
	again := Map(orig)
	if again != orig {
		t.Fatalf("expected same pointer identity for *Error")
	}
}
