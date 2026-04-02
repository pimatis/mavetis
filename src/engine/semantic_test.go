package engine

import (
	"testing"

	"github.com/Pimatis/mavetis/src/model"
)

func TestSemanticFindingsCatchTaintedFlows(t *testing.T) {
	diff := model.Diff{Files: []model.DiffFile{{
		Path: "service/auth.go",
		Hunks: []model.DiffHunk{{
			Lines: []model.DiffLine{{Kind: "added", Text: `target := ctx.Query("url")`, NewNumber: 1}, {Kind: "added", Text: `http.Get(target)`, NewNumber: 2}},
		}},
	}}}
	findings := semanticFindings(diff)
	if len(findings) == 0 {
		t.Fatal("expected semantic findings")
	}
	if findings[0].RuleID != "semantic.ssrf.flow" {
		t.Fatalf("unexpected rule: %#v", findings[0])
	}
}

func TestCrossFindingsCorrelateBranchSignals(t *testing.T) {
	diff := model.Diff{Files: []model.DiffFile{
		{Path: "auth.go", Hunks: []model.DiffHunk{{Lines: []model.DiffLine{{Kind: "deleted", Text: `requireAuth()`, OldNumber: 1}}}}},
		{Path: "routes.go", Hunks: []model.DiffHunk{{Lines: []model.DiffLine{{Kind: "added", Text: `router.Get("/admin", handler)`, NewNumber: 5}}}}},
	}}
	findings := crossFindings(diff)
	if len(findings) == 0 {
		t.Fatal("expected cross findings")
	}
	if findings[0].RuleID != "branch.guard.regression" {
		t.Fatalf("unexpected rule: %#v", findings[0])
	}
}
