package output

import (
	"strings"
	"testing"

	"github.com/Pimatis/mavetis/src/model"
)

func TestTextUsesColorWhenForced(t *testing.T) {
	t.Setenv("FORCE_COLOR", "1")
	t.Setenv("NO_COLOR", "")
	t.Setenv("TERM", "xterm-256color")
	report := model.Report{
		Meta:    model.DiffMeta{Mode: "staged"},
		Summary: model.Summary{Files: 1, Findings: 1, Critical: 1},
		Findings: []model.Finding{{
			RuleID:      "rule.demo",
			Title:       "Critical finding",
			Severity:    "critical",
			Path:        "app.go",
			Line:        12,
			Side:        "added",
			Message:     "Problem",
			Snippet:     "secret=value",
			Remediation: "Fix it",
		}},
	}
	body := Text(report)
	if !strings.Contains(body, "\033[1;31m") {
		t.Fatalf("expected critical color, got %q", body)
	}
}

func TestTextDisablesColorWithNoColor(t *testing.T) {
	t.Setenv("FORCE_COLOR", "")
	t.Setenv("NO_COLOR", "1")
	t.Setenv("TERM", "xterm-256color")
	report := model.Report{
		Meta:    model.DiffMeta{Mode: "staged"},
		Summary: model.Summary{Files: 1},
	}
	body := Text(report)
	if strings.Contains(body, "\033[") {
		t.Fatalf("expected plain text output, got %q", body)
	}
}
