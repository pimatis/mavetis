package engine

import (
	"testing"

	"github.com/Pimatis/mavetis/src/model"
	"github.com/Pimatis/mavetis/src/rule"
)

func TestReviewDetectsSecurityDowngrades(t *testing.T) {
	diff := model.Diff{
		Files: []model.DiffFile{{
			Path: "auth/session.ts",
			Hunks: []model.DiffHunk{{
				Lines: []model.DiffLine{
					{Kind: "deleted", Text: `sameSite: "strict"`, OldNumber: 1},
					{Kind: "added", Text: `sameSite: "lax"`, NewNumber: 1},
					{Kind: "deleted", Text: `Set-Cookie: sid=abc; Max-Age=3600; HttpOnly`, OldNumber: 2},
					{Kind: "added", Text: `Set-Cookie: sid=abc; Max-Age=86400; HttpOnly`, NewNumber: 2},
					{Kind: "deleted", Text: `bcryptCost = 12`, OldNumber: 3},
					{Kind: "added", Text: `bcryptCost = 8`, NewNumber: 3},
					{Kind: "deleted", Text: `maxAttempts = 5`, OldNumber: 4},
					{Kind: "added", Text: `maxAttempts = 20`, NewNumber: 4},
					{Kind: "deleted", Text: `sessionTimeout = 15m`, OldNumber: 5},
					{Kind: "added", Text: `sessionTimeout = 24h`, NewNumber: 5},
					{Kind: "deleted", Text: `mfaRequired = true`, OldNumber: 6},
					{Kind: "added", Text: `mfaRequired = false`, NewNumber: 6},
				},
			}},
		}},
	}
	report, err := Review(diff, model.Config{Severity: "low"}, rule.Builtins(model.Config{}))
	if err != nil {
		t.Fatalf("review failed: %v", err)
	}
	expected := []string{
		"downgrade.cookie.samesite",
		"downgrade.cookie.lifetime",
		"downgrade.crypto.bcrypt",
		"downgrade.auth.ratelimit",
		"downgrade.timeout",
		"downgrade.auth.mfa",
		"auth.mfa.disabled",
	}
	for _, id := range expected {
		if !hasFinding(report.Findings, id) {
			t.Fatalf("expected finding %s in %#v", id, report.Findings)
		}
	}
}

func TestReviewAvoidsCacheLifetimeFalsePositive(t *testing.T) {
	diff := model.Diff{
		Files: []model.DiffFile{{
			Path: "proxy/nginx.conf",
			Hunks: []model.DiffHunk{{
				Lines: []model.DiffLine{
					{Kind: "deleted", Text: `Cache-Control: public, max-age=60`, OldNumber: 10},
					{Kind: "added", Text: `Cache-Control: public, max-age=600`, NewNumber: 10},
				},
			}},
		}},
	}
	report, err := Review(diff, model.Config{Severity: "low"}, rule.Builtins(model.Config{}))
	if err != nil {
		t.Fatalf("review failed: %v", err)
	}
	if hasFinding(report.Findings, "downgrade.cookie.lifetime") {
		t.Fatalf("unexpected cookie lifetime finding: %#v", report.Findings)
	}
	if hasFinding(report.Findings, "downgrade.timeout") {
		t.Fatalf("unexpected timeout finding: %#v", report.Findings)
	}
}

func hasFinding(findings []model.Finding, ruleID string) bool {
	for _, item := range findings {
		if item.RuleID == ruleID {
			return true
		}
	}
	return false
}
