package cli

import (
	"fmt"
	"sort"

	"github.com/Pimatis/mavetis/src/model"
	"github.com/Pimatis/mavetis/src/output"
)

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
	if spec.Profile != "" {
		config.Profile = spec.Profile
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
		effectiveThreshold := threshold
		if finding.EffectiveFailOn != "" {
			effectiveThreshold = finding.EffectiveFailOn
		}
		if model.SeverityRank(finding.Severity) >= model.SeverityRank(effectiveThreshold) {
			return true
		}
	}
	return false
}

func sortFindings(findings []model.Finding) {
	sort.Slice(findings, func(left int, right int) bool {
		if model.SeverityRank(findings[left].Severity) != model.SeverityRank(findings[right].Severity) {
			return model.SeverityRank(findings[left].Severity) > model.SeverityRank(findings[right].Severity)
		}
		if findings[left].Path != findings[right].Path {
			return findings[left].Path < findings[right].Path
		}
		if findings[left].Line != findings[right].Line {
			return findings[left].Line < findings[right].Line
		}
		return findings[left].RuleID < findings[right].RuleID
	})
}
