package cli

import (
	"errors"
	"fmt"
	"os"

	"github.com/Pimatis/mavetis/src/config"
	"github.com/Pimatis/mavetis/src/diff"
	"github.com/Pimatis/mavetis/src/engine"
	"github.com/Pimatis/mavetis/src/git"
	"github.com/Pimatis/mavetis/src/hook"
	"github.com/Pimatis/mavetis/src/model"
	"github.com/Pimatis/mavetis/src/output"
)

func Execute(arguments []string) int {
	if len(arguments) == 0 {
		usage()
		return 0
	}
	command := arguments[0]
	if command == "help" || command == "-h" || command == "--help" {
		usage()
		return 0
	}
	if command == "review" {
		return runReview(arguments[1:], false)
	}
	if command == "ci" {
		return runReview(arguments[1:], true)
	}
	if command == "hooks" {
		return runHooks(arguments[1:])
	}
	if command == "rules" {
		return runRules(arguments[1:])
	}
	if command == "update" {
		return runUpdate(arguments[1:])
	}
	if command == "version" {
		fmt.Printf("%s %s\n", model.Name, model.Version)
		return 0
	}
	usage()
	return 1
}

func runReview(arguments []string, ci bool) int {
	spec, err := parseReview(arguments, ci)
	if err != nil {
		return fail(err)
	}
	cfg, err := config.Load(spec.ConfigPath)
	if err != nil {
		return fail(err)
	}
	merge(&cfg, spec)
	rules, err := loadAllRules(cfg, spec.RulesPath)
	if err != nil {
		return fail(err)
	}
	raw, meta, err := git.Review(spec)
	if err != nil {
		return fail(err)
	}
	parsed, err := diff.Parse(raw, meta)
	if err != nil {
		return fail(err)
	}
	parsed = diff.Filter(parsed, spec.Path)
	report, err := engine.Review(parsed, cfg, rules)
	if err != nil {
		return fail(err)
	}
	if spec.Path != "" {
		report.Meta.Mode = meta.Mode + ":" + spec.Path
	}
	if err := render(report, cfg.Output, spec.Explain); err != nil {
		return fail(err)
	}
	if blocked(report, cfg.FailOn) {
		return 1
	}
	return 0
}

func runHooks(arguments []string) int {
	if len(arguments) == 0 {
		return fail(errors.New("hooks command requires install or uninstall"))
	}
	root, err := git.Root()
	if err != nil {
		return fail(err)
	}
	if arguments[0] == "install" {
		if err := hook.Install(root); err != nil {
			return fail(err)
		}
		fmt.Println("hooks installed")
		return 0
	}
	if arguments[0] == "uninstall" {
		if err := hook.Uninstall(root); err != nil {
			return fail(err)
		}
		fmt.Println("hooks removed")
		return 0
	}
	return fail(errors.New("hooks command requires install or uninstall"))
}

func merge(config *model.Config, spec model.Review) {
	if spec.Severity != "" {
		config.Severity = spec.Severity
	}
	if spec.FailOn != "" {
		config.FailOn = spec.FailOn
	}
	if spec.Format != "" {
		config.Output = spec.Format
	}
}

func render(report model.Report, format string, explain bool) error {
	if format == "json" {
		body, err := output.JSON(report)
		if err != nil {
			return err
		}
		fmt.Println(body)
		return nil
	}
	if format == "sarif" {
		body, err := output.SARIF(report)
		if err != nil {
			return err
		}
		fmt.Println(body)
		return nil
	}
	fmt.Print(output.TextExplain(report, explain))
	return nil
}

func blocked(report model.Report, threshold string) bool {
	for _, finding := range report.Findings {
		if model.SeverityRank(finding.Severity) >= model.SeverityRank(threshold) {
			return true
		}
	}
	return false
}

func fail(err error) int {
	fmt.Fprintln(os.Stderr, err.Error())
	return 2
}
