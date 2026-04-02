package config

import (
	"fmt"
	"regexp"
	"strings"

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
	if err := ValidateProfile(data.Profile); err != nil {
		return err
	}
	if err := validateZones(data.Zones); err != nil {
		return err
	}
	return nil
}

func ValidateProfile(value string) error {
	if value == "" {
		return nil
	}
	if value == "auth" {
		return nil
	}
	if value == "fintech" {
		return nil
	}
	if value == "backend" {
		return nil
	}
	if value == "frontend" {
		return nil
	}
	return fmt.Errorf("invalid profile: %s", value)
}

func validateZones(zones model.Zones) error {
	if err := validateZonePaths(zones.Critical, "zones.critical"); err != nil {
		return err
	}
	if err := validateZonePaths(zones.Restricted, "zones.restricted"); err != nil {
		return err
	}
	return nil
}

func validateZonePaths(paths []string, field string) error {
	seen := map[string]struct{}{}
	for _, item := range paths {
		value := strings.TrimSpace(item)
		if value == "" {
			return fmt.Errorf("invalid %s: empty path", field)
		}
		if _, ok := seen[value]; ok {
			return fmt.Errorf("invalid %s: duplicate path %s", field, value)
		}
		seen[value] = struct{}{}
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
