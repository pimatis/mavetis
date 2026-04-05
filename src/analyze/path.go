package analyze

import (
	"path/filepath"
	"strings"
)

func Extension(path string) string {
	return strings.ToLower(filepath.Ext(path))
}

func Language(path string) string {
	ext := Extension(path)
	if ext == ".go" {
		return "go"
	}
	if ext == ".js" || ext == ".jsx" || ext == ".mjs" || ext == ".cjs" {
		return "javascript"
	}
	if ext == ".ts" || ext == ".tsx" {
		return "typescript"
	}
	if ext == ".py" {
		return "python"
	}
	if ext == ".rb" {
		return "ruby"
	}
	if ext == ".java" || ext == ".kt" {
		return "jvm"
	}
	if ext == ".json" || ext == ".yaml" || ext == ".yml" || ext == ".toml" {
		return "config"
	}
	return "unknown"
}

func Executable(path string) bool {
	language := Language(path)
	if language == "unknown" {
		return false
	}
	if ReviewArtifact(path) {
		return false
	}
	if language == "config" {
		return true
	}
	return !Documentation(path)
}

func Documentation(path string) bool {
	normalized := strings.ToLower(filepath.ToSlash(path))
	if strings.HasSuffix(normalized, ".md") || strings.HasSuffix(normalized, ".rst") || strings.HasSuffix(normalized, ".txt") {
		return true
	}
	if strings.Contains(normalized, "/docs/") {
		return true
	}
	return false
}

func Fixture(path string) bool {
	normalized := strings.ToLower(filepath.ToSlash(path))
	if strings.HasSuffix(normalized, "_test.go") {
		return true
	}
	if strings.Contains(normalized, "/testdata/") {
		return true
	}
	if strings.Contains(normalized, "/fixtures/") {
		return true
	}
	if strings.Contains(normalized, "/examples/") {
		return true
	}
	return false
}

func ReviewArtifact(path string) bool {
	normalized := strings.ToLower(filepath.ToSlash(path))
	if Documentation(path) {
		return true
	}
	if Fixture(path) {
		return true
	}
	if strings.Contains(normalized, "/src/rule/") || strings.HasPrefix(normalized, "src/rule/") {
		return true
	}
	if strings.Contains(normalized, "/src/analyze/") || strings.HasPrefix(normalized, "src/analyze/") {
		return true
	}
	return false
}
