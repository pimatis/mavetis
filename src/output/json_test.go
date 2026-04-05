package output

import (
	"strings"
	"testing"

	"github.com/Pimatis/mavetis/src/model"
)

func TestJSONOmitsUnmatchedRules(t *testing.T) {
	report := model.Report{
		Meta:    model.DiffMeta{Mode: "file"},
		Summary: model.Summary{Files: 1, Findings: 1, High: 1},
		Rules: []model.RuleInfo{
			{ID: "rule.keep", Title: "Keep", Category: "auth", Severity: "high"},
			{ID: "rule.drop", Title: "Drop", Category: "auth", Severity: "high"},
		},
		Findings: []model.Finding{{RuleID: "rule.keep", Title: "Keep", Severity: "high", Path: "app.go", Line: 1, Side: "added", Message: "Problem"}},
	}
	body, err := JSON(report)
	if err != nil {
		t.Fatalf("json output: %v", err)
	}
	if !strings.Contains(body, `"rule.keep"`) {
		t.Fatalf("expected matched rule, got %q", body)
	}
	if strings.Contains(body, `"rule.drop"`) {
		t.Fatalf("expected unmatched rule omission, got %q", body)
	}
}

func TestJSONOmitsRulesWhenNoFindings(t *testing.T) {
	report := model.Report{
		Meta:    model.DiffMeta{Mode: "file"},
		Summary: model.Summary{Files: 1},
		Rules:   []model.RuleInfo{{ID: "rule.demo", Title: "Demo", Category: "auth", Severity: "high"}},
	}
	body, err := JSON(report)
	if err != nil {
		t.Fatalf("json output: %v", err)
	}
	if strings.Contains(body, `"rules"`) {
		t.Fatalf("expected rules omission, got %q", body)
	}
}
