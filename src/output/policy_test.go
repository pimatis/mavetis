package output

import (
	"strings"
	"testing"

	"github.com/Pimatis/mavetis/src/model"
)

func TestPolicyOutputsIncludeProfileAndZoneMetadata(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	report := model.Report{
		Meta:    model.DiffMeta{Mode: "staged"},
		Policy:  &model.Policy{Profile: "auth", FailOn: "critical", Zones: []model.PolicyZone{{Name: "critical", SeverityOffset: 2, FailOn: "low"}}},
		Summary: model.Summary{Files: 1, Findings: 1, Critical: 1},
		Findings: []model.Finding{{
			RuleID:          "token.claims.unchecked",
			Title:           "Token claims validation appears incomplete",
			Category:        "token",
			Severity:        "critical",
			BaseSeverity:    "medium",
			Confidence:      "high",
			Path:            "src/auth/service.go",
			Line:            12,
			Side:            "added",
			Zone:            "critical",
			EffectiveFailOn: "low",
			Message:         "Problem",
			Snippet:         "token.ParseWithClaims(raw)",
			Remediation:     "Fix",
			Standards:       []string{"OWASP-ASVS-V3.5"},
		}},
	}
	text := TextExplain(report, true)
	if !strings.Contains(text, "Profile: auth") || !strings.Contains(text, "Zone: critical") || !strings.Contains(text, "BaseSeverity: medium") {
		t.Fatalf("unexpected text output: %q", text)
	}
	jsonBody, err := JSON(report)
	if err != nil {
		t.Fatalf("json output: %v", err)
	}
	if !strings.Contains(jsonBody, `"profile": "auth"`) || !strings.Contains(jsonBody, `"zone": "critical"`) {
		t.Fatalf("unexpected json output: %q", jsonBody)
	}
	sarifBody, err := SARIF(report)
	if err != nil {
		t.Fatalf("sarif output: %v", err)
	}
	if !strings.Contains(sarifBody, `"profile": "auth"`) || !strings.Contains(sarifBody, `"zone": "critical"`) {
		t.Fatalf("unexpected sarif output: %q", sarifBody)
	}
}
