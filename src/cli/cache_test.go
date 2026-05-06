package cli

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Pimatis/mavetis/src/model"
	"github.com/Pimatis/mavetis/src/rule"
)

func TestBuildFileReportWritesAndInvalidatesReviewCache(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "go.mod"), []byte("module example.com/app\n"), 0o600); err != nil {
		t.Fatalf("write go.mod: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "app.go"), []byte("package main\nconst token = \"ghp_0123456789abcdefghijklmnopqrstuvwxyzABCD\"\n"), 0o600); err != nil {
		t.Fatalf("write app: %v", err)
	}
	t.Chdir(root)
	cachePath := filepath.Join(root, "review-cache.json")
	cfg := model.Config{Severity: "low", FailOn: "high", Output: "text"}
	spec := model.Review{Files: []string{"app.go"}, CachePath: cachePath}
	first, err := buildFileReport(spec, cfg, rule.Builtins(cfg))
	if err != nil {
		t.Fatalf("first report: %v", err)
	}
	if len(first.Findings) == 0 {
		t.Fatalf("expected first finding")
	}
	if _, err := os.Stat(cachePath); err != nil {
		t.Fatalf("expected cache file: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "app.go"), []byte("package main\nfunc token() string { return os.Getenv(\"API_TOKEN\") }\n"), 0o600); err != nil {
		t.Fatalf("rewrite app: %v", err)
	}
	second, err := buildFileReport(spec, cfg, rule.Builtins(cfg))
	if err != nil {
		t.Fatalf("second report: %v", err)
	}
	if len(second.Findings) != 0 {
		t.Fatalf("expected invalidated clean report, got %#v", second.Findings)
	}
}
