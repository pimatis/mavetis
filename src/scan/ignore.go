package scan

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"

	"github.com/Pimatis/mavetis/src/match"
)

func LoadGitignorePatterns(root string) []string {
	path := filepath.Join(root, ".gitignore")
	file, err := os.Open(path)
	if err != nil {
		return nil
	}
	defer file.Close()
	patterns := make([]string, 0)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		trimmed := strings.TrimPrefix(line, "!")
		patterns = append(patterns, trimmed)
	}
	return patterns
}

func IsGitignored(patterns []string, relPath string) bool {
	if len(patterns) == 0 {
		return false
	}
	normalized := filepath.ToSlash(relPath)
	for _, pattern := range patterns {
		if gitignoreMatches(pattern, normalized) {
			return true
		}
	}
	return false
}

func gitignoreMatches(pattern string, path string) bool {
	pattern = filepath.ToSlash(pattern)
	normalized := strings.TrimSuffix(pattern, "/")
	if strings.HasPrefix(pattern, "/") {
		normalized = strings.TrimPrefix(normalized, "/")
		if hasPattern(normalized) {
			return match.Glob(normalized, path)
		}
		return strings.HasPrefix(path+"/", normalized+"/") || path == normalized
	}
	if match.Glob("**/"+normalized, path) {
		return true
	}
	if match.Glob(normalized, path) {
		return true
	}
	if match.Glob("**/"+normalized+"/**", path) {
		return true
	}
	return false
}
