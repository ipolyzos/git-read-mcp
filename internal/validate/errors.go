package validate

import "errors"

var (
	// ErrNotAbsolute indicates a path was expected to be absolute.
	ErrNotAbsolute = errors.New("validate: path must be absolute")
	// ErrNotDirectory indicates a path exists but is not a directory.
	ErrNotDirectory = errors.New("validate: path is not a directory")
	// ErrOutsideRepo indicates a resolved path escapes the repository root.
	ErrOutsideRepo = errors.New("validate: path escapes repository root")
	// ErrStringTooLong indicates a string exceeded a byte-length cap.
	ErrStringTooLong = errors.New("validate: string exceeds maximum byte length")
	// ErrIntOutOfRange indicates an integer was outside an inclusive range.
	ErrIntOutOfRange = errors.New("validate: integer out of allowed range")
)
