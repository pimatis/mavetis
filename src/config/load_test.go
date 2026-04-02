package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".mavetis.yaml")
	snapshotPath := filepath.Join(dir, "snapshots.yaml")
	content := `severity: medium
failon: critical
profile: auth
ignore:
  - vendor/**
allow:
  values:
    - safe-example
company:
  prefixes:
    - corp_
supply:
  allow-packages:
    - '@company/*'
  deny-packages:
    - left-pad
  trusted-registries:
    - registry.company.local
snapshot:
  path: ` + snapshotPath + `
zones:
  critical:
    - src/auth/**
  restricted:
    - src/api/admin/**
`
	snapshotContent := `snapshots:
  - id: auth.service.verify
    path: src/auth/service.go
    anchor: verifyToken
    category: token
    severity: high
    require:
      - verify
`
	if err := os.WriteFile(snapshotPath, []byte(snapshotContent), 0o600); err != nil {
		t.Fatalf("write snapshots: %v", err)
	}
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
	if config.Profile != "auth" {
		t.Fatalf("unexpected profile: %#v", config.Profile)
	}
	if len(config.Zones.Critical) != 1 || config.Zones.Critical[0] != "src/auth/**" {
		t.Fatalf("unexpected critical zones: %#v", config.Zones.Critical)
	}
	if len(config.Zones.Restricted) != 1 || config.Zones.Restricted[0] != "src/api/admin/**" {
		t.Fatalf("unexpected restricted zones: %#v", config.Zones.Restricted)
	}
	if len(config.Supply.DenyPackages) != 1 || config.Supply.DenyPackages[0] != "left-pad" {
		t.Fatalf("unexpected supply policy: %#v", config.Supply)
	}
	if config.Snapshot.Path != snapshotPath || len(config.Snapshots) != 1 {
		t.Fatalf("unexpected snapshots: %#v %#v", config.Snapshot, config.Snapshots)
	}
}

func TestLoadRejectsInvalidValues(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".mavetis.yaml")
	content := `severity: urgent
output: xml
profile: invalid
`
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}
	_, err := Load(path)
	if err == nil {
		t.Fatal("expected invalid config error")
	}
}

func TestLoadRejectsInvalidZones(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".mavetis.yaml")
	content := `zones:
  critical:
    - src/auth/**
    - src/auth/**
  restricted:
    - ''
`
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}
	_, err := Load(path)
	if err == nil {
		t.Fatal("expected invalid zone config error")
	}
}

func TestLoadRejectsInvalidSupplyPolicy(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".mavetis.yaml")
	content := `supply:
  allow-packages:
    - '@company/*'
    - '@company/*'
`
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}
	_, err := Load(path)
	if err == nil {
		t.Fatal("expected invalid supply policy error")
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
  - id: custom.boundary
    type: forbiddenImport
    title: Boundary
    message: Boundary violated
    remediation: Fix it
    category: boundary
    severity: high
    confidence: high
    target: added
    paths:
      - src/ui/**
    imports:
      - '(?i)auth'
`
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write rules: %v", err)
	}
	rules, err := LoadRules(path)
	if err != nil {
		t.Fatalf("load rules: %v", err)
	}
	if len(rules) != 3 || rules[0].ID != "custom.secret" {
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
	if rules[2].Type != "forbiddenImport" || len(rules[2].Imports) != 1 {
		t.Fatalf("unexpected typed rule decoding: %#v", rules[2])
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
