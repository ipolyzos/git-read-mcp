// Package config holds process-wide settings loaded from the environment.
package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	envDiffTimeout    = "GIT_READ_MCP_DIFF_TIMEOUT"
	envMaxOutputBytes = "GIT_READ_MCP_MAX_OUTPUT_BYTES"
	envProbeTimeout   = "GIT_READ_MCP_GIT_PROBE_TIMEOUT"
	envProbeMaxBytes  = "GIT_READ_MCP_GIT_PROBE_MAX_BYTES"
)

const (
	defaultDiffTimeout    = 30 * time.Second
	defaultMaxOutputBytes = 8 << 20
	defaultProbeTimeout   = 5 * time.Second
	defaultProbeMaxBytes  = int64(64 * 1024)
	minDiffTimeout        = time.Second
	minGitProbeTimeout    = 100 * time.Millisecond
	minMaxOutputBytes     = 4096
	minGitProbeMaxBytes   = 1024
)

var (
	mu      sync.RWMutex
	current = defaultConfiguration()
)

// Config is a snapshot of server tuning knobs.
type Config struct {
	DiffTimeout      time.Duration
	MaxOutputBytes   int64
	GitProbeTimeout  time.Duration
	GitProbeMaxBytes int64
}

func defaultConfiguration() Config {
	return Config{
		DiffTimeout:      defaultDiffTimeout,
		MaxOutputBytes:   defaultMaxOutputBytes,
		GitProbeTimeout:  defaultProbeTimeout,
		GitProbeMaxBytes: defaultProbeMaxBytes,
	}
}

// LoadFromEnv reparses environment variables and replaces the global config.
// Call once from main before serving; safe for tests to call to reset.
func LoadFromEnv() error {
	mu.Lock()
	defer mu.Unlock()
	c, err := readEnv()
	if err != nil {
		return err
	}
	current = c
	return nil
}

// Get returns the current global configuration.
func Get() Config {
	mu.RLock()
	defer mu.RUnlock()
	return current
}

func readEnv() (Config, error) {
	c := defaultConfiguration()
	var errs []string

	if v := strings.TrimSpace(os.Getenv(envDiffTimeout)); v != "" {
		d, err := time.ParseDuration(v)
		if err != nil || d < minDiffTimeout {
			errs = append(errs, fmt.Sprintf("%s must be a duration >= %s", envDiffTimeout, minDiffTimeout))
		} else {
			c.DiffTimeout = d
		}
	}

	if v := strings.TrimSpace(os.Getenv(envMaxOutputBytes)); v != "" {
		n, err := strconv.ParseInt(v, 10, 64)
		if err != nil || n < minMaxOutputBytes {
			errs = append(errs, fmt.Sprintf("%s must be an integer >= %d", envMaxOutputBytes, minMaxOutputBytes))
		} else {
			c.MaxOutputBytes = n
		}
	}

	if v := strings.TrimSpace(os.Getenv(envProbeTimeout)); v != "" {
		d, err := time.ParseDuration(v)
		if err != nil || d < minGitProbeTimeout {
			errs = append(errs, fmt.Sprintf("%s must be a duration >= %s", envProbeTimeout, minGitProbeTimeout))
		} else {
			c.GitProbeTimeout = d
		}
	}

	if v := strings.TrimSpace(os.Getenv(envProbeMaxBytes)); v != "" {
		n, err := strconv.ParseInt(v, 10, 64)
		if err != nil || n < minGitProbeMaxBytes {
			errs = append(errs, fmt.Sprintf("%s must be an integer >= %d", envProbeMaxBytes, minGitProbeMaxBytes))
		} else {
			c.GitProbeMaxBytes = n
		}
	}

	if len(errs) > 0 {
		return Config{}, fmt.Errorf("config: %s", strings.Join(errs, "; "))
	}
	return c, nil
}
