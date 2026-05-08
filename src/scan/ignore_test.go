package scan

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadGitignorePatternsParsesPatterns(t *testing.T) {
	root := t.TempDir()
	content := `# comment
node_modules/
*.bak
.DS_Store

/dist/
`
	if err := os.WriteFile(filepath.Join(root, ".gitignore"), []byte(content), 0o600); err != nil {
		t.Fatalf("write .gitignore: %v", err)
	}
	patterns := LoadGitignorePatterns(root)
	if len(patterns) != 4 {
		t.Fatalf("unexpected pattern count: %d", len(patterns))
	}
}

func TestLoadGitignorePatternsNoFile(t *testing.T) {
	root := t.TempDir()
	patterns := LoadGitignorePatterns(root)
	if len(patterns) != 0 {
		t.Fatalf("expected no patterns: %v", patterns)
	}
}

func TestIsGitignoredWildcardPattern(t *testing.T) {
	patterns := []string{"*.bak"}
	if !IsGitignored(patterns, "src/test.bak") {
		t.Fatal("expected test.bak to be ignored")
	}
	if IsGitignored(patterns, "src/test.go") {
		t.Fatal("expected test.go not to be ignored")
	}
}

func TestIsGitignoredDirPattern(t *testing.T) {
	patterns := []string{"node_modules/"}
	if !IsGitignored(patterns, "node_modules") {
		t.Fatal("expected node_modules to be ignored")
	}
	if !IsGitignored(patterns, "node_modules/package.json") {
		t.Fatal("expected node_modules/package.json to be ignored")
	}
	if IsGitignored(patterns, "src/node_modules_helper.go") {
		t.Fatal("expected src/node_modules_helper.go not to be ignored")
	}
}

func TestIsGitignoredRootDirPattern(t *testing.T) {
	patterns := []string{"/dist/"}
	if !IsGitignored(patterns, "dist") {
		t.Fatal("expected dist to be ignored")
	}
	if !IsGitignored(patterns, "dist/bundle.js") {
		t.Fatal("expected dist/bundle.js to be ignored")
	}
	if IsGitignored(patterns, "src/dist") {
		t.Fatal("expected src/dist not to be ignored by root pattern")
	}
}

func TestIsGitignoredSimpleName(t *testing.T) {
	patterns := []string{".DS_Store"}
	if !IsGitignored(patterns, ".DS_Store") {
		t.Fatal("expected .DS_Store to be ignored")
	}
	if !IsGitignored(patterns, "subdir/.DS_Store") {
		t.Fatal("expected subdir/.DS_Store to be ignored")
	}
}

func TestIsGitignoredEmptyPatterns(t *testing.T) {
	if IsGitignored(nil, "any/file.go") {
		t.Fatal("expected no ignore with nil patterns")
	}
}

func TestLoadAllFilesIgnoresGitignoredFiles(t *testing.T) {
	root := t.TempDir()
	gitignore := `*.bak
ignored/
`
	if err := os.WriteFile(filepath.Join(root, ".gitignore"), []byte(gitignore), 0o600); err != nil {
		t.Fatalf("write .gitignore: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(root, "ignored"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "app.go"), []byte("package main\n"), 0o600); err != nil {
		t.Fatalf("write app.go: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "test.bak"), []byte("backup"), 0o600); err != nil {
		t.Fatalf("write test.bak: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "ignored", "secret.go"), []byte("package secret\n"), 0o600); err != nil {
		t.Fatalf("write secret.go: %v", err)
	}
	files, err := LoadAllFiles(root)
	if err != nil {
		t.Fatalf("load all files: %v", err)
	}
	for _, f := range files {
		if f.Path == "test.bak" || f.Path == "ignored/secret.go" {
			t.Fatalf("gitignored file should not be loaded: %s", f.Path)
		}
	}
	foundApp := false
	for _, f := range files {
		if f.Path == "app.go" {
			foundApp = true
			break
		}
	}
	if !foundApp {
		t.Fatal("app.go should be loaded")
	}
}

func TestLoadFilesDirectoryRespectsGitignore(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, ".gitignore"), []byte("*.bak\n"), 0o600); err != nil {
		t.Fatalf("write .gitignore: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(root, "pkg"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "pkg", "a.go"), []byte("package pkg\n"), 0o600); err != nil {
		t.Fatalf("write a.go: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "pkg", "b.bak"), []byte("backup"), 0o600); err != nil {
		t.Fatalf("write b.bak: %v", err)
	}
	files, err := LoadFiles(root, []string{"pkg"})
	if err != nil {
		t.Fatalf("load files: %v", err)
	}
	if len(files) != 1 {
		t.Fatalf("unexpected file count: %d", len(files))
	}
	if files[0].Path != "pkg/a.go" {
		t.Fatalf("unexpected file: %s", files[0].Path)
	}
}
