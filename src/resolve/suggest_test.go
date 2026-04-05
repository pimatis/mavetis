package resolve

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/Pimatis/mavetis/src/scan"
)

func TestSuggestSingleDependency(t *testing.T) {
	root := t.TempDir()
	mustWriteSuggestFile(t, filepath.Join(root, "src", "app.ts"), "import auth from './auth'\n")
	mustWriteSuggestFile(t, filepath.Join(root, "src", "auth.ts"), "export const auth = true\n")
	seeds := []scan.ScannedFile{{Path: "src/app.ts", Content: "import auth from './auth'\n"}}
	suggestions, err := Suggest(root, seeds, DefaultLimits())
	if err != nil {
		t.Fatalf("suggest: %v", err)
	}
	if len(suggestions) != 1 {
		t.Fatalf("unexpected suggestions: %#v", suggestions)
	}
	if suggestions[0].Path != "src/auth.ts" || suggestions[0].Depth != 1 {
		t.Fatalf("unexpected suggestion: %#v", suggestions[0])
	}
}

func TestDiscoverReturnsSuggestedFiles(t *testing.T) {
	root := t.TempDir()
	mustWriteSuggestFile(t, filepath.Join(root, "src", "app.ts"), "import auth from './auth'\n")
	mustWriteSuggestFile(t, filepath.Join(root, "src", "auth.ts"), "export const auth = true\n")
	seeds := []scan.ScannedFile{{Path: "src/app.ts", Content: "import auth from './auth'\n"}}
	files, suggestions, err := Discover(root, seeds, DefaultLimits())
	if err != nil {
		t.Fatalf("discover: %v", err)
	}
	if len(files) != 1 || files[0].Path != "src/auth.ts" {
		t.Fatalf("unexpected files: %#v", files)
	}
	if len(suggestions) != 1 || suggestions[0].Path != "src/auth.ts" {
		t.Fatalf("unexpected suggestions: %#v", suggestions)
	}
}

func TestSuggestRespectsDepthLimit(t *testing.T) {
	root := t.TempDir()
	mustWriteSuggestFile(t, filepath.Join(root, "src", "app.ts"), "import auth from './auth'\n")
	mustWriteSuggestFile(t, filepath.Join(root, "src", "auth.ts"), "import policy from './policy'\n")
	mustWriteSuggestFile(t, filepath.Join(root, "src", "policy.ts"), "export const policy = true\n")
	seeds := []scan.ScannedFile{{Path: "src/app.ts", Content: "import auth from './auth'\n"}}
	limits := DefaultLimits()
	limits.MaxDepth = 1
	suggestions, err := Suggest(root, seeds, limits)
	if err != nil {
		t.Fatalf("suggest: %v", err)
	}
	if len(suggestions) != 1 || suggestions[0].Path != "src/auth.ts" {
		t.Fatalf("unexpected suggestions: %#v", suggestions)
	}
}

func TestSuggestHandlesCircularImports(t *testing.T) {
	root := t.TempDir()
	mustWriteSuggestFile(t, filepath.Join(root, "src", "a.ts"), "import b from './b'\n")
	mustWriteSuggestFile(t, filepath.Join(root, "src", "b.ts"), "import a from './a'\n")
	seeds := []scan.ScannedFile{{Path: "src/a.ts", Content: "import b from './b'\n"}}
	suggestions, err := Suggest(root, seeds, DefaultLimits())
	if err != nil {
		t.Fatalf("suggest: %v", err)
	}
	if len(suggestions) != 1 || suggestions[0].Path != "src/b.ts" {
		t.Fatalf("unexpected suggestions: %#v", suggestions)
	}
}

func TestSuggestSkipsIgnoredDirectories(t *testing.T) {
	root := t.TempDir()
	mustWriteSuggestFile(t, filepath.Join(root, "src", "app.ts"), "import lib from '../node_modules/lib/index'\n")
	mustWriteSuggestFile(t, filepath.Join(root, "node_modules", "lib", "index.ts"), "export const lib = true\n")
	seeds := []scan.ScannedFile{{Path: "src/app.ts", Content: "import lib from '../node_modules/lib/index'\n"}}
	suggestions, err := Suggest(root, seeds, DefaultLimits())
	if err != nil {
		t.Fatalf("suggest: %v", err)
	}
	if len(suggestions) != 0 {
		t.Fatalf("expected ignored suggestion to be skipped, got %#v", suggestions)
	}
}

func TestSuggestSkipsBinaryAndByteBudget(t *testing.T) {
	root := t.TempDir()
	mustWriteSuggestFile(t, filepath.Join(root, "src", "app.ts"), "import dep from './dep'\n")
	if err := os.MkdirAll(filepath.Join(root, "src"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "src", "dep.ts"), []byte{'a', 0, 'b'}, 0o600); err != nil {
		t.Fatalf("write: %v", err)
	}
	seeds := []scan.ScannedFile{{Path: "src/app.ts", Content: "import dep from './dep'\n"}}
	limits := DefaultLimits()
	limits.MaxTotalBytes = 10
	suggestions, err := Suggest(root, seeds, limits)
	if err != nil {
		t.Fatalf("suggest: %v", err)
	}
	if len(suggestions) != 0 {
		t.Fatalf("expected no suggestions, got %#v", suggestions)
	}
}

func BenchmarkSuggest(b *testing.B) {
	root := b.TempDir()
	for index := 0; index < 10; index++ {
		name := fmt.Sprintf("file%d.ts", index)
		next := ""
		if index < 9 {
			next = fmt.Sprintf("import next from './file%d'\n", index+1)
		}
		mustWriteSuggestFile(b, filepath.Join(root, "src", name), next+"export const value = true\n")
	}
	seeds := []scan.ScannedFile{{Path: "src/file0.ts", Content: "import next from './file1'\n"}}
	limits := DefaultLimits()
	b.ResetTimer()
	for index := 0; index < b.N; index++ {
		_, err := Suggest(root, seeds, limits)
		if err != nil {
			b.Fatalf("suggest: %v", err)
		}
	}
}

func mustWriteSuggestFile(t testing.TB, path string, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write: %v", err)
	}
}
