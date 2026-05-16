package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/Pimatis/mavetis/src/model"
	"github.com/Pimatis/mavetis/src/rule"
)

func BenchmarkBuildFileReportNoCache(b *testing.B) {
	root := benchmarkReviewRepo(b)
	changeWorkingDir(b, root)
	cfg := model.Config{Severity: "low", FailOn: "high", Output: "text"}
	spec := model.Review{Files: []string{"src"}, NoCache: true}
	rules := rule.Builtins(cfg)
	b.ResetTimer()
	for index := 0; index < b.N; index++ {
		if _, err := buildFileReport(spec, cfg, rules); err != nil {
			b.Fatalf("build report: %v", err)
		}
	}
}

func BenchmarkBuildFileReportWarmCache(b *testing.B) {
	root := benchmarkReviewRepo(b)
	changeWorkingDir(b, root)
	cfg := model.Config{Severity: "low", FailOn: "high", Output: "text"}
	spec := model.Review{Files: []string{"src"}, CachePath: filepath.Join(root, "review-cache.json")}
	rules := rule.Builtins(cfg)
	if _, err := buildFileReport(spec, cfg, rules); err != nil {
		b.Fatalf("prime cache: %v", err)
	}
	b.ResetTimer()
	for index := 0; index < b.N; index++ {
		if _, err := buildFileReport(spec, cfg, rules); err != nil {
			b.Fatalf("build report: %v", err)
		}
	}
}

func benchmarkReviewRepo(b *testing.B) string {
	b.Helper()
	root := b.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "src"), 0o700); err != nil {
		b.Fatalf("mkdir src: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "go.mod"), []byte("module example.com/app\n"), 0o600); err != nil {
		b.Fatalf("write go.mod: %v", err)
	}
	for index := 0; index < 120; index++ {
		content := fmt.Sprintf("package src\n\nfunc value%d() string { return \"public-value-%d\" }\n", index, index)
		if index%40 == 0 {
			content = fmt.Sprintf("package src\n\nconst token%d = \"ghp_0123456789abcdefghijklmnopqrstuvwxyzABCD\"\n", index)
		}
		path := filepath.Join(root, "src", fmt.Sprintf("file%d.go", index))
		if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
			b.Fatalf("write file: %v", err)
		}
	}
	return root
}
