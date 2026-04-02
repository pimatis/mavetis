package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestDefaultBaseFallsBackToExistingBranch(t *testing.T) {
	root := t.TempDir()
	mustGit(t, root, "init")
	mustGit(t, root, "config", "user.email", "test@example.com")
	mustGit(t, root, "config", "user.name", "Test User")
	mustWrite(t, filepath.Join(root, "README.md"), "demo")
	mustGit(t, root, "add", ".")
	mustGit(t, root, "commit", "-m", "init")
	mustGit(t, root, "branch", "-M", "trunk")
	if DefaultBase(root) != "trunk" {
		t.Fatalf("expected trunk base, got %s", DefaultBase(root))
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

func mustWrite(t *testing.T, path string, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write file: %v", err)
	}
}
