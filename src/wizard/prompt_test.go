package wizard

import (
	"bufio"
	"strings"
	"testing"
)

func TestPromptReturnsDefaultOnEmpty(t *testing.T) {
	reader := bufio.NewReader(strings.NewReader("\n"))
	result := prompt(reader, "Test", "default")
	if result != "default" {
		t.Fatalf("expected default, got %s", result)
	}
}

func TestPromptReturnsInput(t *testing.T) {
	reader := bufio.NewReader(strings.NewReader("custom\n"))
	result := prompt(reader, "Test", "default")
	if result != "custom" {
		t.Fatalf("expected custom, got %s", result)
	}
}

func TestPromptYesNoDefaultYes(t *testing.T) {
	reader := bufio.NewReader(strings.NewReader("\n"))
	if !promptYesNo(reader, "Test", true) {
		t.Fatal("expected true for default yes")
	}
}

func TestPromptYesNoReturnsNo(t *testing.T) {
	reader := bufio.NewReader(strings.NewReader("n\n"))
	if promptYesNo(reader, "Test", true) {
		t.Fatal("expected false")
	}
}

func TestRunInteractiveAcceptsDefaults(t *testing.T) {
	reader := bufio.NewReader(strings.NewReader("\n\n\n\n\n\n"))
	project := Project{Profile: "backend", Ignore: []string{"vendor/**"}, Critical: []string{"auth/**"}, Restricted: []string{"api/**"}}
	template := RunInteractive(reader, project)
	if template.Profile != "backend" {
		t.Fatalf("expected backend profile, got %s", template.Profile)
	}
	if template.FailOn != "high" {
		t.Fatalf("expected high fail-on, got %s", template.FailOn)
	}
}
