package output

import (
	"testing"

	"github.com/Pimatis/mavetis/src/model"
)

func TestGoldenOutputs(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	report := model.Report{
		Meta:     model.DiffMeta{Mode: "staged"},
		Summary:  model.Summary{Files: 1, Findings: 1, High: 1},
		Rules:    []model.RuleInfo{{ID: "rule.demo", Title: "Demo", Category: "auth", Severity: "high", Standards: []string{"OWASP-ASVS"}}},
		Findings: []model.Finding{{RuleID: "rule.demo", Title: "Demo", Category: "auth", Severity: "high", Confidence: "high", Path: "app.go", Line: 7, Side: "added", Message: "Problem", Snippet: "bad()", Remediation: "Fix", Standards: []string{"OWASP-ASVS"}}},
	}
	text := TextExplain(report, true)
	jsonBody, err := JSON(report)
	if err != nil {
		t.Fatalf("json output: %v", err)
	}
	sarifBody, err := SARIF(report)
	if err != nil {
		t.Fatalf("sarif output: %v", err)
	}
	if text == "" || jsonBody == "" || sarifBody == "" {
		t.Fatal("expected non-empty golden outputs")
	}
	if text[:13] != "Mode: staged\n" {
		t.Fatalf("unexpected text output: %q", text)
	}
	if jsonBody[:2] != "{\n" {
		t.Fatalf("unexpected json output: %q", jsonBody)
	}
	if sarifBody[:2] != "{\n" {
		t.Fatalf("unexpected sarif output: %q", sarifBody)
	}
}
