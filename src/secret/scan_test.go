package secret

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Pimatis/mavetis/src/model"
)

func TestScanDetectsPatternAndEntropySecrets(t *testing.T) {
	root := t.TempDir()
	content := "package main\nconst apiKey = \"ghp_0123456789abcdefghijklmnopqrstuvwxyzABCD\"\n"
	if err := os.WriteFile(filepath.Join(root, "app.go"), []byte(content), 0o600); err != nil {
		t.Fatalf("write file: %v", err)
	}
	report, err := Scan(root, model.Config{Severity: "low", FailOn: "high", Output: "text"}, Options{Targets: []string{"."}, NoCache: true})
	if err != nil {
		t.Fatalf("scan: %v", err)
	}
	if len(report.Findings) != 1 {
		t.Fatalf("expected one finding, got %#v", report.Findings)
	}
	if report.Findings[0].RuleID != "secret.scan.github.token" {
		t.Fatalf("unexpected rule: %#v", report.Findings[0])
	}
	if strings.Contains(report.Findings[0].Snippet, "abcdefghijklmnopqrstuvwxyz") {
		t.Fatalf("expected masked snippet, got %q", report.Findings[0].Snippet)
	}
}

func TestScanDetectsDotenvSecretOnlyInDotenvPath(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, ".env"), []byte("API_TOKEN=abcdefghijklmnopqrstuvwxyz123456\n"), 0o600); err != nil {
		t.Fatalf("write env: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "example.txt"), []byte("API_TOKEN=abcdefghijklmnopqrstuvwxyz123456\n"), 0o600); err != nil {
		t.Fatalf("write example: %v", err)
	}
	report, err := Scan(root, model.Config{Severity: "low", FailOn: "high"}, Options{Targets: []string{"."}, NoCache: true})
	if err != nil {
		t.Fatalf("scan: %v", err)
	}
	if !hasFinding(report, "secret.scan.dotenv", ".env") {
		t.Fatalf("expected dotenv finding, got %#v", report.Findings)
	}
}

func TestScanRespectsAllowValuesAndPathFilter(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "src"), 0o700); err != nil {
		t.Fatalf("mkdir src: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "src", "app.go"), []byte("token := \"ghp_0123456789abcdefghijklmnopqrstuvwxyzABCD\"\n"), 0o600); err != nil {
		t.Fatalf("write app: %v", err)
	}
	report, err := Scan(root, model.Config{Severity: "low", FailOn: "high", Allow: model.Allow{Values: []string{"ghp_0123456789abcdefghijklmnopqrstuvwxyzABCD"}}}, Options{Targets: []string{"."}, Path: "src/**", NoCache: true})
	if err != nil {
		t.Fatalf("scan: %v", err)
	}
	if len(report.Findings) != 0 {
		t.Fatalf("expected allowlist suppression, got %#v", report.Findings)
	}
}

func TestScanRejectsEscapingTarget(t *testing.T) {
	root := t.TempDir()
	_, err := Scan(root, model.Config{Severity: "low", FailOn: "high"}, Options{Targets: []string{".."}, NoCache: true})
	if err == nil {
		t.Fatal("expected escaping target error")
	}
}

func TestScanUsesCacheForUnchangedUnreadableFile(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "app.go")
	cachePath := filepath.Join(root, "cache.json")
	content := "token := \"ghp_0123456789abcdefghijklmnopqrstuvwxyzABCD\"\n"
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write file: %v", err)
	}
	config := model.Config{Severity: "low", FailOn: "high"}
	first, err := Scan(root, config, Options{Targets: []string{"app.go"}, Cache: cachePath})
	if err != nil {
		t.Fatalf("first scan: %v", err)
	}
	if len(first.Findings) != 1 {
		t.Fatalf("expected first finding, got %#v", first.Findings)
	}
	if err := os.Chmod(path, 0o000); err != nil {
		t.Fatalf("chmod file: %v", err)
	}
	defer func() {
		_ = os.Chmod(path, 0o600)
	}()
	second, err := Scan(root, config, Options{Targets: []string{"app.go"}, Cache: cachePath})
	if err != nil {
		t.Fatalf("second scan should use cache: %v", err)
	}
	if len(second.Findings) != 1 || second.Findings[0].ID != first.Findings[0].ID {
		t.Fatalf("expected cached finding, got %#v", second.Findings)
	}
}

func TestScanInvalidatesCacheWhenFileChanges(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "app.go")
	cachePath := filepath.Join(root, "cache.json")
	if err := os.WriteFile(path, []byte("token := \"ghp_0123456789abcdefghijklmnopqrstuvwxyzABCD\"\n"), 0o600); err != nil {
		t.Fatalf("write file: %v", err)
	}
	config := model.Config{Severity: "low", FailOn: "high"}
	first, err := Scan(root, config, Options{Targets: []string{"app.go"}, Cache: cachePath})
	if err != nil {
		t.Fatalf("first scan: %v", err)
	}
	if len(first.Findings) != 1 {
		t.Fatalf("expected first finding, got %#v", first.Findings)
	}
	if err := os.WriteFile(path, []byte("token := os.Getenv(\"API_TOKEN\")\n"), 0o600); err != nil {
		t.Fatalf("rewrite file: %v", err)
	}
	changed, err := Scan(root, config, Options{Targets: []string{"app.go"}, Cache: cachePath})
	if err != nil {
		t.Fatalf("changed scan: %v", err)
	}
	if len(changed.Findings) != 0 {
		t.Fatalf("expected cache invalidation, got %#v", changed.Findings)
	}
}

func hasFinding(report model.Report, ruleID string, path string) bool {
	for _, finding := range report.Findings {
		if finding.RuleID == ruleID && finding.Path == path {
			return true
		}
	}
	return false
}
