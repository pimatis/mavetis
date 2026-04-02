package engine

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Pimatis/mavetis/src/diff"
	"github.com/Pimatis/mavetis/src/model"
	"github.com/Pimatis/mavetis/src/rule"
)

func TestFixtureBranchRegression(t *testing.T) {
	content, err := os.ReadFile(filepath.Join("testdata", "branch.diff"))
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	parsed, err := diff.Parse(string(content), model.DiffMeta{Mode: "branch"})
	if err != nil {
		t.Fatalf("parse fixture: %v", err)
	}
	report, err := Review(parsed, model.Config{Severity: "low"}, rule.Builtins(model.Config{}))
	if err != nil {
		t.Fatalf("review fixture: %v", err)
	}
	if !hasRule(report, "branch.guard.regression") {
		t.Fatalf("expected branch regression finding, got %#v", report.Findings)
	}
}
