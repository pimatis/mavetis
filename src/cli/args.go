package cli

import (
	"bufio"
	"errors"
	"flag"
	"os"
	"strings"

	"github.com/Pimatis/mavetis/src/model"
)

const maxFileTargets = 128

func parseReview(arguments []string, ci bool) (model.Review, error) {
	spec := model.Review{}
	flagArguments, fileArguments, err := splitReviewArguments(arguments)
	if err != nil {
		return spec, err
	}
	flags := flag.NewFlagSet("review", flag.ContinueOnError)
	flags.SetOutput(os.Stderr)
	flags.BoolVar(&spec.Staged, "staged", false, "Review staged changes")
	flags.StringVar(&spec.Base, "base", "", "Base branch or ref")
	flags.StringVar(&spec.Head, "head", "", "Head branch or ref")
	flags.StringVar(&spec.Format, "format", "", "Output format: text, json, sarif")
	flags.StringVar(&spec.Severity, "severity", "", "Minimum severity: low, medium, high, critical")
	flags.StringVar(&spec.FailOn, "fail-on", "", "Fail threshold: low, medium, high, critical")
	flags.StringVar(&spec.Profile, "profile", "", "Rule profile: auth, fintech, backend, frontend")
	flags.StringVar(&spec.ConfigPath, "config", "", "Config path")
	flags.StringVar(&spec.RulesPath, "rules", "", "Custom rules path")
	flags.StringVar(&spec.Path, "path", "", "Limit review to a path glob")
	flags.StringVar(&spec.BaselinePath, "baseline", "", "Baseline file path")
	flags.BoolVar(&spec.Explain, "explain", false, "Include finding reasons in text output")
	flags.BoolVar(&spec.WithSuggested, "with-suggested", false, "Review bounded suggested local dependencies together with the requested files")
	flags.BoolVar(&spec.WithSuggested, "follow-imports", false, "Review bounded suggested local dependencies together with the requested files")
	flags.BoolVar(&spec.StdinTargets, "stdin-targets", false, "Read newline-separated review targets from stdin")
	if err := flags.Parse(flagArguments); err != nil {
		return spec, err
	}
	remaining := flags.Args()
	if len(remaining) != 0 {
		return spec, errors.New("unexpected positional arguments")
	}
	for _, argument := range fileArguments {
		target := normalizeReviewTarget(argument)
		if target == "" {
			return spec, errors.New("empty @file target")
		}
		spec.Files = append(spec.Files, target)
		if len(spec.Files) > maxFileTargets {
			return spec, errors.New("too many @file targets")
		}
	}
	if spec.StdinTargets {
		targets, readErr := readReviewTargets(os.Stdin)
		if readErr != nil {
			return spec, readErr
		}
		spec.Files = append(spec.Files, targets...)
		if len(spec.Files) == 0 {
			return spec, errors.New("stdin produced no review targets")
		}
		if len(spec.Files) > maxFileTargets {
			return spec, errors.New("too many @file targets")
		}
	}
	if ci {
		spec.Base = defaultBase(spec.Base)
		spec.Mode = "ci"
	}
	if len(spec.Files) != 0 {
		if ci {
			return spec, errors.New("ci mode does not support @file targets")
		}
		if spec.Staged || spec.Base != "" || spec.Head != "" {
			return spec, errors.New("@file targets cannot be combined with --staged, --base, or --head")
		}
		spec.Mode = "file"
		if err := validateReview(spec); err != nil {
			return spec, err
		}
		return spec, nil
	}
	if !ci && spec.Staged {
		spec.Mode = "review"
		if err := validateReview(spec); err != nil {
			return spec, err
		}
		return spec, nil
	}
	if spec.Base != "" {
		spec.Mode = "review"
		if err := validateReview(spec); err != nil {
			return spec, err
		}
		return spec, nil
	}
	if ci {
		return spec, validateReview(spec)
	}
	spec.Mode = "review"
	if err := validateReview(spec); err != nil {
		return spec, err
	}
	return spec, nil
}

func splitReviewArguments(arguments []string) ([]string, []string, error) {
	flagArguments := make([]string, 0, len(arguments))
	fileArguments := make([]string, 0, len(arguments))
	valueFlags := map[string]struct{}{
		"--base":     {},
		"--head":     {},
		"--format":   {},
		"--severity": {},
		"--fail-on":  {},
		"--profile":  {},
		"--config":   {},
		"--rules":    {},
		"--path":     {},
		"--baseline": {},
	}
	boolFlags := map[string]struct{}{
		"--staged":         {},
		"--explain":        {},
		"--with-suggested": {},
		"--follow-imports": {},
		"--stdin-targets":  {},
	}
	consumeFiles := false
	for index := 0; index < len(arguments); index++ {
		argument := arguments[index]
		if consumeFiles {
			fileArguments = append(fileArguments, argument)
			continue
		}
		if argument == "--" {
			consumeFiles = true
			continue
		}
		if strings.HasPrefix(argument, "--") {
			name := argument
			if cut := strings.Index(argument, "="); cut >= 0 {
				name = argument[:cut]
			}
			if _, ok := valueFlags[name]; ok {
				flagArguments = append(flagArguments, argument)
				if strings.Contains(argument, "=") {
					continue
				}
				if index+1 >= len(arguments) {
					return nil, nil, errors.New("missing value for " + name)
				}
				index++
				flagArguments = append(flagArguments, arguments[index])
				continue
			}
			if _, ok := boolFlags[name]; ok {
				flagArguments = append(flagArguments, argument)
				continue
			}
			flagArguments = append(flagArguments, argument)
			continue
		}
		if strings.HasPrefix(argument, "-") {
			flagArguments = append(flagArguments, argument)
			continue
		}
		fileArguments = append(fileArguments, argument)
	}
	return flagArguments, fileArguments, nil
}

func normalizeReviewTarget(argument string) string {
	return strings.TrimSpace(strings.TrimPrefix(argument, "@"))
}

func readReviewTargets(file *os.File) ([]string, error) {
	scanner := bufio.NewScanner(file)
	targets := make([]string, 0)
	for scanner.Scan() {
		line := normalizeReviewTarget(scanner.Text())
		if line == "" {
			continue
		}
		targets = append(targets, line)
		if len(targets) > maxFileTargets {
			return nil, errors.New("too many @file targets")
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return targets, nil
}

func defaultBase(value string) string {
	if value != "" {
		return value
	}
	return "main"
}

func helpMessage() string {
	return `mavetis commands:
  review --staged [--path src/**] [--profile auth] [--explain] [--baseline .mavetis-baseline.yaml]
  review --base main [--path src/**] [--profile backend] [--baseline .mavetis-baseline.yaml]
  review src/file.go [--with-suggested] [--format json]
  ci --base main [--path src/**] [--profile fintech] [--baseline .mavetis-baseline.yaml]
  init [--default] [--force]
  baseline --create [--output .mavetis-baseline.yaml] [--base main]
  hooks install
  hooks uninstall
  shell init zsh
  rules validate --rules rules.yaml
  rules list [--rules rules.yaml]
  rules show --id rule.id [--rules rules.yaml]
  rules test --diff sample.diff [--rules rules.yaml]
  rules matrix [--rules rules.yaml] [--profile auth]
  rules snapshot --output snapshots.yaml [--path src/auth/**]
  update [--check]
  version
  -v, --version

file review:
  mavetis review src/auth/login.go src/api/handler.ts --explain
  mavetis review src/rule --with-suggested
  mavetis review @config/nginx.conf --profile backend --format json
  mavetis review src/auth/*.go --severity high
  printf '%s\n' src/rule/token.go src/rule/scope.go | mavetis review --stdin-targets
  mavetis review src/scan/load.go --with-suggested

examples:
  mavetis review --staged --path 'src/**' --profile auth --explain
  mavetis review --base main --path 'src/**' --profile backend --baseline .mavetis-baseline.yaml
  mavetis review src/scan/load.go --with-suggested
  mavetis ci --base main --format json --profile fintech --baseline .mavetis-baseline.yaml
  mavetis init
  mavetis init --force
  mavetis baseline --create --base main
  mavetis rules validate --rules rules.yaml
  mavetis rules snapshot --output .mavetis-snapshots.yaml --path 'src/auth/**'
  mavetis update --check
  mavetis hooks install

exit codes:
  0 no blocking findings or help output
  1 blocking findings matched --fail-on
  2 usage or runtime error`
}

func usage() {
	println(helpMessage())
}
