package cli

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Pimatis/mavetis/src/scan"
	"github.com/Pimatis/mavetis/src/wizard"
)

func runInit(arguments []string) int {
	flags := flag.NewFlagSet("init", flag.ContinueOnError)
	flags.SetOutput(os.Stderr)
	defaultMode := flags.Bool("default", false, "Create default config without prompts")
	force := flags.Bool("force", false, "Overwrite existing config")
	if err := flags.Parse(arguments); err != nil {
		return 2
	}
	root, err := scan.Root()
	if err != nil {
		return fail(err)
	}
	configPath := filepath.Join(root, ".mavetis.yaml")
	if !*force {
		if _, err := os.Stat(configPath); err == nil {
			return fail(errors.New(".mavetis.yaml already exists; use --force to overwrite"))
		}
	}
	project := wizard.Detect(root)
	template := wizard.ConfigTemplate{
		Profile:    project.Profile,
		Severity:   "low",
		FailOn:     "high",
		Output:     "text",
		Ignore:     project.Ignore,
		Critical:   project.Critical,
		Restricted: project.Restricted,
	}
	if !*defaultMode {
		reader := bufio.NewReader(os.Stdin)
		template = wizard.RunInteractive(reader, project)
	}
	content := wizard.Generate(template)
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		return fail(fmt.Errorf("write config: %w", err))
	}
	if err := appendGitignore(root, ".mavetis.yaml"); err != nil {
		return fail(fmt.Errorf("update .gitignore: %w", err))
	}
	fmt.Printf("Created %s\n", configPath)
	return 0
}

func appendGitignore(root string, entry string) error {
	path := filepath.Join(root, ".gitignore")
	existing := ""
	data, err := os.ReadFile(path)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return err
		}
	} else {
		existing = string(data)
		if strings.Contains(existing, entry) {
			return nil
		}
	}
	var b strings.Builder
	if existing != "" {
		b.WriteString(existing)
		if !strings.HasSuffix(existing, "\n") {
			b.WriteString("\n")
		}
	}
	b.WriteString(entry)
	b.WriteString("\n")
	return os.WriteFile(path, []byte(b.String()), 0644)
}
