package cli

import (
	"errors"
	"flag"
	"os"
	"strings"

	"github.com/Pimatis/mavetis/src/config"
	"github.com/Pimatis/mavetis/src/model"
	"github.com/Pimatis/mavetis/src/scan"
	secretscan "github.com/Pimatis/mavetis/src/secret"
)

var secretsRun = secretscan.Scan

func runSecrets(arguments []string) int {
	if len(arguments) == 0 {
		return fail(errors.New("secrets command requires scan"))
	}
	if arguments[0] == "scan" {
		return runSecretsScan(arguments[1:])
	}
	return fail(errors.New("secrets command requires scan"))
}

func runSecretsScan(arguments []string) int {
	flags := flag.NewFlagSet("secrets scan", flag.ContinueOnError)
	flags.SetOutput(os.Stderr)
	format := flags.String("format", "", "Output format: text, json, sarif")
	severity := flags.String("severity", "", "Minimum severity: low, medium, high, critical")
	failOn := flags.String("fail-on", "", "Fail threshold: low, medium, high, critical")
	configPath := flags.String("config", "", "Config path")
	path := flags.String("path", "", "Limit scan to a path glob")
	cachePath := flags.String("cache", "", "Secrets scan cache path")
	noCache := flags.Bool("no-cache", false, "Disable incremental secrets scan cache")
	flagArguments, targets, err := splitSecretsScanArguments(arguments)
	if err != nil {
		return fail(err)
	}
	if err := flags.Parse(flagArguments); err != nil {
		return fail(err)
	}
	targets = append(targets, flags.Args()...)
	if err := config.ValidateOutput(*format); err != nil {
		return fail(err)
	}
	if err := config.ValidateSeverity(*severity, "severity"); err != nil {
		return fail(err)
	}
	if err := config.ValidateSeverity(*failOn, "fail-on"); err != nil {
		return fail(err)
	}
	cfg, err := config.Load(*configPath)
	if err != nil {
		return fail(err)
	}
	applySecretsOverrides(&cfg, *format, *severity, *failOn)
	root, err := scan.Root()
	if err != nil {
		return fail(err)
	}
	report, err := secretsRun(root, cfg, secretscan.Options{Targets: targets, Path: *path, Cache: *cachePath, NoCache: *noCache})
	if err != nil {
		return fail(err)
	}
	if err := render(report, cfg.Output, false); err != nil {
		return fail(err)
	}
	if blocked(report, cfg.FailOn) {
		return 1
	}
	return 0
}

func splitSecretsScanArguments(arguments []string) ([]string, []string, error) {
	flagArguments := make([]string, 0, len(arguments))
	targets := make([]string, 0, len(arguments))
	valueFlags := map[string]struct{}{
		"--format":   {},
		"--severity": {},
		"--fail-on":  {},
		"--config":   {},
		"--path":     {},
		"--cache":    {},
	}
	boolFlags := map[string]struct{}{
		"--no-cache": {},
	}
	consumeTargets := false
	for index := 0; index < len(arguments); index++ {
		argument := arguments[index]
		if consumeTargets {
			targets = append(targets, argument)
			continue
		}
		if argument == "--" {
			consumeTargets = true
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
		targets = append(targets, argument)
	}
	return flagArguments, targets, nil
}

func applySecretsOverrides(configData *model.Config, format string, severity string, failOn string) {
	if format != "" {
		configData.Output = format
	}
	if severity != "" {
		configData.Severity = severity
	}
	if failOn != "" {
		configData.FailOn = failOn
	}
}
