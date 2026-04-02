package engine

import (
	"testing"

	"github.com/Pimatis/mavetis/src/model"
)

func TestManifestFindingsDetectAdvancedDependencyRisks(t *testing.T) {
	diff := model.Diff{Files: []model.DiffFile{{
		Path:  ".npmrc",
		Hunks: []model.DiffHunk{{Lines: []model.DiffLine{{Kind: "added", Text: `registry=https://registry.npmjs.org`, NewNumber: 1}}}},
	}, {
		Path:  "package.json",
		Hunks: []model.DiffHunk{{Lines: []model.DiffLine{{Kind: "added", Text: `"leftpad": "latest"`, NewNumber: 2}, {Kind: "added", Text: `"lodas": "1.0.0"`, NewNumber: 3}}}},
	}, {
		Path:  "go.mod",
		Hunks: []model.DiffHunk{{Lines: []model.DiffLine{{Kind: "added", Text: `replace example.com/lib => github.com/example/lib v1.2.3`, NewNumber: 4}}}},
	}}}
	findings := manifestFindings(diff)
	if len(findings) < 4 {
		t.Fatalf("expected dependency findings, got %#v", findings)
	}
}
