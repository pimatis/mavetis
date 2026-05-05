package scan

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadExistingFilesSkipsMissingAndBinaryTargets(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "app.go"), []byte("package app\n"), 0o600); err != nil {
		t.Fatalf("write app: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "binary.go"), []byte{'a', 0, 'b'}, 0o600); err != nil {
		t.Fatalf("write binary: %v", err)
	}
	files, err := LoadExistingFiles(root, []string{"missing.go", "binary.go", "app.go"})
	if err != nil {
		t.Fatalf("load files: %v", err)
	}
	if len(files) != 1 || files[0].Path != "app.go" {
		t.Fatalf("unexpected files: %#v", files)
	}
}

func TestLoadExistingFilesRejectsRootEscape(t *testing.T) {
	root := t.TempDir()
	_, err := LoadExistingFiles(root, []string{"../secret.go"})
	if err == nil {
		t.Fatal("expected root escape rejection")
	}
}
