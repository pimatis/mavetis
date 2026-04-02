package engine

import (
	"testing"

	"github.com/Pimatis/mavetis/src/model"
	"github.com/Pimatis/mavetis/src/rule"
)

func TestReviewDetectsObservabilityLeakFindings(t *testing.T) {
	diff := model.Diff{
		Files: []model.DiffFile{{
			Path: "src/http/logger.ts",
			Hunks: []model.DiffHunk{{
				Lines: []model.DiffLine{
					{Kind: "added", Text: `logger.info("request body", req.body)`, NewNumber: 10},
					{Kind: "added", Text: `logger.info("authorization", req.headers.authorization)`, NewNumber: 11},
					{Kind: "added", Text: `logger.info("email", user.email)`, NewNumber: 12},
					{Kind: "added", Text: `console.error(JSON.stringify(err))`, NewNumber: 13},
					{Kind: "added", Text: `span.SetAttribute("sessionId", sessionId)`, NewNumber: 14},
				},
			}},
		}},
	}
	report, err := Review(diff, model.Config{Severity: "low"}, rule.Builtins(model.Config{}))
	if err != nil {
		t.Fatalf("review failed: %v", err)
	}
	expected := []string{
		"observe.request.body",
		"observe.auth.material",
		"observe.pii",
		"observe.error.stringify",
		"observe.trace.sensitive",
	}
	for _, id := range expected {
		if !hasFinding(report.Findings, id) {
			t.Fatalf("expected finding %s in %#v", id, report.Findings)
		}
	}
}
