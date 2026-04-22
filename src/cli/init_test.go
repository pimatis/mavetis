package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInitCreatesDefaultConfig(t *testing.T) {
	dir := t.TempDir()
	prevWd, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(prevWd)

	os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module test\n"), 0644)

	if code := runInit([]string{"--default"}); code != 0 {
		t.Fatalf("expected exit code 0, got %d", code)
	}

	content, err := os.ReadFile(filepath.Join(dir, ".mavetis.yaml"))
	if err != nil {
		t.Fatal("expected .mavetis.yaml to be created")
	}
	if !strings.Contains(string(content), "profile: backend") {
		t.Fatalf("expected backend profile in config, got:\n%s", string(content))
	}

	gi, err := os.ReadFile(filepath.Join(dir, ".gitignore"))
	if err != nil {
		t.Fatal("expected .gitignore to be created")
	}
	if !strings.Contains(string(gi), ".mavetis.yaml") {
		t.Fatalf("expected .mavetis.yaml in .gitignore, got:\n%s", string(gi))
	}
}

func TestInitFailsWithoutForce(t *testing.T) {
	dir := t.TempDir()
	prevWd, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(prevWd)

	os.WriteFile(filepath.Join(dir, ".mavetis.yaml"), []byte("existing\n"), 0644)

	if code := runInit([]string{"--default"}); code != 2 {
		t.Fatalf("expected exit code 2, got %d", code)
	}
}

func TestInitForceOverwrites(t *testing.T) {
	dir := t.TempDir()
	prevWd, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(prevWd)

	os.WriteFile(filepath.Join(dir, ".mavetis.yaml"), []byte("existing\n"), 0644)
	os.WriteFile(filepath.Join(dir, "package.json"), []byte(`{"dependencies":{"react":"18"}}`), 0644)

	if code := runInit([]string{"--default", "--force"}); code != 0 {
		t.Fatalf("expected exit code 0, got %d", code)
	}

	content, _ := os.ReadFile(filepath.Join(dir, ".mavetis.yaml"))
	if !strings.Contains(string(content), "profile: frontend") {
		t.Fatalf("expected config to be overwritten with frontend profile, got:\n%s", string(content))
	}
}

func TestInitAppendsGitignore(t *testing.T) {
	dir := t.TempDir()
	prevWd, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(prevWd)

	os.WriteFile(filepath.Join(dir, ".gitignore"), []byte("node_modules/\n"), 0644)
	os.WriteFile(filepath.Join(dir, "package.json"), []byte(`{"dependencies":{"express":"4"}}`), 0644)

	if code := runInit([]string{"--default"}); code != 0 {
		t.Fatalf("expected exit code 0, got %d", code)
	}

	gi, _ := os.ReadFile(filepath.Join(dir, ".gitignore"))
	if !strings.Contains(string(gi), "node_modules/") {
		t.Fatal("expected existing node_modules/ to be preserved")
	}
	if !strings.Contains(string(gi), ".mavetis.yaml") {
		t.Fatalf("expected .mavetis.yaml appended to .gitignore, got:\n%s", string(gi))
	}
}

func TestInitSkipsDuplicateGitignoreEntry(t *testing.T) {
	dir := t.TempDir()
	prevWd, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(prevWd)

	os.WriteFile(filepath.Join(dir, ".gitignore"), []byte(".mavetis.yaml\n"), 0644)
	os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module test\n"), 0644)

	if code := runInit([]string{"--default"}); code != 0 {
		t.Fatalf("expected exit code 0, got %d", code)
	}

	gi, _ := os.ReadFile(filepath.Join(dir, ".gitignore"))
	count := strings.Count(string(gi), ".mavetis.yaml")
	if count != 1 {
		t.Fatalf("expected exactly 1 .mavetis.yaml entry, got %d", count)
	}
}
