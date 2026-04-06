package cli

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestShellInitScriptForZshWrapsMavetisWithNoglob(t *testing.T) {
	script, err := shellInitScript("zsh")
	if err != nil {
		t.Fatalf("shell init script: %v", err)
	}
	want := "unalias mavetis 2>/dev/null || true\nalias mavetis='noglob mavetis'\n"
	if script != want {
		t.Fatalf("unexpected shell init script: %q", script)
	}
}

func TestShellInitScriptRejectsUnsupportedShell(t *testing.T) {
	_, err := shellInitScript("bash")
	if err == nil {
		t.Fatal("expected unsupported shell rejection")
	}
}

func TestShellInitScriptAllowsLiteralBracketsInZsh(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("zsh integration is not relevant on Windows")
	}
	if _, err := exec.LookPath("zsh"); err != nil {
		t.Skip("zsh is not available")
	}
	dir := t.TempDir()
	binary := filepath.Join(dir, "mavetis")
	if err := os.WriteFile(binary, []byte("#!/bin/sh\nprintf '<%s>\\n' \"$@\"\n"), 0o755); err != nil {
		t.Fatalf("write fake binary: %v", err)
	}
	script, err := shellInitScript("zsh")
	if err != nil {
		t.Fatalf("shell init script: %v", err)
	}
	loader := filepath.Join(dir, "loader.zsh")
	content := script + "mavetis review src/routes/profile/reports/[reportId]/+page.svelte\n"
	if err := os.WriteFile(loader, []byte(content), 0o600); err != nil {
		t.Fatalf("write loader: %v", err)
	}
	command := exec.Command("zsh", "-dfic", "source "+loader)
	command.Env = append(os.Environ(), "PATH="+dir+string(os.PathListSeparator)+os.Getenv("PATH"))
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("run zsh: %v\n%s", err, output)
	}
	want := "<review>\n<src/routes/profile/reports/[reportId]/+page.svelte>\n"
	if string(output) != want {
		t.Fatalf("unexpected zsh output: %q", string(output))
	}
}

func TestHelpMessageIncludesShellInit(t *testing.T) {
	message := helpMessage()
	if !strings.Contains(message, "shell init zsh") {
		t.Fatalf("expected shell init help text: %q", message)
	}
}
