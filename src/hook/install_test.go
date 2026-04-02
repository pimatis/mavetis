package hook

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestInstallAndUninstall(t *testing.T) {
	root := t.TempDir()
	mustGit(t, root, "init")
	mustGit(t, root, "config", "user.email", "test@example.com")
	mustGit(t, root, "config", "user.name", "Test User")
	if err := os.WriteFile(filepath.Join(root, "README.md"), []byte("demo"), 0o600); err != nil {
		t.Fatalf("write readme: %v", err)
	}
	mustGit(t, root, "add", ".")
	mustGit(t, root, "commit", "-m", "init")
	mustGit(t, root, "branch", "-M", "trunk")
	if err := os.WriteFile(filepath.Join(root, ".git", "hooks", "pre-push"), []byte("legacy"), 0o755); err != nil {
		t.Fatalf("write legacy hook: %v", err)
	}
	if err := Install(root); err != nil {
		t.Fatalf("install: %v", err)
	}
	body, err := os.ReadFile(filepath.Join(root, ".git", "hooks", "pre-push"))
	if err != nil {
		t.Fatalf("read pre-push: %v", err)
	}
	if !strings.Contains(string(body), "--base trunk") {
		t.Fatalf("expected detected base branch, got %s", string(body))
	}
	if _, err := os.Stat(filepath.Join(root, ".git", "hooks", "pre-push.bak")); err != nil {
		t.Fatalf("expected backup hook: %v", err)
	}
	if err := Uninstall(root); err != nil {
		t.Fatalf("uninstall: %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, ".git", "hooks", "pre-commit")); !os.IsNotExist(err) {
		t.Fatalf("expected removed hook, got %v", err)
	}
}

func mustGit(t *testing.T, root string, args ...string) {
	t.Helper()
	command := exec.Command("git", args...)
	command.Dir = root
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v failed: %v %s", args, err, string(output))
	}
}
