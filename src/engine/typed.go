package engine

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/Pimatis/mavetis/src/match"
	"github.com/Pimatis/mavetis/src/model"
)

var envLine = regexp.MustCompile(`^\s*([A-Za-z_][A-Za-z0-9_\.-]*)\s*[:=]\s*(.+)$`)
var importLine = regexp.MustCompile(`(?i)(import\s+.*from\s+["']([^"']+)["']|require\(["']([^"']+)["']\)|import\s+["']([^"']+)["'])`)
var configLine = regexp.MustCompile(`(?i)^\s*["']?([A-Za-z0-9_\.-]+)["']?\s*[:=]\s*([^#]+)$`)

func evaluateTyped(item compiled, path string, hunkText string, line model.DiffLine) (model.Finding, bool) {
	if item.rule.Type == "forbiddenImport" {
		return evaluateForbiddenImport(item, path, line)
	}
	if item.rule.Type == "deletedLineGuard" {
		return evaluateDeletedLineGuard(item, path, line)
	}
	if item.rule.Type == "forbiddenEnv" {
		return evaluateForbiddenEnv(item, path, line)
	}
	if item.rule.Type == "requiredMiddleware" {
		return evaluateRequiredMiddleware(item, path, hunkText, line)
	}
	if item.rule.Type == "requiredCall" {
		return evaluateRequiredCall(item, path, hunkText, line)
	}
	if item.rule.Type == "configKeyConstraint" {
		return evaluateConfigKeyConstraint(item, path, line)
	}
	if item.rule.Type == "pathBoundary" {
		return evaluatePathBoundary(item, path, line)
	}
	return model.Finding{}, false
}

func evaluateForbiddenImport(item compiled, path string, line model.DiffLine) (model.Finding, bool) {
	module := importTarget(line.Text)
	if module == "" {
		return model.Finding{}, false
	}
	if !matchAny(item.imports, module) && !matchAny(item.imports, line.Text) {
		return model.Finding{}, false
	}
	reasons := []string{"source path matched the typed import boundary", "added import path: " + module}
	return typedFinding(item, path, line, reasons...), true
}

func evaluateDeletedLineGuard(item compiled, path string, line model.DiffLine) (model.Finding, bool) {
	if line.Kind != "deleted" {
		return model.Finding{}, false
	}
	if !matchAll(item.require, line.Text) {
		return model.Finding{}, false
	}
	reasons := []string{"a protected guard-like line was deleted", "deleted content matched the configured guard pattern"}
	return typedFinding(item, path, line, reasons...), true
}

func evaluateForbiddenEnv(item compiled, path string, line model.DiffLine) (model.Finding, bool) {
	if line.Kind != "added" {
		return model.Finding{}, false
	}
	match := envLine.FindStringSubmatch(strings.TrimSpace(line.Text))
	if len(match) < 3 {
		return model.Finding{}, false
	}
	key := strings.TrimSpace(match[1])
	value := trimConfigValue(match[2])
	if !matchAny(item.keys, key) && !matchAny(item.keys, strings.ToLower(key)) {
		return model.Finding{}, false
	}
	if len(item.rule.ForbiddenValues) != 0 && !matchesValueList(item.rule.ForbiddenValues, value) {
		return model.Finding{}, false
	}
	reasons := []string{"environment key matched a forbidden policy key", "added env assignment: " + key + "=" + value}
	return typedFinding(item, path, line, reasons...), true
}

func evaluateRequiredMiddleware(item compiled, path string, hunkText string, line model.DiffLine) (model.Finding, bool) {
	if line.Kind != "added" {
		return model.Finding{}, false
	}
	if len(item.require) != 0 && !matchAny(item.require, line.Text) {
		return model.Finding{}, false
	}
	if matchAny(item.middleware, hunkText) {
		return model.Finding{}, false
	}
	reasons := []string{"route or handler trigger matched the typed rule", "required middleware was absent from the same diff hunk"}
	return typedFinding(item, path, line, reasons...), true
}

func evaluateRequiredCall(item compiled, path string, hunkText string, line model.DiffLine) (model.Finding, bool) {
	if line.Kind != "added" {
		return model.Finding{}, false
	}
	if len(item.require) != 0 && !matchAny(item.require, line.Text) {
		return model.Finding{}, false
	}
	if matchAny(item.calls, hunkText) {
		return model.Finding{}, false
	}
	reasons := []string{"trigger line matched the typed required-call rule", "required call was absent from the same diff hunk"}
	return typedFinding(item, path, line, reasons...), true
}

func evaluateConfigKeyConstraint(item compiled, path string, line model.DiffLine) (model.Finding, bool) {
	if line.Kind != "added" {
		return model.Finding{}, false
	}
	match := configLine.FindStringSubmatch(strings.TrimSpace(line.Text))
	if len(match) < 3 {
		return model.Finding{}, false
	}
	key := strings.TrimSpace(match[1])
	if item.rule.ConstraintKey != "" && !strings.EqualFold(key, item.rule.ConstraintKey) {
		return model.Finding{}, false
	}
	if item.rule.ConstraintKey == "" && len(item.keys) != 0 && !matchAny(item.keys, key) {
		return model.Finding{}, false
	}
	value := trimConfigValue(match[2])
	if len(item.rule.AllowedValues) != 0 && !matchesValueList(item.rule.AllowedValues, value) {
		return typedFinding(item, path, line, "config key matched the typed constraint", "value was outside the allowed set: "+value), true
	}
	if len(item.rule.ForbiddenValues) != 0 && matchesValueList(item.rule.ForbiddenValues, value) {
		return typedFinding(item, path, line, "config key matched the typed constraint", "value matched a forbidden entry: "+value), true
	}
	if item.constraint != nil && !item.constraint.MatchString(value) {
		return typedFinding(item, path, line, "config key matched the typed constraint", "value failed the required pattern: "+value), true
	}
	if item.rule.MinValue != 0 || item.rule.MaxValue != 0 {
		numeric, err := strconv.ParseFloat(strings.Trim(value, `"'`), 64)
		if err != nil {
			return typedFinding(item, path, line, "config key matched the typed constraint", "value was not numeric: "+value), true
		}
		if item.rule.MinValue != 0 && numeric < item.rule.MinValue {
			return typedFinding(item, path, line, "config key matched the typed constraint", fmt.Sprintf("value %.2f is below the minimum %.2f", numeric, item.rule.MinValue)), true
		}
		if item.rule.MaxValue != 0 && numeric > item.rule.MaxValue {
			return typedFinding(item, path, line, "config key matched the typed constraint", fmt.Sprintf("value %.2f exceeds the maximum %.2f", numeric, item.rule.MaxValue)), true
		}
	}
	return model.Finding{}, false
}

func evaluatePathBoundary(item compiled, path string, line model.DiffLine) (model.Finding, bool) {
	module := importTarget(line.Text)
	if module == "" {
		return model.Finding{}, false
	}
	if len(item.rule.ForbiddenPaths) != 0 && !match.Any(item.rule.ForbiddenPaths, module) && !matchAny(item.imports, module) && !matchAny(item.imports, line.Text) {
		return model.Finding{}, false
	}
	if len(item.imports) != 0 && !matchAny(item.imports, module) && !matchAny(item.imports, line.Text) {
		return model.Finding{}, false
	}
	reasons := []string{"source path matched a typed path boundary", "forbidden target import: " + module}
	return typedFinding(item, path, line, reasons...), true
}

func typedFinding(item compiled, path string, line model.DiffLine, reasons ...string) model.Finding {
	snippet := strings.TrimSpace(line.Text)
	return model.Finding{
		ID:          identity(item.rule.ID, path, number(line), line.Kind, snippet),
		RuleID:      item.rule.ID,
		Title:       item.rule.Title,
		Category:    item.rule.Category,
		Severity:    item.rule.Severity,
		Confidence:  item.rule.Confidence,
		Path:        path,
		Line:        number(line),
		Side:        line.Kind,
		Message:     item.rule.Message,
		Snippet:     snippet,
		Remediation: item.rule.Remediation,
		Reasons:     reasons,
		Standards:   append([]string{}, item.rule.Standards...),
	}
}

func matchAll(items []*regexp.Regexp, value string) bool {
	for _, item := range items {
		if !item.MatchString(value) {
			return false
		}
	}
	return true
}

func matchesValueList(values []string, value string) bool {
	for _, item := range values {
		if strings.EqualFold(strings.TrimSpace(item), value) {
			return true
		}
	}
	return false
}

func trimConfigValue(value string) string {
	trimmed := strings.TrimSpace(value)
	trimmed = strings.TrimSuffix(trimmed, ",")
	trimmed = strings.TrimSpace(trimmed)
	trimmed = strings.Trim(trimmed, `"'`)
	return trimmed
}

func importTarget(value string) string {
	match := importLine.FindStringSubmatch(value)
	for index := 2; index < len(match); index++ {
		if strings.TrimSpace(match[index]) != "" {
			return strings.TrimSpace(match[index])
		}
	}
	return ""
}
