package cli

import (
	"testing"

	updater "github.com/Pimatis/mavetis/src/update"
)

func TestExecuteHelpReturnsZero(t *testing.T) {
	if code := Execute([]string{"-h"}); code != 0 {
		t.Fatalf("expected zero exit code for help, got %d", code)
	}
	if code := Execute([]string{"--help"}); code != 0 {
		t.Fatalf("expected zero exit code for long help, got %d", code)
	}
	if code := Execute([]string{"help"}); code != 0 {
		t.Fatalf("expected zero exit code for help command, got %d", code)
	}
}

func TestExecuteWithoutArgumentsReturnsZero(t *testing.T) {
	if code := Execute(nil); code != 0 {
		t.Fatalf("expected zero exit code without arguments, got %d", code)
	}
}

func TestExecuteDispatchesUpdate(t *testing.T) {
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
	if code := Execute([]string{"update", "--check"}); code != 0 {
		t.Fatalf("expected zero exit code for update, got %d", code)
	}
	if !called {
		t.Fatal("expected update dispatch")
	}
}
