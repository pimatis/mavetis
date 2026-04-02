package engine

import (
	"testing"

	"github.com/Pimatis/mavetis/src/model"
	"github.com/Pimatis/mavetis/src/rule"
)

func TestReviewDetectsBuiltInBoundaryViolations(t *testing.T) {
	diff := model.Diff{Files: []model.DiffFile{
		{Path: "src/routes/admin.ts", Hunks: []model.DiffHunk{{Lines: []model.DiffLine{{Kind: "added", Text: `import admin from "../internal/admin/service"`, NewNumber: 1}}}}},
		{Path: "src/ui/page.tsx", Hunks: []model.DiffHunk{{Lines: []model.DiffLine{{Kind: "added", Text: `import auth from "../auth/helper"`, NewNumber: 1}}}}},
	}}
	report, err := Review(diff, model.Config{Severity: "low"}, rule.Builtins(model.Config{}))
	if err != nil {
		t.Fatalf("review failed: %v", err)
	}
	if !hasFinding(report.Findings, "boundary.admin.public") {
		t.Fatalf("expected admin boundary finding: %#v", report.Findings)
	}
	if !hasFinding(report.Findings, "boundary.ui.auth") {
		t.Fatalf("expected ui boundary finding: %#v", report.Findings)
	}
}

func TestReviewEvaluatesTypedCustomRules(t *testing.T) {
	rules := []model.Rule{
		{ID: "typed.import", Type: "forbiddenImport", Title: "Forbidden import", Message: "Forbidden import", Remediation: "Fix", Category: "boundary", Severity: "high", Confidence: "high", Target: "added", Paths: []string{"src/ui/**"}, Imports: []string{`(?i)auth`}},
		{ID: "typed.guard", Type: "deletedLineGuard", Title: "Deleted guard", Message: "Deleted guard", Remediation: "Fix", Category: "auth", Severity: "critical", Confidence: "high", Target: "deleted", Require: []string{`requireAuth`}},
		{ID: "typed.env", Type: "forbiddenEnv", Title: "Forbidden env", Message: "Forbidden env", Remediation: "Fix", Category: "config", Severity: "high", Confidence: "high", Target: "added", Keys: []string{`DEBUG`}, ForbiddenValues: []string{"true"}},
		{ID: "typed.middleware", Type: "requiredMiddleware", Title: "Required middleware", Message: "Required middleware", Remediation: "Fix", Category: "boundary", Severity: "high", Confidence: "high", Target: "added", Require: []string{`router\.get`}, Middleware: []string{`requireAuth`}},
		{ID: "typed.call", Type: "requiredCall", Title: "Required call", Message: "Required call", Remediation: "Fix", Category: "boundary", Severity: "high", Confidence: "high", Target: "added", Require: []string{`saveUser`}, Calls: []string{`auditLog`}},
		{ID: "typed.config", Type: "configKeyConstraint", Title: "Config constraint", Message: "Config constraint", Remediation: "Fix", Category: "config", Severity: "high", Confidence: "high", Target: "added", ConstraintKey: "mode", AllowedValues: []string{"production"}},
		{ID: "typed.path", Type: "pathBoundary", Title: "Path boundary", Message: "Path boundary", Remediation: "Fix", Category: "boundary", Severity: "high", Confidence: "high", Target: "added", Paths: []string{"src/routes/**"}, Imports: []string{`(?i)internal/admin`}},
	}
	diff := model.Diff{Files: []model.DiffFile{
		{Path: "src/ui/page.tsx", Hunks: []model.DiffHunk{{Lines: []model.DiffLine{{Kind: "added", Text: `import auth from "../auth/helper"`, NewNumber: 1}}}}},
		{Path: "src/app/auth.ts", Hunks: []model.DiffHunk{{Lines: []model.DiffLine{{Kind: "deleted", Text: `requireAuth(ctx)`, OldNumber: 2}}}}},
		{Path: ".env", Hunks: []model.DiffHunk{{Lines: []model.DiffLine{{Kind: "added", Text: `DEBUG=true`, NewNumber: 3}}}}},
		{Path: "src/routes/user.ts", Hunks: []model.DiffHunk{{Lines: []model.DiffLine{{Kind: "added", Text: `router.get("/users", handler)`, NewNumber: 4}, {Kind: "added", Text: `saveUser(input)`, NewNumber: 5}, {Kind: "added", Text: `import admin from "../internal/admin/service"`, NewNumber: 6}}}}},
		{Path: "config/app.yaml", Hunks: []model.DiffHunk{{Lines: []model.DiffLine{{Kind: "added", Text: `mode: development`, NewNumber: 7}}}}},
	}}
	report, err := Review(diff, model.Config{Severity: "low"}, rules)
	if err != nil {
		t.Fatalf("review failed: %v", err)
	}
	ids := []string{"typed.import", "typed.guard", "typed.env", "typed.middleware", "typed.call", "typed.config", "typed.path"}
	for _, id := range ids {
		if !hasFinding(report.Findings, id) {
			t.Fatalf("expected typed finding %s in %#v", id, report.Findings)
		}
	}
}
