package cli

import (
	"testing"

	"github.com/Pimatis/mavetis/src/model"
	secretscan "github.com/Pimatis/mavetis/src/secret"
)

func TestSplitSecretsScanArgumentsAllowsInterspersedFlags(t *testing.T) {
	flags, targets, err := splitSecretsScanArguments([]string{".", "--path", "src/**", "--format=json", "--no-cache", "config"})
	if err != nil {
		t.Fatalf("split: %v", err)
	}
	if len(targets) != 2 || targets[0] != "." || targets[1] != "config" {
		t.Fatalf("unexpected targets: %#v", targets)
	}
	if len(flags) != 4 || flags[0] != "--path" || flags[1] != "src/**" || flags[2] != "--format=json" || flags[3] != "--no-cache" {
		t.Fatalf("unexpected flags: %#v", flags)
	}
}

func TestRunSecretsRequiresScanSubcommand(t *testing.T) {
	if code := runSecrets(nil); code != 2 {
		t.Fatalf("expected usage error, got %d", code)
	}
}

func TestExecuteDispatchesSecretsScan(t *testing.T) {
	called := false
	previous := secretsRun
	secretsRun = func(root string, config model.Config, options secretscan.Options) (model.Report, error) {
		called = true
		if options.Path != "src/**" {
			t.Fatalf("unexpected path: %s", options.Path)
		}
		return model.Report{Meta: model.DiffMeta{Mode: "secrets"}}, nil
	}
	defer func() {
		secretsRun = previous
	}()
	if code := Execute([]string{"secrets", "scan", ".", "--path", "src/**"}); code != 0 {
		t.Fatalf("expected success, got %d", code)
	}
	if !called {
		t.Fatal("expected secrets scan dispatch")
	}
}
