package engine

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Pimatis/mavetis/src/diff"
	"github.com/Pimatis/mavetis/src/model"
	"github.com/Pimatis/mavetis/src/rule"
)

func TestRealWorldFalsePositiveFixtureStaysQuiet(t *testing.T) {
	content, err := os.ReadFile(filepath.Join("testdata", "realfp.diff"))
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	parsed, err := diff.Parse(string(content), model.DiffMeta{Mode: "staged"})
	if err != nil {
		t.Fatalf("parse fixture: %v", err)
	}
	report, err := Review(parsed, model.Config{Severity: "low"}, rule.Builtins(model.Config{}))
	if err != nil {
		t.Fatalf("review fixture: %v", err)
	}
	for _, finding := range report.Findings {
		if finding.RuleID == "inject.traversal" || finding.RuleID == "inject.sql.raw" || finding.RuleID == "crypto.compare.missing" || finding.RuleID == "inject.ssrf.fetch" {
			t.Fatalf("unexpected false positive: %#v", finding)
		}
	}
}
