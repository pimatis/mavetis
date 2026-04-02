package analyze

import (
	"regexp"
	"strings"
)

type Step struct {
	Line   string
	Taints []string
}

var assignLine = regexp.MustCompile(`(?i)\b([A-Za-z_][A-Za-z0-9_]*)\b\s*(:=|=)\s*(.+)$`)

func Track(lines []string) []Step {
	state := map[string]struct{}{}
	steps := make([]Step, 0, len(lines))
	for _, line := range lines {
		lower := strings.ToLower(line)
		match := assignLine.FindStringSubmatch(line)
		if len(match) >= 4 {
			name := strings.ToLower(match[1])
			right := strings.ToLower(match[3])
			if sourceUse.MatchString(right) || containsTaint(right, state) {
				state[name] = struct{}{}
			}
			if Guarded(right) {
				delete(state, name)
			}
		}
		if Guarded(lower) {
			steps = append(steps, Step{Line: line, Taints: list(state)})
			continue
		}
		steps = append(steps, Step{Line: line, Taints: list(state)})
	}
	return steps
}

func containsTaint(value string, state map[string]struct{}) bool {
	for item := range state {
		if strings.Contains(value, item) {
			return true
		}
	}
	return false
}

func list(state map[string]struct{}) []string {
	values := make([]string, 0, len(state))
	for item := range state {
		values = append(values, item)
	}
	return values
}
