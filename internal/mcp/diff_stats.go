package mcp

import "strings"

// DiffLineStatsFromUnifiedDiff counts non-header + and - lines in unified diff output.
func DiffLineStatsFromUnifiedDiff(diff string) *DiffLineStats {
	if diff == "" {
		return &DiffLineStats{}
	}
	var ins, del int
	for _, line := range strings.Split(diff, "\n") {
		if len(line) == 0 {
			continue
		}
		switch line[0] {
		case '+':
			if strings.HasPrefix(line, "+++") {
				continue
			}
			ins++
		case '-':
			if strings.HasPrefix(line, "---") {
				continue
			}
			del++
		}
	}
	return &DiffLineStats{Insertions: ins, Deletions: del}
}
