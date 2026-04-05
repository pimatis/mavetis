package resolve

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolveGoLocalPackage(t *testing.T) {
	root := t.TempDir()
	mustWriteResolveFile(t, filepath.Join(root, "go.mod"), "module example.com/demo\n")
	mustWriteResolveFile(t, filepath.Join(root, "pkg", "auth", "auth.go"), "package auth\n")
	mustWriteResolveFile(t, filepath.Join(root, "pkg", "auth", "helper.go"), "package auth\n")
	resolved := ResolveLocal(root, "main.go", "example.com/demo/pkg/auth", 16)
	if len(resolved) != 2 {
		t.Fatalf("unexpected resolution: %#v", resolved)
	}
	if resolved[0] != "pkg/auth/auth.go" {
		t.Fatalf("unexpected resolution: %#v", resolved)
	}
}

func TestResolveGoStdlib(t *testing.T) {
	root := t.TempDir()
	mustWriteResolveFile(t, filepath.Join(root, "go.mod"), "module example.com/demo\n")
	resolved := ResolveLocal(root, "main.go", "fmt", 16)
	if len(resolved) != 0 {
		t.Fatalf("expected no resolution, got %#v", resolved)
	}
}

func TestResolveScriptRelativePath(t *testing.T) {
	root := t.TempDir()
	mustWriteResolveFile(t, filepath.Join(root, "src", "auth.ts"), "export const auth = true\n")
	resolved := ResolveLocal(root, "src/app.ts", "./auth", 16)
	if len(resolved) != 1 || resolved[0] != "src/auth.ts" {
		t.Fatalf("unexpected resolution: %#v", resolved)
	}
}

func TestResolveScriptIndexFile(t *testing.T) {
	root := t.TempDir()
	mustWriteResolveFile(t, filepath.Join(root, "src", "bar", "index.js"), "module.exports = {}\n")
	resolved := ResolveLocal(root, "src/app.js", "./bar", 16)
	if len(resolved) != 1 || resolved[0] != "src/bar/index.js" {
		t.Fatalf("unexpected resolution: %#v", resolved)
	}
}

func TestResolvePythonRelative(t *testing.T) {
	root := t.TempDir()
	mustWriteResolveFile(t, filepath.Join(root, "pkg", "utils.py"), "value = 1\n")
	resolved := ResolveLocal(root, "pkg/app.py", ".utils", 16)
	if len(resolved) != 1 || resolved[0] != "pkg/utils.py" {
		t.Fatalf("unexpected resolution: %#v", resolved)
	}
}

func TestResolvePythonAbsolute(t *testing.T) {
	root := t.TempDir()
	mustWriteResolveFile(t, filepath.Join(root, "pkg", "auth", "__init__.py"), "")
	resolved := ResolveLocal(root, "pkg/app.py", "pkg.auth", 16)
	if len(resolved) != 1 || resolved[0] != "pkg/auth/__init__.py" {
		t.Fatalf("unexpected resolution: %#v", resolved)
	}
}

func TestResolveJVM(t *testing.T) {
	root := t.TempDir()
	mustWriteResolveFile(t, filepath.Join(root, "src", "main", "java", "com", "example", "Auth.java"), "class Auth {}\n")
	resolved := ResolveLocal(root, "src/main/java/com/example/App.java", "com.example.Auth", 16)
	if len(resolved) != 1 || resolved[0] != "src/main/java/com/example/Auth.java" {
		t.Fatalf("unexpected resolution: %#v", resolved)
	}
}

func mustWriteResolveFile(t *testing.T, path string, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write: %v", err)
	}
}
