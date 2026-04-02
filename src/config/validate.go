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
	if err := validateSupply(data.Supply); err != nil {
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
	if err := validateRuleType(data); err != nil {
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
	if err := compileAll(data.Imports); err != nil {
		return err
	}
	if err := compileAll(data.Calls); err != nil {
		return err
	}
	if err := compileAll(data.Middleware); err != nil {
		return err
	}
	if err := compileAll(data.Keys); err != nil {
		return err
	}
	if data.ConstraintPattern != "" {
		if err := compileAll([]string{data.ConstraintPattern}); err != nil {
			return err
		}
	}
	return nil
}

func validateSupply(data model.Supply) error {
	if err := validateListEntries(data.AllowPackages, "supply.allow-packages"); err != nil {
		return err
	}
	if err := validateListEntries(data.DenyPackages, "supply.deny-packages"); err != nil {
		return err
	}
	if err := validateListEntries(data.TrustedRegistries, "supply.trusted-registries"); err != nil {
		return err
	}
	return nil
}

func validateListEntries(values []string, field string) error {
	seen := map[string]struct{}{}
	for _, item := range values {
		value := strings.TrimSpace(item)
		if value == "" {
			return fmt.Errorf("invalid %s: empty value", field)
		}
		if _, ok := seen[value]; ok {
			return fmt.Errorf("invalid %s: duplicate value %s", field, value)
		}
		seen[value] = struct{}{}
	}
	return nil
}

func validateRuleType(data model.Rule) error {
	if data.Type == "" {
		return nil
	}
	if data.Type == "forbiddenImport" {
		if len(data.Imports) == 0 {
			return fmt.Errorf("invalid rule %s: forbiddenImport requires imports", data.ID)
		}
		return nil
	}
	if data.Type == "deletedLineGuard" {
		if len(data.Require) == 0 {
			return fmt.Errorf("invalid rule %s: deletedLineGuard requires require", data.ID)
		}
		return nil
	}
	if data.Type == "forbiddenEnv" {
		if len(data.Keys) == 0 {
			return fmt.Errorf("invalid rule %s: forbiddenEnv requires keys", data.ID)
		}
		return nil
	}
	if data.Type == "requiredMiddleware" {
		if len(data.Middleware) == 0 {
			return fmt.Errorf("invalid rule %s: requiredMiddleware requires middleware", data.ID)
		}
		return nil
	}
	if data.Type == "requiredCall" {
		if len(data.Calls) == 0 {
			return fmt.Errorf("invalid rule %s: requiredCall requires calls", data.ID)
		}
		return nil
	}
	if data.Type == "configKeyConstraint" {
		if data.ConstraintKey == "" && len(data.Keys) == 0 {
			return fmt.Errorf("invalid rule %s: configKeyConstraint requires key", data.ID)
		}
		return nil
	}
	if data.Type == "pathBoundary" {
		if len(data.Imports) == 0 && len(data.ForbiddenPaths) == 0 {
			return fmt.Errorf("invalid rule %s: pathBoundary requires imports or forbidden-paths", data.ID)
		}
		return nil
	}
	return fmt.Errorf("invalid rule type: %s", data.Type)
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
