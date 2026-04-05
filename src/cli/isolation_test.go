package cli

import (
	"go/parser"
	"go/token"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestFileReviewPathHasNoNetworkImports(t *testing.T) {
	_, current, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime caller failed")
	}
	paths := []string{
		filepath.Join(filepath.Dir(current), "..", "scan", "load.go"),
		filepath.Join(filepath.Dir(current), "..", "scan", "synthetic.go"),
		filepath.Join(filepath.Dir(current), "..", "resolve", "imports.go"),
		filepath.Join(filepath.Dir(current), "..", "resolve", "resolve.go"),
		filepath.Join(filepath.Dir(current), "..", "resolve", "suggest.go"),
	}
	for _, path := range paths {
		fset := token.NewFileSet()
		file, err := parser.ParseFile(fset, path, nil, parser.ImportsOnly)
		if err != nil {
			t.Fatalf("parse %s: %v", path, err)
		}
		for _, item := range file.Imports {
			value := strings.Trim(item.Path.Value, `"`)
			if value == "net" || strings.HasPrefix(value, "net/") {
				t.Fatalf("unexpected network import %q in %s", value, path)
			}
			if value == "github.com/Pimatis/mavetis/src/update" {
				t.Fatalf("unexpected update import %q in %s", value, path)
			}
		}
	}
}
