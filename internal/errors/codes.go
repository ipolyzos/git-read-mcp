package errors

// Code is a stable machine-readable error class for tool results and logs.
type Code string

const (
	CodeInvalidArgument Code = "INVALID_ARGUMENT"
	CodeNotFound        Code = "NOT_FOUND"
	CodeTimeout         Code = "TIMEOUT"
	CodeTooLarge        Code = "TOO_LARGE"
	CodeGitError        Code = "GIT_ERROR"
	CodeInternal        Code = "INTERNAL"
)
