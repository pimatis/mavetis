package cli

import (
	"testing"

	updater "github.com/Pimatis/mavetis/src/update"
)

func TestParseUpdate(t *testing.T) {
	spec, err := parseUpdate([]string{"--check"})
	if err != nil {
		t.Fatalf("parse update: %v", err)
	}
	if !spec.Check {
		t.Fatal("expected check mode")
	}
}

func TestParseUpdateRejectsUnexpectedArguments(t *testing.T) {
	_, err := parseUpdate([]string{"extra"})
	if err == nil {
		t.Fatal("expected parse error")
	}
}

func TestRunUpdateCallsUpdater(t *testing.T) {
	called := false
	previous := updateRun
	updateRun = func(spec updater.Spec) error {
		called = true
		if !spec.Check {
			t.Fatal("expected check spec")
		}
		return nil
	}
	defer func() {
		updateRun = previous
	}()
	code := runUpdate([]string{"--check"})
	if code != 0 {
		t.Fatalf("expected success, got %d", code)
	}
	if !called {
		t.Fatal("expected updater call")
	}
}
