package errors

import (
	"errors"
	"fmt"
)

// Sentinel errors for classifying exec outcomes.
var (
	ErrExecTimeout   = errors.New("exec: timed out")
	ErrExecTruncated = errors.New("exec: output truncated")
	ErrExecNonZero   = errors.New("exec: non-zero exit")
)

// NonZeroExitError carries exit metadata for git or other CLIs.
type NonZeroExitError struct {
	Code   int
	Stderr []byte
}

func (e *NonZeroExitError) Error() string {
	return fmt.Sprintf("exec: exit %d", e.Code)
}

func (e *NonZeroExitError) Unwrap() error { return ErrExecNonZero }

// IsTimeout reports whether err is or wraps ErrExecTimeout.
func IsTimeout(err error) bool {
	return errors.Is(err, ErrExecTimeout)
}

// IsTruncated reports whether err is or wraps ErrExecTruncated.
func IsTruncated(err error) bool {
	return errors.Is(err, ErrExecTruncated)
}

// IsNonZeroExit returns (*NonZeroExitError, true) when err unwraps to that type.
func IsNonZeroExit(err error) (*NonZeroExitError, bool) {
	var nz *NonZeroExitError
	if errors.As(err, &nz) {
		return nz, true
	}
	return nil, false
}

// NonZeroExit builds a NonZeroExitError when code is non-zero.
func NonZeroExit(code int, stderr []byte) error {
	if code == 0 {
		return nil
	}
	return &NonZeroExitError{Code: code, Stderr: append([]byte(nil), stderr...)}
}

// ExecResult is the minimal surface execx results expose to errors without an import cycle.
type ExecResult struct {
	TimedOut  bool
	Truncated bool
	ExitCode  int
	Stderr    []byte
}

// FromExecResult maps a finished run into a single error suitable for logging or APIs.
// Priority: timeout, then truncation, then non-zero exit. Returns nil when none apply.
func FromExecResult(res ExecResult) error {
	if res.TimedOut {
		return ErrExecTimeout
	}
	if res.Truncated {
		return ErrExecTruncated
	}
	return NonZeroExit(res.ExitCode, res.Stderr)
}
