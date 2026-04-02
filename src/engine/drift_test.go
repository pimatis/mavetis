package engine

import (
	"testing"

	"github.com/Pimatis/mavetis/src/model"
	"github.com/Pimatis/mavetis/src/rule"
)

func TestReviewDetectsConfigDriftFindings(t *testing.T) {
	diff := model.Diff{
		Files: []model.DiffFile{
			{
				Path: ".env",
				Hunks: []model.DiffHunk{{
					Lines: []model.DiffLine{
						{Kind: "added", Text: `DEBUG=true`, NewNumber: 1},
						{Kind: "added", Text: `NODE_ENV=development`, NewNumber: 2},
					},
				}},
			},
			{
				Path: "deploy/nginx.conf",
				Hunks: []model.DiffHunk{{
					Lines: []model.DiffLine{
						{Kind: "added", Text: `add_header Content-Security-Policy "default-src 'self' 'unsafe-inline'";`, NewNumber: 4},
						{Kind: "added", Text: `ssl_protocols TLSv1 TLSv1.2;`, NewNumber: 5},
						{Kind: "added", Text: `add_header Access-Control-Allow-Origin *;`, NewNumber: 6},
					},
				}},
			},
			{
				Path: "docker-compose.yml",
				Hunks: []model.DiffHunk{{
					Lines: []model.DiffLine{{Kind: "added", Text: `privileged: true`, NewNumber: 3}},
				}},
			},
		},
	}
	report, err := Review(diff, model.Config{Severity: "low"}, rule.Builtins(model.Config{}))
	if err != nil {
		t.Fatalf("review failed: %v", err)
	}
	expected := []string{
		"config.debug.enabled",
		"config.env.production",
		"config.csp.disabled",
		"config.tls.legacy",
		"config.cors.wildcard",
		"config.container.privileged",
	}
	for _, id := range expected {
		if !hasFinding(report.Findings, id) {
			t.Fatalf("expected finding %s in %#v", id, report.Findings)
		}
	}
}
