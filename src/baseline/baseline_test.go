package baseline

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Pimatis/mavetis/src/model"
)

func TestLoadMissingReturnsEmpty(t *testing.T) {
	file, err := Load("/nonexistent/path/.mavetis-baseline.yaml")
	if err != nil {
		t.Fatalf("expected no error for missing file, got %v", err)
	}
	if len(file.Entries) != 0 {
		t.Fatalf("expected empty entries, got %d", len(file.Entries))
	}
}

func TestLoadParsesBaseline(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".mavetis-baseline.yaml")
	content := `baseline:
  - rule: inject.sql.raw
    path: src/api/handler.go
    line: 45
  - rule: secret.generic
    path: config/.env
    line: 3
`
	os.WriteFile(path, []byte(content), 0644)

	file, err := Load(path)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(file.Entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(file.Entries))
	}
	if file.Entries[0].RuleID != "inject.sql.raw" {
		t.Fatalf("expected inject.sql.raw, got %s", file.Entries[0].RuleID)
	}
	if file.Entries[0].Line != 45 {
		t.Fatalf("expected line 45, got %d", file.Entries[0].Line)
	}
}

func TestCreateWritesBaseline(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".mavetis-baseline.yaml")
	report := model.Report{
		Findings: []model.Finding{
			{RuleID: "a", Path: "x.go", Line: 1},
			{RuleID: "b", Path: "y.go", Line: 2},
		},
	}
	if err := Create(path, report); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal("expected file to be created")
	}
	if !strings.Contains(string(data), "a") {
		t.Fatal("expected rule a in baseline")
	}
	if !strings.Contains(string(data), "y.go") {
		t.Fatal("expected path y.go in baseline")
	}
}

func TestCreateDedupesFindings(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".mavetis-baseline.yaml")
	report := model.Report{
		Findings: []model.Finding{
			{RuleID: "a", Path: "x.go", Line: 1},
			{RuleID: "a", Path: "x.go", Line: 1},
		},
	}
	if err := Create(path, report); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	file, _ := Load(path)
	if len(file.Entries) != 1 {
		t.Fatalf("expected 1 entry after dedupe, got %d", len(file.Entries))
	}
}

func TestCreateSortsEntries(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".mavetis-baseline.yaml")
	report := model.Report{
		Findings: []model.Finding{
			{RuleID: "b", Path: "z.go", Line: 2},
			{RuleID: "a", Path: "a.go", Line: 1},
		},
	}
	if err := Create(path, report); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	file, _ := Load(path)
	if file.Entries[0].Path != "a.go" {
		t.Fatalf("expected sorted by path, got %s", file.Entries[0].Path)
	}
}

func TestFilterRemovesKnown(t *testing.T) {
	report := model.Report{
		Findings: []model.Finding{
			{RuleID: "a", Path: "x.go", Line: 1},
			{RuleID: "b", Path: "y.go", Line: 2},
			{RuleID: "c", Path: "z.go", Line: 3},
		},
	}
	baseline := File{
		Entries: []Entry{
			{RuleID: "a", Path: "x.go", Line: 1},
			{RuleID: "c", Path: "z.go", Line: 3},
		},
	}
	filtered := Filter(report, baseline)
	if len(filtered.Findings) != 1 {
		t.Fatalf("expected 1 remaining finding, got %d", len(filtered.Findings))
	}
	if filtered.Findings[0].RuleID != "b" {
		t.Fatalf("expected rule b remaining, got %s", filtered.Findings[0].RuleID)
	}
}

func TestFilterPreservesAllWhenEmptyBaseline(t *testing.T) {
	report := model.Report{
		Findings: []model.Finding{
			{RuleID: "a", Path: "x.go", Line: 1},
		},
	}
	filtered := Filter(report, File{})
	if len(filtered.Findings) != 1 {
		t.Fatalf("expected 1 finding, got %d", len(filtered.Findings))
	}
}
