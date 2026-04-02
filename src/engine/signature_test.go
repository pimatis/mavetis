package engine

import (
	"testing"

	"github.com/Pimatis/mavetis/src/model"
)

func TestSignatureFindingsDetectAdvancedVerificationRisks(t *testing.T) {
	diff := model.Diff{Files: []model.DiffFile{{
		Path: "token.go",
		Hunks: []model.DiffHunk{{Lines: []model.DiffLine{{
			Kind:      "added",
			Text:      `alg := token.Header["alg"]; kid := token.Header["kid"]; jwk := token.Header["jku"]; http.Get(jwk)`,
			NewNumber: 7,
		}}}},
	}}}
	findings := signatureFindings(diff)
	if len(findings) < 3 {
		t.Fatalf("expected multiple signature findings, got %#v", findings)
	}
}
