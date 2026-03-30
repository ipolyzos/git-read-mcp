package errors

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"strings"

	"github.com/ipolyzos/git-read-mcp/internal/validate"
)

const maxPublicMsgLen = 2048

// Error is a tool-safe, JSON-serializable failure. Error() returns a single JSON object line for MCP text content.
type Error struct {
	Code    Code   `json:"code"`
	Message string `json:"message"`
	cause   error
}

func (e *Error) Error() string {
	if e == nil {
		return ""
	}
	b, err := json.Marshal(struct {
		Code    Code   `json:"code"`
		Message string `json:"message"`
	}{e.Code, e.Message})
	if err != nil {
		return `{"code":"INTERNAL","message":"failed to marshal error"}`
	}
	return string(b)
}

// Unwrap returns the underlying error when present.
func (e *Error) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.cause
}

// NewError builds a coded error without wrapping.
func NewError(code Code, message string) *Error {
	return &Error{Code: code, Message: trimPublicMessage(message), cause: nil}
}

// WrapError attaches a cause for logging/unwrapping; the message is client-facing.
func WrapError(code Code, message string, cause error) *Error {
	return &Error{Code: code, Message: trimPublicMessage(message), cause: cause}
}

// Map classifies any error into a stable *Error. A nil input yields nil.
func Map(err error) *Error {
	if err == nil {
		return nil
	}
	var ce *Error
	if errors.As(err, &ce) {
		return ce
	}

	code, msg := classify(err)
	return &Error{Code: code, Message: trimPublicMessage(msg), cause: err}
}

func classify(err error) (Code, string) {
	switch {
	case errors.Is(err, context.Canceled), errors.Is(err, context.DeadlineExceeded):
		return CodeTimeout, "operation timed out or was canceled"
	case errors.Is(err, ErrExecTimeout):
		return CodeTimeout, "git subprocess timed out"
	case errors.Is(err, ErrExecTruncated):
		return CodeTooLarge, "git output exceeded the configured byte cap"
	case errors.Is(err, validate.ErrOutsideRepo),
		errors.Is(err, validate.ErrNotDirectory),
		errors.Is(err, validate.ErrNotAbsolute),
		errors.Is(err, validate.ErrStringTooLong),
		errors.Is(err, validate.ErrIntOutOfRange):
		return CodeInvalidArgument, err.Error()
	}

	if errors.Is(err, os.ErrNotExist) {
		return CodeNotFound, err.Error()
	}

	if nz, ok := IsNonZeroExit(err); ok {
		s := strings.TrimSpace(string(nz.Stderr))
		if s == "" {
			s = nz.Error()
		}
		return CodeGitError, s
	}

	msg := err.Error()
	lower := strings.ToLower(msg)

	if strings.HasPrefix(msg, "validate:") ||
		(strings.HasPrefix(msg, "git: ") && !strings.Contains(lower, "timed out")) ||
		strings.HasPrefix(msg, "summarize_diff:") ||
		strings.HasPrefix(msg, "ignore_whitespace:") {
		return CodeInvalidArgument, msg
	}

	if strings.Contains(lower, "no space left") {
		return CodeInternal, "temporary workspace error"
	}

	if strings.Contains(msg, "executable file not found") {
		return CodeGitError, "git executable not available or failed to start"
	}

	if strings.HasPrefix(msg, "execx:") && strings.Contains(msg, "pipe") {
		return CodeInternal, "failed to capture subprocess output"
	}

	if strings.HasPrefix(msg, "execx:") {
		return CodeInternal, msg
	}

	return CodeInternal, msg
}

func trimPublicMessage(s string) string {
	s = strings.TrimSpace(s)
	if len(s) <= maxPublicMsgLen {
		return s
	}
	return s[:maxPublicMsgLen] + "…"
}
