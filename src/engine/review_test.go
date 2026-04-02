package engine

import (
	"testing"

	"github.com/Pimatis/mavetis/src/model"
	"github.com/Pimatis/mavetis/src/rule"
)

func TestReviewFindsAddedAndDeletedRisks(t *testing.T) {
	diff := model.Diff{
		Meta: model.DiffMeta{Mode: "staged"},
		Files: []model.DiffFile{{
			Path: "web/app.js",
			Hunks: []model.DiffHunk{{
				Lines: []model.DiffLine{
					{Kind: "added", Text: `localStorage.setItem("token", jwt)`, NewNumber: 10},
					{Kind: "deleted", Text: `requireRole("admin")`, OldNumber: 9},
				},
			}},
		}},
	}
	config := model.Config{Severity: "low"}
	report, err := Review(diff, config, rule.Builtins(config))
	if err != nil {
		t.Fatalf("review failed: %v", err)
	}
	if len(report.Findings) < 2 {
		t.Fatalf("expected findings, got %d", len(report.Findings))
	}
}

func TestReviewRespectsAllowlist(t *testing.T) {
	diff := model.Diff{
		Files: []model.DiffFile{{
			Path: "app.js",
			Hunks: []model.DiffHunk{{
				Lines: []model.DiffLine{{Kind: "added", Text: `const secret = "safe-example-value"`, NewNumber: 1}},
			}},
		}},
	}
	config := model.Config{
		Severity: "low",
		Allow:    model.Allow{Values: []string{"safe-example-value"}},
	}
	rules := []model.Rule{{
		ID:         "secret.test",
		Title:      "Secret",
		Message:    "Secret",
		Category:   "secret",
		Severity:   "high",
		Confidence: "high",
		Target:     "added",
		Require:    []string{`secret`},
	}}
	report, err := Review(diff, config, rules)
	if err != nil {
		t.Fatalf("review failed: %v", err)
	}
	if len(report.Findings) != 0 {
		t.Fatalf("expected allowlist to suppress findings, got %d", len(report.Findings))
	}
}

func TestReviewSupportsAbsentPatterns(t *testing.T) {
	diff := model.Diff{
		Files: []model.DiffFile{{
			Path: "auth.ts",
			Hunks: []model.DiffHunk{{
				Lines: []model.DiffLine{{Kind: "added", Text: `jwt.decode(token)`, NewNumber: 4}},
			}},
		}},
	}
	rules := []model.Rule{{
		ID:         "token.decode",
		Title:      "Decode only",
		Message:    "Decode only",
		Category:   "token",
		Severity:   "high",
		Confidence: "high",
		Target:     "added",
		Require:    []string{`decode`},
		Absent:     []string{`verify`},
	}}
	report, err := Review(diff, model.Config{Severity: "low"}, rules)
	if err != nil {
		t.Fatalf("review failed: %v", err)
	}
	if len(report.Findings) != 1 {
		t.Fatalf("expected absent-aware finding, got %d", len(report.Findings))
	}
}

func TestBuiltinsIgnoreDocumentationForCodeRules(t *testing.T) {
	diff := model.Diff{
		Files: []model.DiffFile{{
			Path: "DMCHAT_GUIDE.md",
			Hunks: []model.DiffHunk{{
				Lines: []model.DiffLine{{Kind: "added", Text: `Open websocket with conversation.id and profile path.`, NewNumber: 12}},
			}},
		}},
	}
	report, err := Review(diff, model.Config{Severity: "low"}, rule.Builtins(model.Config{Severity: "low"}))
	if err != nil {
		t.Fatalf("review failed: %v", err)
	}
	if len(report.Findings) != 0 {
		t.Fatalf("expected documentation diff to stay clean, got %#v", report.Findings)
	}
}

func TestBuiltinsAvoidCommonFalsePositives(t *testing.T) {
	diff := model.Diff{
		Files: []model.DiffFile{{
			Path: "app.go",
			Hunks: []model.DiffHunk{{
				Lines: []model.DiffLine{
					{Kind: "added", Text: `if token == "" {`, NewNumber: 10},
					{Kind: "added", Text: `delete(userLimits, scope+":"+identifier)`, NewNumber: 11},
				},
			}},
		}},
	}
	report, err := Review(diff, model.Config{Severity: "low"}, rule.Builtins(model.Config{Severity: "low"}))
	if err != nil {
		t.Fatalf("review failed: %v", err)
	}
	for _, finding := range report.Findings {
		if finding.RuleID == "crypto.compare.missing" {
			t.Fatalf("unexpected constant-time finding: %#v", finding)
		}
		if finding.RuleID == "inject.sql.raw" {
			t.Fatalf("unexpected sql finding: %#v", finding)
		}
	}
}
