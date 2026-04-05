package scan

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestLoadFilesValidFile(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "app.go")
	mustWriteFile(t, path, "package main\n")
	files, err := LoadFiles(root, []string{"app.go"})
	if err != nil {
		t.Fatalf("load files: %v", err)
	}
	if len(files) != 1 {
		t.Fatalf("unexpected file count: %d", len(files))
	}
	if files[0].Path != "app.go" {
		t.Fatalf("unexpected path: %s", files[0].Path)
	}
}

func TestLoadFilesMultipleFilesAndGlob(t *testing.T) {
	root := t.TempDir()
	mustWriteFile(t, filepath.Join(root, "src", "a.go"), "package src\n")
	mustWriteFile(t, filepath.Join(root, "src", "b.go"), "package src\n")
	files, err := LoadFiles(root, []string{"src/*.go"})
	if err != nil {
		t.Fatalf("load files: %v", err)
	}
	if len(files) != 2 {
		t.Fatalf("unexpected file count: %d", len(files))
	}
	if files[0].Path != "src/a.go" || files[1].Path != "src/b.go" {
		t.Fatalf("unexpected files: %#v", files)
	}
}

func TestLoadFilesRejectsPathTraversal(t *testing.T) {
	root := t.TempDir()
	outside := filepath.Join(filepath.Dir(root), "outside.txt")
	mustWriteFile(t, outside, "secret")
	_, err := LoadFiles(root, []string{"../outside.txt"})
	if err == nil {
		t.Fatal("expected traversal rejection")
	}
}

func TestLoadFilesExpandsDirectory(t *testing.T) {
	root := t.TempDir()
	mustWriteFile(t, filepath.Join(root, "pkg", "a.go"), "package pkg\n")
	mustWriteFile(t, filepath.Join(root, "pkg", "b.go"), "package pkg\n")
	files, err := LoadFiles(root, []string{"pkg"})
	if err != nil {
		t.Fatalf("load files: %v", err)
	}
	if len(files) != 2 {
		t.Fatalf("unexpected file count: %d", len(files))
	}
	if files[0].Path != "pkg/a.go" || files[1].Path != "pkg/b.go" {
		t.Fatalf("unexpected files: %#v", files)
	}
}

func TestLoadFilesRejectsEmptyDirectory(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "empty"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	_, err := LoadFiles(root, []string{"empty"})
	if err == nil {
		t.Fatal("expected empty directory rejection")
	}
}

func TestLoadFilesRejectsBinaryFile(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "data.bin")
	if err := os.WriteFile(path, []byte{'a', 0, 'b'}, 0o600); err != nil {
		t.Fatalf("write file: %v", err)
	}
	_, err := LoadFiles(root, []string{"data.bin"})
	if err == nil {
		t.Fatal("expected binary rejection")
	}
}

func TestLoadFilesRejectsOversizedFile(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "large.txt")
	mustWriteFile(t, path, strings.Repeat("a", maxExplicitFileSize+1))
	_, err := LoadFiles(root, []string{"large.txt"})
	if err == nil {
		t.Fatal("expected size rejection")
	}
}

func TestLoadFilesRejectsNULPath(t *testing.T) {
	root := t.TempDir()
	_, err := LoadFiles(root, []string{"bad\x00.txt"})
	if err == nil {
		t.Fatal("expected NUL rejection")
	}
}

func TestLoadFilesResolvesRelativePath(t *testing.T) {
	root := t.TempDir()
	mustWriteFile(t, filepath.Join(root, "nested", "file.go"), "package nested\n")
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	defer func() {
		_ = os.Chdir(cwd)
	}()
	if err := os.Chdir(root); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	files, err := LoadFiles(".", []string{"nested/file.go"})
	if err != nil {
		t.Fatalf("load files: %v", err)
	}
	if len(files) != 1 || files[0].Path != "nested/file.go" {
		t.Fatalf("unexpected files: %#v", files)
	}
}

func TestLoadFilesRejectsSymlinkEscape(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("symlink permissions vary on Windows")
	}
	root := t.TempDir()
	outsideDir := t.TempDir()
	outside := filepath.Join(outsideDir, "outside.txt")
	mustWriteFile(t, outside, "secret")
	link := filepath.Join(root, "escape.txt")
	if err := os.Symlink(outside, link); err != nil {
		t.Skipf("symlink unsupported: %v", err)
	}
	_, err := LoadFiles(root, []string{"escape.txt"})
	if err == nil {
		t.Fatal("expected symlink escape rejection")
	}
}

func TestLoadFilesAcceptsSymlinkInsideRoot(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("symlink permissions vary on Windows")
	}
	root := t.TempDir()
	target := filepath.Join(root, "target.go")
	mustWriteFile(t, target, "package main\n")
	link := filepath.Join(root, "link.go")
	if err := os.Symlink(target, link); err != nil {
		t.Skipf("symlink unsupported: %v", err)
	}
	files, err := LoadFiles(root, []string{"link.go"})
	if err != nil {
		t.Fatalf("load files: %v", err)
	}
	if len(files) != 1 {
		t.Fatalf("unexpected file count: %d", len(files))
	}
	if files[0].Path != "target.go" {
		t.Fatalf("unexpected resolved path: %s", files[0].Path)
	}
}

func mustWriteFile(t *testing.T, path string, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write file: %v", err)
	}
}
