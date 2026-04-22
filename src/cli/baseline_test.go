package cli

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func initGitRepo(t *testing.T, dir string) {
	cmd := exec.Command("git", "init")
	cmd.Dir = dir
	if err := cmd.Run(); err != nil {
		t.Fatalf("git init failed: %v", err)
	}
	cmd = exec.Command("git", "config", "user.email", "test@test.com")
	cmd.Dir = dir
	cmd.Run()
	cmd = exec.Command("git", "config", "user.name", "Test")
	cmd.Dir = dir
	cmd.Run()
}

func commitFile(t *testing.T, dir string, name string, content string) {
	path := filepath.Join(dir, name)
	os.WriteFile(path, []byte(content), 0644)
	cmd := exec.Command("git", "add", name)
	cmd.Dir = dir
	if err := cmd.Run(); err != nil {
		t.Fatalf("git add failed: %v", err)
	}
	cmd = exec.Command("git", "commit", "-m", "add "+name)
	cmd.Dir = dir
	if err := cmd.Run(); err != nil {
		t.Fatalf("git commit failed: %v", err)
	}
}

func TestBaselineCreateWritesBaseline(t *testing.T) {
	dir := t.TempDir()
	prevWd, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(prevWd)

	initGitRepo(t, dir)
	commitFile(t, dir, "go.mod", "module test\n")
	os.WriteFile(filepath.Join(dir, "main.go"), []byte("package main\n"), 0644)

	if code := runBaseline([]string{"--create", "--base", "HEAD"}); code != 0 {
		t.Fatalf("expected exit code 0, got %d", code)
	}

	content, err := os.ReadFile(filepath.Join(dir, ".mavetis-baseline.yaml"))
	if err != nil {
		t.Fatal("expected .mavetis-baseline.yaml to be created")
	}
	if !strings.Contains(string(content), "baseline:") {
		t.Fatalf("expected baseline header, got:\n%s", string(content))
	}
}

func TestBaselineRequiresCreate(t *testing.T) {
	if code := runBaseline([]string{}); code != 2 {
		t.Fatalf("expected exit code 2, got %d", code)
	}
}

func TestBaselineCustomOutput(t *testing.T) {
	dir := t.TempDir()
	prevWd, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(prevWd)

	initGitRepo(t, dir)
	commitFile(t, dir, "go.mod", "module test\n")
	os.WriteFile(filepath.Join(dir, "main.go"), []byte("package main\n"), 0644)

	custom := filepath.Join(dir, "custom-baseline.yaml")

	if code := runBaseline([]string{"--create", "--output", custom, "--base", "HEAD"}); code != 0 {
		t.Fatalf("expected exit code 0, got %d", code)
	}

	if _, err := os.Stat(custom); err != nil {
		t.Fatal("expected custom baseline file to be created")
	}
}
