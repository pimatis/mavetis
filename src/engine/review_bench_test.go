package engine

import (
	"strconv"
	"testing"

	"github.com/Pimatis/mavetis/src/model"
	"github.com/Pimatis/mavetis/src/rule"
)

func BenchmarkReview(b *testing.B) {
	lines := make([]model.DiffLine, 0, 400)
	for index := 0; index < 200; index++ {
		lines = append(lines, model.DiffLine{Kind: "added", Text: `token := ctx.Query("token")`, NewNumber: index*2 + 1})
		lines = append(lines, model.DiffLine{Kind: "added", Text: `http.Get(target` + strconv.Itoa(index) + `)`, NewNumber: index*2 + 2})
	}
	diff := model.Diff{Files: []model.DiffFile{{Path: "service/auth.go", Hunks: []model.DiffHunk{{Lines: lines}}}}}
	config := model.Config{Severity: "low"}
	ruleset := rule.Builtins(config)
	b.ResetTimer()
	for index := 0; index < b.N; index++ {
		_, err := Review(diff, config, ruleset)
		if err != nil {
			b.Fatalf("review failed: %v", err)
		}
	}
}

func BenchmarkReviewPolicyProfile(b *testing.B) {
	files := make([]model.DiffFile, 0, 200)
	for index := 0; index < 100; index++ {
		files = append(files, model.DiffFile{Path: "src/auth/file" + strconv.Itoa(index) + ".go", Hunks: []model.DiffHunk{{Lines: []model.DiffLine{{Kind: "added", Text: `token.ParseWithClaims(rawToken)`, NewNumber: 1}}}}})
		files = append(files, model.DiffFile{Path: "src/ui/file" + strconv.Itoa(index) + ".tsx", Hunks: []model.DiffHunk{{Lines: []model.DiffLine{{Kind: "added", Text: `element.innerHTML = body`, NewNumber: 1}}}}})
	}
	diff := model.Diff{Files: files}
	config := model.Config{Severity: "low", FailOn: "critical", Profile: "auth", Zones: model.Zones{Critical: []string{"src/auth/**"}, Restricted: []string{"src/ui/**"}}}
	ruleset := FilterRulesForProfile(rule.Builtins(config), config.Profile)
	b.ResetTimer()
	for index := 0; index < b.N; index++ {
		_, err := Review(diff, config, ruleset)
		if err != nil {
			b.Fatalf("review failed: %v", err)
		}
	}
}

func BenchmarkReviewFileMode(b *testing.B) {
	lines := make([]model.DiffLine, 0, 500)
	for index := 0; index < 250; index++ {
		lines = append(lines, model.DiffLine{Kind: "added", Text: `target := r.URL.Query().Get("url")`, NewNumber: index*2 + 1})
		lines = append(lines, model.DiffLine{Kind: "added", Text: `http.Get(target)`, NewNumber: index*2 + 2})
	}
	diff := model.Diff{Meta: model.DiffMeta{Mode: "file"}, Files: []model.DiffFile{{Path: "service/review.go", Hunks: []model.DiffHunk{{Lines: lines}}}}}
	config := model.Config{Severity: "low"}
	ruleset := rule.Builtins(config)
	b.ResetTimer()
	for index := 0; index < b.N; index++ {
		_, err := Review(diff, config, ruleset)
		if err != nil {
			b.Fatalf("review failed: %v", err)
		}
	}
}
