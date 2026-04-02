package config

import (
	"fmt"
	"os"

	"github.com/Pimatis/mavetis/src/model"
	"github.com/Pimatis/mavetis/src/yaml"
)

func LoadRules(path string) ([]model.Rule, error) {
	if path == "" {
		return nil, nil
	}
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read rules: %w", err)
	}
	value, err := yaml.Parse(string(content))
	if err != nil {
		return nil, fmt.Errorf("parse rules: %w", err)
	}
	mapped, err := yaml.Map(value)
	if err != nil {
		return nil, fmt.Errorf("decode rules: %w", err)
	}
	items, err := yaml.List(mapped["rules"])
	if err != nil {
		return nil, fmt.Errorf("decode rules: expected rules list")
	}
	rules := make([]model.Rule, 0, len(items))
	seen := map[string]struct{}{}
	for _, item := range items {
		entry, err := yaml.Map(item)
		if err != nil {
			return nil, fmt.Errorf("decode rule: %w", err)
		}
		rule := model.Rule{}
		decodeRule(entry, &rule)
		if rule.ID == "" {
			return nil, fmt.Errorf("decode rule: missing id")
		}
		if rule.Title == "" {
			return nil, fmt.Errorf("decode rule: missing title for %s", rule.ID)
		}
		if rule.Type == "" && len(rule.Require) == 0 && len(rule.Any) == 0 {
			return nil, fmt.Errorf("decode rule: missing require or any for %s", rule.ID)
		}
		if rule.Severity == "" {
			rule.Severity = "medium"
		}
		if rule.Confidence == "" {
			rule.Confidence = "medium"
		}
		if rule.Target == "" {
			rule.Target = "added"
		}
		if _, ok := seen[rule.ID]; ok {
			return nil, fmt.Errorf("decode rule: duplicate id %s", rule.ID)
		}
		if err := validateRule(rule); err != nil {
			return nil, err
		}
		seen[rule.ID] = struct{}{}
		rules = append(rules, rule)
	}
	return rules, nil
}

func mergeStrings(sets ...[]string) []string {
	values := make([]string, 0)
	for _, set := range sets {
		for _, item := range set {
			if item == "" {
				continue
			}
			values = append(values, item)
		}
	}
	return values
}

func decodeRule(mapped map[string]any, rule *model.Rule) {
	rule.ID, _ = yaml.String(mapped["id"])
	rule.Type, _ = yaml.String(mapped["type"])
	rule.Title, _ = yaml.String(mapped["title"])
	rule.Message, _ = yaml.String(mapped["message"])
	rule.Remediation, _ = yaml.String(mapped["remediation"])
	rule.Category, _ = yaml.String(mapped["category"])
	rule.Severity, _ = yaml.String(mapped["severity"])
	rule.Confidence, _ = yaml.String(mapped["confidence"])
	rule.Target, _ = yaml.String(mapped["target"])
	rule.Paths = yaml.Strings(mapped["paths"])
	rule.FromPaths = yaml.Strings(mapped["from-paths"])
	rule.ForbiddenPaths = yaml.Strings(mapped["forbidden-paths"])
	rule.Ignore = yaml.Strings(mapped["ignore"])
	rule.Require = mergeStrings(yaml.Strings(mapped["require"]), yaml.Strings(mapped["forbidden"]), yaml.Strings(mapped["protected"]), yaml.Strings(mapped["required"]), yaml.Strings(mapped["when"]))
	rule.Any = mergeStrings(yaml.Strings(mapped["any"]), yaml.Strings(mapped["matchany"]))
	rule.Near = mergeStrings(yaml.Strings(mapped["near"]), yaml.Strings(mapped["context"]))
	rule.Absent = mergeStrings(yaml.Strings(mapped["absent"]), yaml.Strings(mapped["mitigate"]))
	rule.Imports = mergeStrings(yaml.Strings(mapped["imports"]), yaml.Strings(mapped["forbidden-imports"]))
	rule.Calls = mergeStrings(yaml.Strings(mapped["calls"]), yaml.Strings(mapped["required-calls"]))
	rule.Middleware = mergeStrings(yaml.Strings(mapped["middleware"]), yaml.Strings(mapped["required-middleware"]))
	rule.Keys = mergeStrings(yaml.Strings(mapped["keys"]), yaml.Strings(mapped["envs"]), yaml.Strings(mapped["forbidden-env"]))
	rule.AllowedValues = mergeStrings(yaml.Strings(mapped["allowed-values"]), yaml.Strings(mapped["allow-values"]))
	rule.ForbiddenValues = mergeStrings(yaml.Strings(mapped["forbidden-values"]), yaml.Strings(mapped["deny-values"]))
	rule.ConstraintKey, _ = yaml.String(mapped["key"])
	rule.ConstraintPattern, _ = yaml.String(mapped["pattern"])
	rule.Standards = yaml.Strings(mapped["standards"])
	if len(yaml.Strings(mapped["protected"])) != 0 && rule.Target == "" {
		rule.Target = "deleted"
	}
	if rule.Type == "deletedLineGuard" && rule.Target == "" {
		rule.Target = "deleted"
	}
	minValue, ok := yaml.Float(mapped["min"])
	if ok {
		rule.MinValue = minValue
	}
	maxValue, ok := yaml.Float(mapped["max"])
	if ok {
		rule.MaxValue = maxValue
	}
	entropy, ok := yaml.Float(mapped["entropy"])
	if ok {
		rule.Entropy = entropy
	}
	mask, ok := yaml.Bool(mapped["mask"])
	if ok {
		rule.Mask = mask
	}
}
