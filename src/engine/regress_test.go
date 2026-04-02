package engine

import (
	"testing"

	"github.com/Pimatis/mavetis/src/model"
	"github.com/Pimatis/mavetis/src/rule"
)

func TestReviewDetectsRegressionGuardFindings(t *testing.T) {
	diff := model.Diff{
		Files: []model.DiffFile{{
			Path: "src/auth/login.ts",
			Hunks: []model.DiffHunk{{
				Lines: []model.DiffLine{
					{Kind: "added", Text: `mfaRequired = false`, NewNumber: 10},
					{Kind: "deleted", Text: `verifyTOTP(code)`, OldNumber: 11},
					{Kind: "deleted", Text: `loginRateLimit(userID)`, OldNumber: 12},
				},
			}},
		}},
	}
	report, err := Review(diff, model.Config{Severity: "low"}, rule.Builtins(model.Config{}))
	if err != nil {
		t.Fatalf("review failed: %v", err)
	}
	expected := []string{
		"auth.mfa.disabled",
		"auth.mfa.deleted",
		"auth.ratelimit.deleted",
	}
	for _, id := range expected {
		if !hasFinding(report.Findings, id) {
			t.Fatalf("expected finding %s in %#v", id, report.Findings)
		}
	}
}
