package analyze

import (
	"regexp"
	"strings"
)

var sourceAssign = regexp.MustCompile(`(?i)\b([A-Za-z_][A-Za-z0-9_]*)\b\s*(:=|=).*(query|params|param|header|cookie|body|input|request|ctx\.(query|params)|urlparam|formvalue)`)
var sourceUse = regexp.MustCompile(`(?i)(query|params|param|header|cookie|body|input|request|ctx\.(query|params)|urlparam|formvalue)`)

func Tainted(text string) []string {
	matches := sourceAssign.FindAllStringSubmatch(text, -1)
	values := make([]string, 0, len(matches))
	for _, item := range matches {
		if len(item) < 2 {
			continue
		}
		values = append(values, strings.ToLower(item[1]))
	}
	return values
}

func TaintedUse(text string, vars []string) bool {
	if len(vars) == 0 {
		return sourceUse.MatchString(text)
	}
	lower := strings.ToLower(text)
	for _, item := range vars {
		if strings.Contains(lower, strings.ToLower(item)) {
			return true
		}
	}
	return false
}

func Guarded(text string) bool {
	lower := strings.ToLower(text)
	if strings.Contains(lower, "authorize") || strings.Contains(lower, "requireauth") {
		return true
	}
	if strings.Contains(lower, "requirerole") || strings.Contains(lower, "permission") {
		return true
	}
	if strings.Contains(lower, "tenant") || strings.Contains(lower, "owner") || strings.Contains(lower, "scope") {
		return true
	}
	if strings.Contains(lower, "validateredirect") || strings.Contains(lower, "allowlist") {
		return true
	}
	if strings.Contains(lower, "filepath.clean") || strings.Contains(lower, "path.clean") {
		return true
	}
	return false
}
