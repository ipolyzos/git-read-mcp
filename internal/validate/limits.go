package validate

import (
	"fmt"
)

// MaxStringBytes enforces len(s) <= maxBytes. maxBytes < 0 means unlimited.
func MaxStringBytes(label, s string, maxBytes int) error {
	if maxBytes < 0 {
		return nil
	}
	if label == "" {
		label = "value"
	}
	if len(s) > maxBytes {
		return fmt.Errorf("validate: %s exceeds %d bytes: %w", label, maxBytes, ErrStringTooLong)
	}
	return nil
}

// IntInRange verifies v is within [min, max] inclusive.
func IntInRange(label string, v, min, max int) error {
	if label == "" {
		label = "value"
	}
	if v < min || v > max {
		return fmt.Errorf("validate: %s=%d not in [%d,%d]: %w", label, v, min, max, ErrIntOutOfRange)
	}
	return nil
}

// BoolDefault returns def when p is nil, otherwise *p.
func BoolDefault(p *bool, def bool) bool {
	if p == nil {
		return def
	}
	return *p
}
