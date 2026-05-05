package output

import (
	"strings"
	"testing"

	"github.com/Pimatis/mavetis/src/model"
)

func TestRuleExplanationRendersRequiredSections(t *testing.T) {
	body := RuleExplanation(model.RuleExplanation{
		ID:                "inject.sql.raw",
		Title:             "Raw SQL query introduced",
		Severity:          "high",
		Confidence:        "medium",
		Category:          "injection",
		Target:            "added",
		Engine:            "regex rule engine",
		Message:           "String-built SQL was introduced.",
		Remediation:       "Use parameterized queries.",
		Standards:         []string{"OWASP-ASVS-V1.2", "OWASP-SQL"},
		Scope:             []string{"target side: added"},
		Triggers:          []string{"required pattern: select"},
		NegativeContext:   []string{"absent guard: parameterized query"},
		VulnerableExample: "query := \"SELECT \" + userInput",
		SafeExample:       "db.QueryContext(ctx, \"SELECT ?\", userInput)",
	})
	checks := []string{
		"Rule: inject.sql.raw",
		"ASVS mappings:\n  - OWASP-ASVS-V1.2",
		"Standards:\n  - OWASP-SQL",
		"Trigger patterns:\n  - required pattern: select",
		"Negative context / absent guards:",
		"Example vulnerable snippet:",
		"Example safe pattern:",
		"Remediation: Use parameterized queries.",
	}
	for _, check := range checks {
		if !strings.Contains(body, check) {
			t.Fatalf("expected %q in %q", check, body)
		}
	}
}
