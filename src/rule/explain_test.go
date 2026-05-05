package rule

import (
	"strings"
	"testing"

	"github.com/Pimatis/mavetis/src/model"
)

func TestExplainBuiltinRuleIncludesPatternsAndExamples(t *testing.T) {
	explanation, ok := Explain("inject.sql.raw", Builtins(model.Config{}))
	if !ok {
		t.Fatal("expected builtin explanation")
	}
	if explanation.Title != "Raw SQL query introduced" {
		t.Fatalf("unexpected title: %s", explanation.Title)
	}
	if !hasExplanationValue(explanation.Triggers, "required pattern:") {
		t.Fatalf("expected trigger patterns: %#v", explanation.Triggers)
	}
	if !hasExplanationValue(explanation.Standards, "OWASP-ASVS-V1.2") {
		t.Fatalf("expected ASVS mapping: %#v", explanation.Standards)
	}
	if !strings.Contains(explanation.VulnerableExample, "SELECT") {
		t.Fatalf("expected vulnerable example: %q", explanation.VulnerableExample)
	}
	if !strings.Contains(explanation.SafeExample, "?") {
		t.Fatalf("expected safe example: %q", explanation.SafeExample)
	}
}

func TestExplainSyntheticRuleIncludesSemanticTriggers(t *testing.T) {
	explanation, ok := Explain("semantic.go.ssrf", nil)
	if !ok {
		t.Fatal("expected synthetic explanation")
	}
	if explanation.Type != "synthetic" {
		t.Fatalf("unexpected type: %s", explanation.Type)
	}
	if !hasExplanationValue(explanation.Triggers, "http.Get") {
		t.Fatalf("expected semantic sink trigger: %#v", explanation.Triggers)
	}
	if !hasExplanationValue(explanation.NegativeContext, "no regex absent guard") {
		t.Fatalf("expected negative context: %#v", explanation.NegativeContext)
	}
	if !hasExplanationValue(explanation.Standards, "OWASP-ASVS-V4.3") {
		t.Fatalf("expected ASVS mapping: %#v", explanation.Standards)
	}
}

func TestExplainCustomRuleUsesCustomExamples(t *testing.T) {
	rules := []model.Rule{{
		ID:                "company.custom",
		Title:             "Company custom rule",
		Message:           "Custom rule fired.",
		Remediation:       "Use the approved helper.",
		Category:          "authorization",
		Severity:          "high",
		Confidence:        "medium",
		Target:            "added",
		Require:           []string{"dangerous"},
		VulnerableExample: "dangerous(userInput)",
		SafeExample:       "approved(userInput)",
	}}
	explanation, ok := Explain("company.custom", rules)
	if !ok {
		t.Fatal("expected custom rule explanation")
	}
	if explanation.VulnerableExample != "dangerous(userInput)" {
		t.Fatalf("unexpected vulnerable example: %q", explanation.VulnerableExample)
	}
	if explanation.SafeExample != "approved(userInput)" {
		t.Fatalf("unexpected safe example: %q", explanation.SafeExample)
	}
}

func hasExplanationValue(values []string, expected string) bool {
	for _, value := range values {
		if strings.Contains(value, expected) {
			return true
		}
	}
	return false
}
