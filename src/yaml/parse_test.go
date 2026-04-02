package yaml

import "testing"

func TestParseMapAndList(t *testing.T) {
	input := `severity: high
ignore:
  - vendor/**
allow:
  values:
    - test-key
rules:
  - id: demo
    title: Demo rule
    require:
      - token
      - localStorage
`
	value, err := Parse(input)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	mapped, err := Map(value)
	if err != nil {
		t.Fatalf("map failed: %v", err)
	}
	if mapped["severity"] != "high" {
		t.Fatalf("unexpected severity: %v", mapped["severity"])
	}
	allow, err := Map(mapped["allow"])
	if err != nil {
		t.Fatalf("allow failed: %v", err)
	}
	values := Strings(allow["values"])
	if len(values) != 1 || values[0] != "test-key" {
		t.Fatalf("unexpected allow values: %#v", values)
	}
	rules, err := List(mapped["rules"])
	if err != nil {
		t.Fatalf("rules failed: %v", err)
	}
	if len(rules) != 1 {
		t.Fatalf("unexpected rule count: %d", len(rules))
	}
}

func TestParseRejectsOddIndent(t *testing.T) {
	_, err := Parse("a:\n   b: c\n")
	if err == nil {
		t.Fatal("expected indentation error")
	}
}
