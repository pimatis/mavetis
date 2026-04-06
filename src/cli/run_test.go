package cli

import (
	"testing"

	"github.com/Pimatis/mavetis/src/model"
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

func TestExecuteDispatchesShell(t *testing.T) {
	if code := Execute([]string{"shell", "init", "zsh"}); code != 0 {
		t.Fatalf("expected zero exit code for shell init, got %d", code)
	}
}

func TestExecuteVersionShortFlagReturnsZero(t *testing.T) {
	if code := Execute([]string{"-v"}); code != 0 {
		t.Fatalf("expected zero exit code for short version flag, got %d", code)
	}
}

func TestBlockedUsesEffectiveFailOnWhenZonePolicyApplies(t *testing.T) {
	report := model.Report{Findings: []model.Finding{{Severity: "low", EffectiveFailOn: "low"}}}
	if !blocked(report, "critical") {
		t.Fatal("expected zone policy to block")
	}
}
