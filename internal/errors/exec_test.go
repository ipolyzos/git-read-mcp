package errors

import (
	"testing"
)

func TestFromExecResult_priority(t *testing.T) {
	t.Parallel()
	if err := FromExecResult(ExecResult{TimedOut: true, Truncated: true, ExitCode: 9}); !IsTimeout(err) {
		t.Fatalf("timeout should win: %v", err)
	}
	if err := FromExecResult(ExecResult{Truncated: true, ExitCode: 9}); !IsTruncated(err) {
		t.Fatalf("truncation expected: %v", err)
	}
	if err := FromExecResult(ExecResult{ExitCode: 4, Stderr: []byte("x")}); err == nil {
		t.Fatal("expected non-zero exit mapping")
	} else if nz, ok := IsNonZeroExit(err); !ok || nz.Code != 4 {
		t.Fatalf("nonzero unwrap: ok=%v err=%v", ok, err)
	}
	if err := FromExecResult(ExecResult{}); err != nil {
		t.Fatalf("want nil, got %v", err)
	}
}
