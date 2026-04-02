package config

import (
	"fmt"
	"regexp"

	"github.com/Pimatis/mavetis/src/model"
)

func ValidateSeverity(value string, field string) error {
	if value == "" {
		return nil
	}
	if value == "low" {
		return nil
	}
	if value == "medium" {
		return nil
	}
	if value == "high" {
		return nil
	}
	if value == "critical" {
		return nil
	}
	return fmt.Errorf("invalid %s: %s", field, value)
}

func ValidateOutput(value string) error {
	if value == "" {
		return nil
	}
	if value == "text" {
		return nil
	}
	if value == "json" {
		return nil
	}
	if value == "sarif" {
		return nil
	}
	return fmt.Errorf("invalid output: %s", value)
}

func validateConfig(data model.Config) error {
	if err := ValidateSeverity(data.Severity, "severity"); err != nil {
		return err
	}
	if err := ValidateSeverity(data.FailOn, "fail-on"); err != nil {
		return err
	}
	if err := ValidateOutput(data.Output); err != nil {
		return err
	}
	return nil
}

func validateRule(data model.Rule) error {
	if err := ValidateSeverity(data.Severity, "rule severity"); err != nil {
		return err
	}
	if err := validateConfidence(data.Confidence); err != nil {
		return err
	}
	if err := validateTarget(data.Target); err != nil {
		return err
	}
	if err := compileAll(data.Require); err != nil {
		return err
	}
	if err := compileAll(data.Any); err != nil {
		return err
	}
	if err := compileAll(data.Near); err != nil {
		return err
	}
	if err := compileAll(data.Absent); err != nil {
		return err
	}
	return nil
}

func validateConfidence(value string) error {
	if value == "low" {
		return nil
	}
	if value == "medium" {
		return nil
	}
	if value == "high" {
		return nil
	}
	return fmt.Errorf("invalid rule confidence: %s", value)
}

func validateTarget(value string) error {
	if value == "added" {
		return nil
	}
	if value == "deleted" {
		return nil
	}
	return fmt.Errorf("invalid rule target: %s", value)
}

func compileAll(patterns []string) error {
	for _, pattern := range patterns {
		if _, err := regexp.Compile(pattern); err != nil {
			return fmt.Errorf("invalid rule pattern %q: %w", pattern, err)
		}
	}
	return nil
}
