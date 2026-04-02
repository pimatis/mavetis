package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".mavetis.yaml")
	content := `severity: medium
failon: critical
ignore:
  - vendor/**
allow:
  values:
    - safe-example
company:
  prefixes:
    - corp_
`
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}
	config, err := Load(path)
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if config.Severity != "medium" || config.FailOn != "critical" {
		t.Fatalf("unexpected config: %#v", config)
	}
	if len(config.Company.Prefixes) != 1 || config.Company.Prefixes[0] != "corp_" {
		t.Fatalf("unexpected company prefixes: %#v", config.Company.Prefixes)
	}
}

func TestLoadRejectsInvalidValues(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".mavetis.yaml")
	content := `severity: urgent
output: xml
`
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}
	_, err := Load(path)
	if err == nil {
		t.Fatal("expected invalid config error")
	}
}

func TestLoadRules(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "rules.yaml")
	content := `rules:
  - id: custom.secret
    title: Custom secret
    message: Custom secret exposed
    category: secret
    severity: high
    confidence: high
    target: added
    any:
      - corp_[A-Za-z0-9]{8,}
    absent:
      - safe_prefix
  - id: custom.guard
    title: Guard removed
    message: Guard removed
    category: authorization
    severity: critical
    protected:
      - requireAuth
`
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write rules: %v", err)
	}
	rules, err := LoadRules(path)
	if err != nil {
		t.Fatalf("load rules: %v", err)
	}
	if len(rules) != 2 || rules[0].ID != "custom.secret" {
		t.Fatalf("unexpected rules: %#v", rules)
	}
	if len(rules[0].Any) != 1 || rules[0].Any[0] != "corp_[A-Za-z0-9]{8,}" {
		t.Fatalf("unexpected any patterns: %#v", rules[0].Any)
	}
	if len(rules[0].Absent) != 1 || rules[0].Absent[0] != "safe_prefix" {
		t.Fatalf("unexpected absent patterns: %#v", rules[0].Absent)
	}
	if rules[1].Target != "deleted" || len(rules[1].Require) != 1 {
		t.Fatalf("unexpected protected rule decoding: %#v", rules[1])
	}
}

func TestLoadRulesRejectsInvalidPatterns(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "rules.yaml")
	content := `rules:
  - id: invalid.rule
    title: Invalid rule
    message: Invalid rule
    category: inject
    require:
      - 'fetch('
`
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write rules: %v", err)
	}
	_, err := LoadRules(path)
	if err == nil {
		t.Fatal("expected invalid rules error")
	}
}

func TestLoadRulesRejectsDuplicateIDs(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "rules.yaml")
	content := `rules:
  - id: duplicate.rule
    title: First
    message: First rule
    category: inject
    require:
      - 'exec'
  - id: duplicate.rule
    title: Second
    message: Second rule
    category: inject
    require:
      - 'fetch'
`
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write rules: %v", err)
	}
	_, err := LoadRules(path)
	if err == nil {
		t.Fatal("expected duplicate rule error")
	}
}
