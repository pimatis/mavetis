package engine

import (
	"testing"

	"github.com/Pimatis/mavetis/src/model"
	"github.com/Pimatis/mavetis/src/rule"
)

type ruleCase struct {
	name     string
	path     string
	line     model.DiffLine
	hunkText string
	expect   string
}

func TestBuiltinsCoverExpandedSecurityRules(t *testing.T) {
	cases := []ruleCase{
		{
			name:     "session fixation heuristic",
			path:     "auth/session.go",
			line:     model.DiffLine{Kind: "added", Text: `sessionID := ctx.Query("session")`, NewNumber: 10},
			hunkText: "login handler\nsessionID := ctx.Query(\"session\")\ncreate session",
			expect:   "session.fixation.input",
		},
		{
			name:     "authorization scope deletion",
			path:     "repository/user.go",
			line:     model.DiffLine{Kind: "deleted", Text: `query = query.Where("tenant_id = ?", tenantID)`, OldNumber: 42},
			hunkText: "find user\nquery = query.Where(\"tenant_id = ?\", tenantID)\nupdate",
			expect:   "authorization.scope.deleted",
		},
		{
			name:     "oauth state disabled",
			path:     "auth/oauth.ts",
			line:     model.DiffLine{Kind: "added", Text: `validateState = false`, NewNumber: 8},
			hunkText: "oauth callback\nvalidateState = false\nauthorize",
			expect:   "oauth.state.disabled",
		},
		{
			name:     "crypto verification removed",
			path:     "auth/token.go",
			line:     model.DiffLine{Kind: "deleted", Text: `claims, err := jwt.ParseWithClaims(token, claims, keyFunc)`, OldNumber: 33},
			hunkText: "token verify\nclaims, err := jwt.ParseWithClaims(token, claims, keyFunc)\nreturn claims",
			expect:   "crypto.verify.deleted",
		},
		{
			name:     "remote dependency source",
			path:     "package.json",
			line:     model.DiffLine{Kind: "added", Text: `"lib": "git+https://github.com/example/lib.git#main"`, NewNumber: 5},
			hunkText: `"dependencies": { "lib": "git+https://github.com/example/lib.git#main" }`,
			expect:   "supply.remote.dependency",
		},
	}
	for _, item := range cases {
		t.Run(item.name, func(t *testing.T) {
			report := runBuiltin(t, item.path, item.line, item.hunkText)
			if !hasRule(report, item.expect) {
				t.Fatalf("expected rule %s, got %#v", item.expect, report.Findings)
			}
		})
	}
}

func TestBuiltinsSuppressWhenMitigationExists(t *testing.T) {
	report := runBuiltin(t, "auth/oauth.ts", model.DiffLine{Kind: "added", Text: `returnTo = request.query.redirect`, NewNumber: 12}, "oauth redirect\nreturnTo = request.query.redirect\nvalidateRedirect(returnTo)")
	if hasRule(report, "auth.redirect.untrusted") {
		t.Fatalf("expected mitigation-aware suppression, got %#v", report.Findings)
	}
}

func runBuiltin(t *testing.T, path string, line model.DiffLine, hunkText string) model.Report {
	t.Helper()
	config := model.Config{Severity: "low"}
	hunk := model.DiffHunk{Lines: []model.DiffLine{line}}
	if hunkText != "" {
		hunk.Lines = []model.DiffLine{{Kind: line.Kind, Text: hunkText, OldNumber: line.OldNumber, NewNumber: line.NewNumber}, line}
	}
	diff := model.Diff{Files: []model.DiffFile{{Path: path, Hunks: []model.DiffHunk{hunk}}}}
	report, err := Review(diff, config, rule.Builtins(config))
	if err != nil {
		t.Fatalf("review failed: %v", err)
	}
	return report
}

func hasRule(report model.Report, ruleID string) bool {
	for _, finding := range report.Findings {
		if finding.RuleID == ruleID {
			return true
		}
	}
	return false
}
