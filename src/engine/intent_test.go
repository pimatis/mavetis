package engine

import (
	"testing"

	"github.com/Pimatis/mavetis/src/model"
)

func TestIntentFindingsDetectMismatch(t *testing.T) {
	diff := model.Diff{Files: []model.DiffFile{{
		Path: "src/auth/service.ts",
		Hunks: []model.DiffHunk{{Lines: []model.DiffLine{
			{Kind: "deleted", Text: `function validateToken(rawToken) {`, OldNumber: 1},
			{Kind: "deleted", Text: `  verifySignature(rawToken)`, OldNumber: 2},
			{Kind: "deleted", Text: `  validateClaims(rawToken)`, OldNumber: 3},
			{Kind: "added", Text: `function validateToken(rawToken) {`, NewNumber: 1},
			{Kind: "added", Text: `  return decode(rawToken)`, NewNumber: 2},
		}}},
	}}}
	findings := intentFindings(diff)
	if len(findings) != 1 || findings[0].RuleID != "intent.mismatch" {
		t.Fatalf("unexpected intent findings: %#v", findings)
	}
}

func TestSnapshotFindingsDetectBaselineRegression(t *testing.T) {
	diff := model.Diff{Files: []model.DiffFile{{
		Path: "src/auth/service.ts",
		Hunks: []model.DiffHunk{{Lines: []model.DiffLine{
			{Kind: "deleted", Text: `function verifyOwnership(user, resource) {`, OldNumber: 1},
			{Kind: "deleted", Text: `  return queryOwner(user, resource)`, OldNumber: 2},
			{Kind: "added", Text: `function verifyOwnership(user, resource) {`, NewNumber: 1},
			{Kind: "added", Text: `  return true`, NewNumber: 2},
		}}},
	}}}
	snapshots := []model.Snapshot{{ID: "snapshot.verify.ownership", Path: "src/auth/service.ts", Anchor: "verifyOwnership", Category: "authorization", Severity: "critical", Require: []string{"owner", "query"}, Standards: []string{"OWASP-ASVS"}, Message: "Snapshot regressed", Remediation: "Restore it"}}
	findings := snapshotFindings(diff, snapshots)
	if len(findings) != 1 || findings[0].RuleID != "snapshot.verify.ownership" {
		t.Fatalf("unexpected snapshot findings: %#v", findings)
	}
}
