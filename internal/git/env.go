package git

import (
	"os"
	"sort"
	"strings"
)

// EnvGitReadOnly returns the host environment overlaid with read-only friendly defaults.
// It never invokes a shell; values are fixed literals only.
func EnvGitReadOnly() []string {
	return mergeEnv(os.Environ(), map[string]string{
		"GIT_PAGER":           "cat",
		"PAGER":               "cat",
		"GIT_TERMINAL_PROMPT": "0",
		"LC_ALL":              "C",
	})
}

func mergeEnv(base []string, overlay map[string]string) []string {
	parsed := make(map[string]string)
	for _, kv := range base {
		key, val, ok := strings.Cut(kv, "=")
		if !ok {
			continue
		}
		parsed[key] = val
	}
	for k, v := range overlay {
		parsed[k] = v
	}
	keys := make([]string, 0, len(parsed))
	for k := range parsed {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	out := make([]string, len(keys))
	for i, k := range keys {
		out[i] = k + "=" + parsed[k]
	}
	return out
}
