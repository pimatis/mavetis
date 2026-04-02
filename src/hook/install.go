package hook

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/Pimatis/mavetis/src/git"
)

func Install(root string) error {
	hooks := filepath.Join(root, ".git", "hooks")
	if err := os.MkdirAll(hooks, 0o755); err != nil {
		return fmt.Errorf("create hooks: %w", err)
	}
	base := git.DefaultBase(root)
	precommit := filepath.Join(hooks, "pre-commit")
	prepush := filepath.Join(hooks, "pre-push")
	if err := writeHook(precommit, script("review --staged --fail-on high")); err != nil {
		return fmt.Errorf("write pre-commit: %w", err)
	}
	if err := writeHook(prepush, script("review --base "+base+" --fail-on high")); err != nil {
		return fmt.Errorf("write pre-push: %w", err)
	}
	return nil
}

func Uninstall(root string) error {
	paths := []string{
		filepath.Join(root, ".git", "hooks", "pre-commit"),
		filepath.Join(root, ".git", "hooks", "pre-push"),
	}
	for _, path := range paths {
		if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("remove hook %s: %w", path, err)
		}
	}
	return nil
}

func writeHook(path string, content string) error {
	existing, err := os.ReadFile(path)
	if err == nil {
		if string(existing) == content {
			return nil
		}
		backup := path + ".bak"
		if writeErr := os.WriteFile(backup, existing, 0o755); writeErr != nil {
			return writeErr
		}
	}
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	return os.WriteFile(path, []byte(content), 0o755)
}

func script(arguments string) string {
	return "#!/bin/sh\nset -eu\nif command -v mavetis >/dev/null 2>&1; then\n  mavetis " + arguments + "\n  exit $?\nfi\necho 'mavetis binary not found in PATH' >&2\nexit 1\n"
}
