package engine

import (
	"testing"

	"github.com/Pimatis/mavetis/src/model"
)

func TestNonceFindingsDetectReuse(t *testing.T) {
	diff := model.Diff{Files: []model.DiffFile{{
		Path: "crypto.go",
		Hunks: []model.DiffHunk{{Lines: []model.DiffLine{
			{Kind: "added", Text: `nonce := make([]byte, 12)`, NewNumber: 1},
			{Kind: "added", Text: `aead.Seal(nil, nonce, first, nil)`, NewNumber: 2},
			{Kind: "added", Text: `aead.Seal(nil, nonce, second, nil)`, NewNumber: 3},
		}}},
	}}}
	findings := nonceFindings(diff)
	if len(findings) != 1 {
		t.Fatalf("expected nonce reuse finding, got %#v", findings)
	}
	if findings[0].RuleID != "crypto.nonce.reuse" {
		t.Fatalf("unexpected nonce rule: %#v", findings[0])
	}
}
