package cli

import (
	"errors"
	"flag"
	"os"

	"github.com/Pimatis/mavetis/src/model"
)

func parseReview(arguments []string, ci bool) (model.Review, error) {
	spec := model.Review{}
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
	flags.BoolVar(&spec.Explain, "explain", false, "Include finding reasons in text output")
	if err := flags.Parse(arguments); err != nil {
		return spec, err
	}
	if ci {
		spec.Base = defaultBase(spec.Base)
		spec.Mode = "ci"
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
	remaining := flags.Args()
	if len(remaining) != 0 {
		return spec, errors.New("unexpected positional arguments")
	}
	spec.Mode = "review"
	if err := validateReview(spec); err != nil {
		return spec, err
	}
	return spec, nil
}

func defaultBase(value string) string {
	if value != "" {
		return value
	}
	return "main"
}

func helpMessage() string {
	return `mavetis commands:
  review --staged [--path src/**] [--profile auth] [--explain]
  review --base main [--path src/**] [--profile backend]
  ci --base main [--path src/**] [--profile fintech]
  hooks install
  hooks uninstall
  rules validate --rules rules.yaml
  rules list [--rules rules.yaml]
  rules show --id rule.id [--rules rules.yaml]
  rules test --diff sample.diff [--rules rules.yaml]
  rules matrix [--rules rules.yaml]
  update [--check]
  version

regression core:
  - security downgrade detection for cookies, bcrypt, timeouts, rate limits, and MFA
  - config drift detection for debug mode, CSP, CORS, TLS, and deployable container settings
  - observability leak detection for request bodies, auth material, PII, raw errors, and traces

policy layer:
  - rule profiles: auth, fintech, backend, frontend
  - trust zones from config: zones.critical and zones.restricted
  - automatic severity uplift and stricter fail-on inside protected paths

examples:
  mavetis review --staged --path 'src/**' --profile auth --explain
  mavetis review --base main --path 'src/**' --profile backend
  mavetis ci --base main --format json --profile fintech
  mavetis rules validate --rules rules.yaml
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
