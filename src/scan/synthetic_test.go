package scan

import "testing"

func TestFromFilesSingleFile(t *testing.T) {
	diff := FromFiles([]ScannedFile{{Path: "app.go", Content: "line1\nline2\n"}})
	if diff.Meta.Mode != "file" {
		t.Fatalf("unexpected mode: %#v", diff.Meta)
	}
	if len(diff.Files) != 1 {
		t.Fatalf("unexpected file count: %d", len(diff.Files))
	}
	file := diff.Files[0]
	if file.Path != "app.go" || len(file.Hunks) != 1 {
		t.Fatalf("unexpected file: %#v", file)
	}
	if len(file.Hunks[0].Lines) != 2 {
		t.Fatalf("unexpected line count: %d", len(file.Hunks[0].Lines))
	}
	if file.Hunks[0].Lines[0].Kind != "added" || file.Hunks[0].Lines[1].NewNumber != 2 {
		t.Fatalf("unexpected lines: %#v", file.Hunks[0].Lines)
	}
}

func TestFromFilesEmptyFile(t *testing.T) {
	diff := FromFiles([]ScannedFile{{Path: "empty.txt", Content: ""}})
	if len(diff.Files) != 1 {
		t.Fatalf("unexpected file count: %d", len(diff.Files))
	}
	if len(diff.Files[0].Hunks[0].Lines) != 0 {
		t.Fatalf("expected empty hunk, got %#v", diff.Files[0].Hunks[0].Lines)
	}
}

func TestSplitLinesNormalizesCRLF(t *testing.T) {
	lines := splitLines("a\r\nb\r\n")
	if len(lines) != 2 || lines[0] != "a" || lines[1] != "b" {
		t.Fatalf("unexpected lines: %#v", lines)
	}
}
