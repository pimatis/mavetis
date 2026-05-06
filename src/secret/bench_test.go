package secret

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/Pimatis/mavetis/src/model"
)

func BenchmarkScanColdNoCache(b *testing.B) {
	root := benchmarkRepo(b)
	config := model.Config{Severity: "low", FailOn: "high"}
	b.ResetTimer()
	for index := 0; index < b.N; index++ {
		if _, err := Scan(root, config, Options{Targets: []string{"."}, NoCache: true}); err != nil {
			b.Fatalf("scan: %v", err)
		}
	}
}

func BenchmarkScanWarmCache(b *testing.B) {
	root := benchmarkRepo(b)
	cachePath := filepath.Join(root, "cache.json")
	config := model.Config{Severity: "low", FailOn: "high"}
	if _, err := Scan(root, config, Options{Targets: []string{"."}, Cache: cachePath}); err != nil {
		b.Fatalf("prime cache: %v", err)
	}
	b.ResetTimer()
	for index := 0; index < b.N; index++ {
		if _, err := Scan(root, config, Options{Targets: []string{"."}, Cache: cachePath}); err != nil {
			b.Fatalf("scan: %v", err)
		}
	}
}

func benchmarkRepo(b *testing.B) string {
	b.Helper()
	root := b.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "src"), 0o700); err != nil {
		b.Fatalf("mkdir src: %v", err)
	}
	for index := 0; index < 300; index++ {
		content := fmt.Sprintf("package src\n\nfunc value%d() string { return \"public-value-%d\" }\n", index, index)
		if index%75 == 0 {
			content = fmt.Sprintf("package src\n\nconst token%d = \"ghp_0123456789abcdefghijklmnopqrstuvwxyzABCD\"\n", index)
		}
		path := filepath.Join(root, "src", fmt.Sprintf("file%d.go", index))
		if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
			b.Fatalf("write file: %v", err)
		}
	}
	return root
}
