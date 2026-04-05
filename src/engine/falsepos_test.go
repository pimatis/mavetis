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

func TestReviewSkipsSelfReviewArtifacts(t *testing.T) {
	diff := model.Diff{Files: []model.DiffFile{{
		Path: "src/engine/review_test.go",
		Hunks: []model.DiffHunk{{
			Lines: []model.DiffLine{{Kind: "added", Text: `http.Get(target)`, NewNumber: 1}},
		}},
	}}}
	report, err := Review(diff, model.Config{Severity: "low"}, rule.Builtins(model.Config{}))
	if err != nil {
		t.Fatalf("review failed: %v", err)
	}
	if len(report.Findings) != 0 {
		t.Fatalf("expected self-review artifact skip, got %#v", report.Findings)
	}
}

func TestObserveRequestBodyDoesNotFlagRenderBodyVariable(t *testing.T) {
	diff := model.Diff{Files: []model.DiffFile{{
		Path: "src/cli/run.go",
		Hunks: []model.DiffHunk{{
			Lines: []model.DiffLine{{Kind: "added", Text: `fmt.Println(body)`, NewNumber: 1}},
		}},
	}}}
	report, err := Review(diff, model.Config{Severity: "low"}, rule.Builtins(model.Config{}))
	if err != nil {
		t.Fatalf("review failed: %v", err)
	}
	for _, finding := range report.Findings {
		if finding.RuleID == "observe.request.body" {
			t.Fatalf("unexpected request body finding: %#v", finding)
		}
	}
}

func TestTemplateRuleDoesNotFlagGenericFlagParsing(t *testing.T) {
	diff := model.Diff{Files: []model.DiffFile{{
		Path: "src/cli/args.go",
		Hunks: []model.DiffHunk{{
			Lines: []model.DiffLine{
				{Kind: "added", Text: `flagArguments := []string{"body"}`, NewNumber: 1},
				{Kind: "added", Text: `if err := flags.Parse(flagArguments); err != nil {`, NewNumber: 2},
			},
		}},
	}}}
	report, err := Review(diff, model.Config{Severity: "low"}, rule.Builtins(model.Config{}))
	if err != nil {
		t.Fatalf("review failed: %v", err)
	}
	for _, finding := range report.Findings {
		if finding.RuleID == "template.ssti.dynamic" {
			t.Fatalf("unexpected template rule finding: %#v", finding)
		}
	}
}
