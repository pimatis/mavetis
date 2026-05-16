package cli

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Pimatis/mavetis/src/model"
	"github.com/Pimatis/mavetis/src/scan"
)

func TestChangedContextSeedPathsSkipsDeletedAndDuplicates(t *testing.T) {
	parsed := model.Diff{Files: []model.DiffFile{
		{Path: "src/app.go", Change: "modified"},
		{Path: "src/app.go", Change: "modified"},
		{Path: "src/old.go", Change: "deleted"},
		{Path: "", Change: "added"},
	}}
	paths := changedContextSeedPaths(parsed)
	if len(paths) != 1 || paths[0] != "src/app.go" {
		t.Fatalf("unexpected paths: %#v", paths)
	}
}

func TestAppendContextFilesAddsSyntheticContextDeterministically(t *testing.T) {
	parsed := model.Diff{
		Meta:  model.DiffMeta{Mode: "staged"},
		Files: []model.DiffFile{{Path: "src/app.go", Change: "modified"}},
	}
	result := appendContextFiles(parsed, []scan.ScannedFile{
		{Path: "src/app.go", Content: "package app\n"},
		{Path: "src/auth.go", Content: "package app\n"},
	})
	if result.Meta.Mode != "staged+context" {
		t.Fatalf("unexpected mode: %s", result.Meta.Mode)
	}
	if len(result.Files) != 2 {
		t.Fatalf("unexpected file count: %d", len(result.Files))
	}
	if result.Files[1].Path != "src/auth.go" || result.Files[1].Change != "context" {
		t.Fatalf("unexpected context file: %#v", result.Files[1])
	}
	if result.Files[1].Hunks[0].Header != "@@ changed context @@" {
		t.Fatalf("unexpected context hunk: %#v", result.Files[1].Hunks[0])
	}
}

func TestWithChangedContextDiscoversLocalImports(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "src", "auth"), 0o700); err != nil {
		t.Fatalf("mkdir auth: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "go.mod"), []byte("module example.com/app\n"), 0o600); err != nil {
		t.Fatalf("write go.mod: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "src", "app.go"), []byte("package src\n\nimport _ \"example.com/app/src/auth\"\n"), 0o600); err != nil {
		t.Fatalf("write app: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "src", "auth", "auth.go"), []byte("package auth\n"), 0o600); err != nil {
		t.Fatalf("write auth: %v", err)
	}
	changeWorkingDir(t, root)
	parsed := model.Diff{
		Meta: model.DiffMeta{Mode: "staged"},
		Files: []model.DiffFile{{
			Path:   "src/app.go",
			Change: "modified",
			Hunks:  []model.DiffHunk{{Lines: []model.DiffLine{{Kind: "added", Text: "package src", NewNumber: 1}}}},
		}},
	}
	result, suggestions, err := withChangedContext(parsed)
	if err != nil {
		t.Fatalf("with context: %v", err)
	}
	if len(suggestions) != 1 || suggestions[0].Path != "src/auth/auth.go" || !suggestions[0].Reviewed {
		t.Fatalf("unexpected suggestions: %#v", suggestions)
	}
	if len(result.Files) != 2 || result.Files[1].Path != "src/auth/auth.go" || result.Files[1].Change != "context" {
		t.Fatalf("unexpected context diff: %#v", result.Files)
	}
}
