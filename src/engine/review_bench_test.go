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
