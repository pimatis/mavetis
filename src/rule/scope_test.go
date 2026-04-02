package rule

import (
	"testing"

	"github.com/Pimatis/mavetis/src/match"
)

func TestCodeFilesClone(t *testing.T) {
	first := codeFiles()
	second := codeFiles()
	if len(first) == 0 || len(second) == 0 {
		t.Fatal("expected non-empty code file scopes")
	}
	first[0] = "changed"
	if second[0] == "changed" {
		t.Fatal("expected clone to isolate callers")
	}
}

func TestScopePatternsMatchExpectedFiles(t *testing.T) {
	if !match.Any(codeFiles(), "service/auth.go") {
		t.Fatal("expected code scope to match go file")
	}
	if match.Any(codeFiles(), "README.md") {
		t.Fatal("expected code scope to skip markdown")
	}
	if !match.Any(workflowFiles(), ".github/workflows/review.yml") {
		t.Fatal("expected workflow scope to match workflow file")
	}
	if !match.Any(manifestFiles(), "package.json") {
		t.Fatal("expected manifest scope to match package.json")
	}
}
