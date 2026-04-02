package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSnapshotRulesWritesSecuritySnapshots(t *testing.T) {
	dir := t.TempDir()
	previous, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	defer func() {
		_ = os.Chdir(previous)
	}()
	if err := os.MkdirAll(filepath.Join(dir, "src", "auth"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	file := filepath.Join(dir, "src", "auth", "service.go")
	content := "package auth\n\nfunc validateToken(rawToken string) {\n\tverifySignature(rawToken)\n}\n"
	if err := os.WriteFile(file, []byte(content), 0o600); err != nil {
		t.Fatalf("write file: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	output := filepath.Join(dir, "snapshots.yaml")
	code := snapshotRules([]string{"--output", output, "--path", "src/auth/**"})
	if code != 0 {
		t.Fatalf("expected success, got %d", code)
	}
	body, err := os.ReadFile(output)
	if err != nil {
		t.Fatalf("read output: %v", err)
	}
	text := string(body)
	if !strings.Contains(text, "validateToken") || !strings.Contains(text, "snapshots:") {
		t.Fatalf("unexpected snapshot output: %s", text)
	}
}
