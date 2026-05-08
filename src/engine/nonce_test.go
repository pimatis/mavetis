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

func TestNonceFindingsIgnoreArchiveVariables(t *testing.T) {
	diff := model.Diff{Files: []model.DiffFile{{
		Path: "src/update/release.go",
		Hunks: []model.DiffHunk{{Lines: []model.DiffLine{
			{Kind: "added", Text: `archivePath := filepath.Join(dir, archiveName)`, NewNumber: 1},
			{Kind: "added", Text: `if err := extractBinary(archivePath, binaryName, binaryPath); err != nil {`, NewNumber: 2},
			{Kind: "added", Text: `archive, err := zip.OpenReader(archivePath)`, NewNumber: 3},
		}}},
	}}}
	findings := nonceFindings(diff)
	if len(findings) != 0 {
		t.Fatalf("expected archive path variables to stay quiet, got %#v", findings)
	}
}
