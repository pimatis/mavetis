package wizard

import (
	"strings"
	"testing"
)

func TestGenerateIncludesHeader(t *testing.T) {
	template := ConfigTemplate{Profile: "backend", Severity: "low", FailOn: "high", Output: "text"}
	out := Generate(template)
	if !strings.Contains(out, "# Mavetis") {
		t.Fatal("expected header comment")
	}
}

func TestGenerateIncludesProfile(t *testing.T) {
	template := ConfigTemplate{Profile: "backend", Severity: "low", FailOn: "high", Output: "text"}
	out := Generate(template)
	if !strings.Contains(out, "profile: backend") {
		t.Fatal("expected profile")
	}
}

func TestGenerateIncludesZones(t *testing.T) {
	template := ConfigTemplate{
		Profile:    "auth",
		Severity:   "low",
		FailOn:     "high",
		Output:     "text",
		Critical:   []string{"src/auth/**"},
		Restricted: []string{"src/api/**"},
	}
	out := Generate(template)
	if !strings.Contains(out, "zones:") {
		t.Fatal("expected zones section")
	}
	if !strings.Contains(out, "src/auth/**") {
		t.Fatal("expected critical zone")
	}
	if !strings.Contains(out, "src/api/**") {
		t.Fatal("expected restricted zone")
	}
}

func TestGenerateOmitsEmptyZones(t *testing.T) {
	template := ConfigTemplate{Profile: "auth", Severity: "low", FailOn: "high", Output: "text"}
	out := Generate(template)
	if strings.Contains(out, "zones:") {
		t.Fatal("expected no zones section")
	}
}

func TestGenerateDeterministic(t *testing.T) {
	template := ConfigTemplate{Profile: "auth", Severity: "low", FailOn: "high", Output: "text", Ignore: []string{"vendor/**"}}
	a := Generate(template)
	b := Generate(template)
	if a != b {
		t.Fatal("expected deterministic output")
	}
}
