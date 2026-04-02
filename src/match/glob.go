package match

import (
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
)

func Any(patterns []string, value string) bool {
	for _, pattern := range patterns {
		if Glob(pattern, value) {
			return true
		}
	}
	return false
}

var cache sync.Map

func Glob(pattern string, value string) bool {
	normalizedPattern := filepath.ToSlash(strings.TrimSpace(pattern))
	normalizedValue := filepath.ToSlash(strings.TrimSpace(value))
	if normalizedPattern == "" {
		return false
	}
	if normalizedPattern == normalizedValue {
		return true
	}
	if !strings.Contains(normalizedPattern, "**") {
		matched, err := path.Match(normalizedPattern, normalizedValue)
		if err == nil && matched {
			return true
		}
	}
	re := compiled(normalizedPattern)
	return re.MatchString(normalizedValue)
}

func compiled(pattern string) *regexp.Regexp {
	cached, ok := cache.Load(pattern)
	if ok {
		return cached.(*regexp.Regexp)
	}
	expression := regexp.MustCompile("^" + quote(pattern) + "$")
	stored, _ := cache.LoadOrStore(pattern, expression)
	return stored.(*regexp.Regexp)
}

func quote(pattern string) string {
	builder := strings.Builder{}
	for index := 0; index < len(pattern); index++ {
		char := pattern[index]
		if char == '*' {
			if index+2 < len(pattern) && pattern[index+1] == '*' && pattern[index+2] == '/' {
				builder.WriteString("(?:.*/)?")
				index += 2
				continue
			}
			if index+1 < len(pattern) && pattern[index+1] == '*' {
				builder.WriteString(".*")
				index++
				continue
			}
			builder.WriteString("[^/]*")
			continue
		}
		if char == '?' {
			builder.WriteString(".")
			continue
		}
		builder.WriteString(regexp.QuoteMeta(string(char)))
	}
	return builder.String()
}
