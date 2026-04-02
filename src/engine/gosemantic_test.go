package engine

import (
	"testing"

	"github.com/Pimatis/mavetis/src/model"
)

func TestGoSemanticFindingsUseASTFlow(t *testing.T) {
	diff := model.Diff{Files: []model.DiffFile{{
		Path: "service/auth.go",
		Hunks: []model.DiffHunk{{Lines: []model.DiffLine{
			{Kind: "added", Text: `target := ctx.Query("url")`, NewNumber: 1},
			{Kind: "added", Text: `copy := target`, NewNumber: 2},
			{Kind: "added", Text: `http.Get(copy)`, NewNumber: 3},
		}}},
	}}}
	findings := goSemanticFindings(diff)
	if len(findings) == 0 {
		t.Fatal("expected go ast semantic findings")
	}
	if findings[0].RuleID != "semantic.go.ssrf" {
		t.Fatalf("unexpected go semantic finding: %#v", findings[0])
	}
}
