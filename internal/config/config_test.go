package config

import (
	"testing"
	"time"
)

func resetGlobals() {
	mu.Lock()
	current = defaultConfiguration()
	mu.Unlock()
}

func TestLoadFromEnv_invalidDuration(t *testing.T) {
	t.Cleanup(resetGlobals)
	t.Setenv(envDiffTimeout, "not-a-duration")
	if err := LoadFromEnv(); err == nil {
		t.Fatal("expected error")
	}
}

func TestLoadFromEnv_validOverrides(t *testing.T) {
	t.Cleanup(resetGlobals)
	t.Setenv(envDiffTimeout, "45s")
	t.Setenv(envMaxOutputBytes, "100000")
	if err := LoadFromEnv(); err != nil {
		t.Fatal(err)
	}
	c := Get()
	if c.DiffTimeout != 45*time.Second {
		t.Fatalf("DiffTimeout: %v", c.DiffTimeout)
	}
	if c.MaxOutputBytes != 100000 {
		t.Fatalf("MaxOutputBytes: %d", c.MaxOutputBytes)
	}
}
