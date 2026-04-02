package engine

import (
	"testing"

	"github.com/Pimatis/mavetis/src/model"
	"github.com/Pimatis/mavetis/src/rule"
)

func TestFilterRulesForProfileKeepsOnlyRelevantRules(t *testing.T) {
	rules := []model.Rule{
		{ID: "auth.mfa.disabled", Category: "auth"},
		{ID: "observe.auth.material", Category: "logging"},
		{ID: "inject.command.exec", Category: "injection"},
		{ID: "inject.xss.innerhtml", Category: "xss"},
	}
	filtered := FilterRulesForProfile(rules, "auth")
	if len(filtered) != 2 {
		t.Fatalf("unexpected filtered rules: %#v", filtered)
	}
	if filtered[0].ID != "auth.mfa.disabled" || filtered[1].ID != "observe.auth.material" {
		t.Fatalf("unexpected filtered rule order: %#v", filtered)
	}
}

func TestApplyPolicyEscalatesSeverityAndFailOnForZones(t *testing.T) {
	finding := model.Finding{RuleID: "token.claims.unchecked", Severity: "medium"}
	zone := zoneMatch{name: "critical", severityOffset: 2, failOn: "low"}
	applied := applyPolicy(finding, zone, "critical")
	if applied.Zone != "critical" {
		t.Fatalf("unexpected zone: %#v", applied)
	}
	if applied.BaseSeverity != "medium" || applied.Severity != "critical" {
		t.Fatalf("unexpected severity adjustment: %#v", applied)
	}
	if applied.EffectiveFailOn != "low" {
		t.Fatalf("unexpected fail-on: %#v", applied)
	}
}

func TestResolveZoneUsesCachedStrongestZone(t *testing.T) {
	config := model.Config{Zones: model.Zones{Critical: []string{"src/auth/**"}, Restricted: []string{"src/**"}}}
	cache := map[string]zoneMatch{}
	zone := resolveZone("src/auth/service.go", config, cache)
	if zone.name != "critical" {
		t.Fatalf("expected critical zone, got %#v", zone)
	}
	cached := resolveZone("src/auth/service.go", config, cache)
	if cached.name != "critical" || len(cache) != 1 {
		t.Fatalf("expected cached critical zone, got %#v cache=%#v", cached, cache)
	}
}

func TestReviewAppliesProfileAndZonePolicy(t *testing.T) {
	config := model.Config{
		Severity: "low",
		FailOn:   "critical",
		Profile:  "auth",
		Zones:    model.Zones{Critical: []string{"src/auth/**"}},
	}
	diff := model.Diff{Files: []model.DiffFile{
		{Path: "src/auth/service.go", Hunks: []model.DiffHunk{{Lines: []model.DiffLine{{Kind: "added", Text: `token.ParseWithClaims(rawToken)`, NewNumber: 1}}}}},
		{Path: "src/ui/view.tsx", Hunks: []model.DiffHunk{{Lines: []model.DiffLine{{Kind: "added", Text: `element.innerHTML = body`, NewNumber: 1}}}}},
	}}
	rules := FilterRulesForProfile(rule.Builtins(config), config.Profile)
	report, err := Review(diff, config, rules)
	if err != nil {
		t.Fatalf("review failed: %v", err)
	}
	if len(report.Findings) != 1 {
		t.Fatalf("expected one auth finding, got %#v", report.Findings)
	}
	finding := report.Findings[0]
	if finding.RuleID != "token.claims.unchecked" {
		t.Fatalf("unexpected rule: %#v", finding)
	}
	if finding.Zone != "critical" || finding.BaseSeverity != "medium" || finding.Severity != "critical" {
		t.Fatalf("expected critical zone uplift, got %#v", finding)
	}
	if finding.EffectiveFailOn != "low" {
		t.Fatalf("expected aggressive fail-on, got %#v", finding)
	}
	if report.Policy == nil || report.Policy.Profile != "auth" {
		t.Fatalf("expected report policy, got %#v", report.Policy)
	}
}
