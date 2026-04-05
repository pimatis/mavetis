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

func TestTextRendersSuggestions(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	report := model.Report{
		Meta:             model.DiffMeta{Mode: "file"},
		Summary:          model.Summary{Files: 1},
		Suggestions:      []model.Suggestion{{Path: "src/auth.ts", From: "src/app.ts", Reason: "imported", Depth: 1}},
		SuggestedCommand: "mavetis review src/app.ts --with-suggested",
	}
	body := Text(report)
	if !strings.Contains(body, "Suggested: 1 additional files to review") {
		t.Fatalf("expected suggestion summary, got %q", body)
	}
	if !strings.Contains(body, "src/auth.ts (imported from src/app.ts)") {
		t.Fatalf("expected suggestion detail, got %q", body)
	}
	if !strings.Contains(body, "Run: mavetis review src/app.ts --with-suggested") {
		t.Fatalf("expected suggested command, got %q", body)
	}
}

func TestTextRendersReviewedSuggestions(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	report := model.Report{
		Meta:        model.DiffMeta{Mode: "file"},
		Summary:     model.Summary{Files: 2},
		Suggestions: []model.Suggestion{{Path: "src/auth.ts", From: "src/app.ts", Reason: "imported", Depth: 1, Reviewed: true}},
	}
	body := Text(report)
	if !strings.Contains(body, "Included: 1 additional files reviewed") {
		t.Fatalf("expected reviewed summary, got %q", body)
	}
	if !strings.Contains(body, "src/auth.ts (imported from src/app.ts; reviewed)") {
		t.Fatalf("expected reviewed detail, got %q", body)
	}
}
