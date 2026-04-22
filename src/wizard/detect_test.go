package wizard

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDetectGoProject(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module example\n"), 0644)
	os.Mkdir(filepath.Join(dir, "auth"), 0755)
	os.WriteFile(filepath.Join(dir, "auth", "login.go"), []byte("package auth\n"), 0644)

	project := Detect(dir)
	if project.Language != "go" {
		t.Fatalf("expected language go, got %s", project.Language)
	}
	if project.Profile != "backend" {
		t.Fatalf("expected backend profile, got %s", project.Profile)
	}
	foundAuth := false
	for _, c := range project.Critical {
		if c == "auth/**" {
			foundAuth = true
		}
	}
	if !foundAuth {
		t.Fatalf("expected auth/** in critical zones, got %v", project.Critical)
	}
}

func TestDetectNodeFrontend(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "package.json"), []byte(`{"dependencies":{"react":"18"}}`), 0644)

	project := Detect(dir)
	if project.Language != "javascript" {
		t.Fatalf("expected javascript, got %s", project.Language)
	}
	if project.Profile != "frontend" {
		t.Fatalf("expected frontend profile, got %s", project.Profile)
	}
}

func TestDetectNodeBackend(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "package.json"), []byte(`{"dependencies":{"express":"4"}}`), 0644)

	project := Detect(dir)
	if project.Profile != "backend" {
		t.Fatalf("expected backend profile, got %s", project.Profile)
	}
}

func TestDetectDefaultProfile(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "README.md"), []byte("# Hello"), 0644)

	project := Detect(dir)
	if project.Profile != "auth" {
		t.Fatalf("expected default auth profile, got %s", project.Profile)
	}
}

func TestDetectIgnoresNestedDeepPaths(t *testing.T) {
	dir := t.TempDir()
	deep := filepath.Join(dir, "a", "b", "c", "d")
	os.MkdirAll(deep, 0755)
	os.WriteFile(filepath.Join(deep, "go.mod"), []byte("module deep\n"), 0644)

	project := Detect(dir)
	if project.Language == "go" {
		t.Fatal("expected go.mod beyond max depth to be ignored")
	}
}

func TestDetectDedupesZones(t *testing.T) {
	dir := t.TempDir()
	os.Mkdir(filepath.Join(dir, "auth"), 0755)
	os.Mkdir(filepath.Join(dir, "api"), 0755)
	os.Mkdir(filepath.Join(dir, "config"), 0755)

	project := Detect(dir)
	seenCritical := map[string]struct{}{}
	for _, c := range project.Critical {
		if _, ok := seenCritical[c]; ok {
			t.Fatalf("duplicate critical zone: %s", c)
		}
		seenCritical[c] = struct{}{}
	}
	seenRestricted := map[string]struct{}{}
	for _, r := range project.Restricted {
		if _, ok := seenRestricted[r]; ok {
			t.Fatalf("duplicate restricted zone: %s", r)
		}
		seenRestricted[r] = struct{}{}
	}
}
