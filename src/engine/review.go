package engine

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/Pimatis/mavetis/src/analyze"
	"github.com/Pimatis/mavetis/src/match"
	"github.com/Pimatis/mavetis/src/model"
)

type compiled struct {
	rule       model.Rule
	require    []*regexp.Regexp
	any        []*regexp.Regexp
	near       []*regexp.Regexp
	absent     []*regexp.Regexp
	imports    []*regexp.Regexp
	calls      []*regexp.Regexp
	middleware []*regexp.Regexp
	keys       []*regexp.Regexp
	constraint *regexp.Regexp
}

type allowance struct {
	values  []string
	regexes []*regexp.Regexp
}

func Review(diff model.Diff, config model.Config, rules []model.Rule) (model.Report, error) {
	report := model.Report{Meta: diff.Meta}
	report.Policy = policyReport(config)
	report.Summary.Files = len(diff.Files)
	report.Rules = infos(rules)
	report.Rules = append(report.Rules, snapshotInfos(config.Snapshots)...)
	compiledRules, err := compile(rules)
	if err != nil {
		return report, err
	}
	index := buildIndex(compiledRules)
	allowlist, err := compileAllow(config)
	if err != nil {
		return report, err
	}
	seen := map[string]struct{}{}
	zoneCache := map[string]zoneMatch{}
	findings := make([]model.Finding, 0)
	for _, file := range diff.Files {
		if analyze.ReviewArtifact(file.Path) {
			continue
		}
		if match.Any(config.Ignore, file.Path) {
			continue
		}
		if match.Any(config.Allow.Paths, file.Path) {
			continue
		}
		zone := resolveZone(file.Path, config, zoneCache)
		for _, hunk := range file.Hunks {
			hunkText := join(hunk)
			for _, line := range hunk.Lines {
				for _, item := range index.selectRules(line.Kind) {
					finding, ok := evaluate(item, file.Path, hunkText, line, allowlist)
					if !ok {
						continue
					}
					finding = applyPolicy(finding, zone, config.FailOn)
					if model.SeverityRank(finding.Severity) < model.SeverityRank(config.Severity) {
						continue
					}
					if _, ok := seen[finding.ID]; ok {
						continue
					}
					seen[finding.ID] = struct{}{}
					findings = append(findings, finding)
				}
			}
		}
	}
	fileMode := diff.Meta.Mode == "file"
	appendFindings(&findings, seen, semanticFindings(diff), config, zoneCache)
	appendFindings(&findings, seen, goSemanticFindings(diff), config, zoneCache)
	appendFindings(&findings, seen, nonceFindings(diff), config, zoneCache)
	appendFindings(&findings, seen, signatureFindings(diff), config, zoneCache)
	if !fileMode {
		appendFindings(&findings, seen, manifestFindings(diff), config, zoneCache)
		appendFindings(&findings, seen, supplyTrustFindings(diff, config), config, zoneCache)
		appendFindings(&findings, seen, downgradeFindings(diff), config, zoneCache)
		appendFindings(&findings, seen, intentFindings(diff), config, zoneCache)
		appendFindings(&findings, seen, snapshotFindings(diff, config.Snapshots), config, zoneCache)
		appendFindings(&findings, seen, crossFindings(diff), config, zoneCache)
	}
	sortFindings(findings)
	report.Findings = findings
	report.Rules = append(report.Rules, syntheticInfosForProfile(config.Profile)...)
	for _, finding := range findings {
		report.Summary.Add(finding)
	}
	return report, nil
}

func compile(rules []model.Rule) ([]compiled, error) {
	items := make([]compiled, 0, len(rules))
	for _, rule := range rules {
		item := compiled{rule: rule}
		for _, pattern := range rule.Require {
			re, err := regexp.Compile(pattern)
			if err != nil {
				return nil, fmt.Errorf("compile rule %s: %w", rule.ID, err)
			}
			item.require = append(item.require, re)
		}
		for _, pattern := range rule.Any {
			re, err := regexp.Compile(pattern)
			if err != nil {
				return nil, fmt.Errorf("compile rule %s: %w", rule.ID, err)
			}
			item.any = append(item.any, re)
		}
		for _, pattern := range rule.Near {
			re, err := regexp.Compile(pattern)
			if err != nil {
				return nil, fmt.Errorf("compile rule %s: %w", rule.ID, err)
			}
			item.near = append(item.near, re)
		}
		for _, pattern := range rule.Absent {
			re, err := regexp.Compile(pattern)
			if err != nil {
				return nil, fmt.Errorf("compile rule %s: %w", rule.ID, err)
			}
			item.absent = append(item.absent, re)
		}
		for _, pattern := range rule.Imports {
			re, err := regexp.Compile(pattern)
			if err != nil {
				return nil, fmt.Errorf("compile rule %s: %w", rule.ID, err)
			}
			item.imports = append(item.imports, re)
		}
		for _, pattern := range rule.Calls {
			re, err := regexp.Compile(pattern)
			if err != nil {
				return nil, fmt.Errorf("compile rule %s: %w", rule.ID, err)
			}
			item.calls = append(item.calls, re)
		}
		for _, pattern := range rule.Middleware {
			re, err := regexp.Compile(pattern)
			if err != nil {
				return nil, fmt.Errorf("compile rule %s: %w", rule.ID, err)
			}
			item.middleware = append(item.middleware, re)
		}
		for _, pattern := range rule.Keys {
			re, err := regexp.Compile(pattern)
			if err != nil {
				return nil, fmt.Errorf("compile rule %s: %w", rule.ID, err)
			}
			item.keys = append(item.keys, re)
		}
		if rule.ConstraintPattern != "" {
			re, err := regexp.Compile(rule.ConstraintPattern)
			if err != nil {
				return nil, fmt.Errorf("compile rule %s: %w", rule.ID, err)
			}
			item.constraint = re
		}
		items = append(items, item)
	}
	return items, nil
}

func evaluate(item compiled, path string, hunkText string, line model.DiffLine, allowlist allowance) (model.Finding, bool) {
	finding := model.Finding{}
	if !target(item.rule.Target, line.Kind) {
		return finding, false
	}
	if len(item.rule.Paths) != 0 && !match.Any(item.rule.Paths, path) {
		return finding, false
	}
	if len(item.rule.FromPaths) != 0 && !match.Any(item.rule.FromPaths, path) {
		return finding, false
	}
	if match.Any(item.rule.Ignore, path) {
		return finding, false
	}
	if allow(allowlist, line.Text) {
		return finding, false
	}
	if item.rule.Type != "" {
		return evaluateTyped(item, path, hunkText, line)
	}
	for _, re := range item.require {
		if !re.MatchString(line.Text) {
			return finding, false
		}
	}
	if len(item.any) != 0 && !matchAny(item.any, line.Text) {
		return finding, false
	}
	for _, re := range item.near {
		if !re.MatchString(hunkText) {
			return finding, false
		}
	}
	for _, re := range item.absent {
		if re.MatchString(hunkText) {
			return finding, false
		}
	}
	if item.rule.Entropy > 0 {
		if entropy(line.Text) < item.rule.Entropy {
			return finding, false
		}
	}
	finding.RuleID = item.rule.ID
	finding.Title = item.rule.Title
	finding.Category = item.rule.Category
	finding.Severity = item.rule.Severity
	finding.Confidence = confidence(item, line.Text, len(item.near) != 0)
	finding.Path = path
	finding.Side = line.Kind
	finding.Message = item.rule.Message
	finding.Remediation = item.rule.Remediation
	finding.Standards = append([]string{}, item.rule.Standards...)
	finding.Reasons = explain(item, path, line, hunkText)
	finding.Line = number(line)
	finding.Snippet = line.Text
	if item.rule.Mask {
		finding.Snippet = mask(line.Text)
	}
	finding.ID = identity(item.rule.ID, path, finding.Line, finding.Side, finding.Snippet)
	return finding, true
}

func join(hunk model.DiffHunk) string {
	parts := make([]string, 0, len(hunk.Lines))
	for _, line := range hunk.Lines {
		parts = append(parts, line.Text)
	}
	return strings.Join(parts, "\n")
}

func matchAny(matchers []*regexp.Regexp, value string) bool {
	for _, re := range matchers {
		if re.MatchString(value) {
			return true
		}
	}
	return false
}

func target(expected string, actual string) bool {
	if expected == "any" {
		return true
	}
	return expected == actual
}

func compileAllow(config model.Config) (allowance, error) {
	result := allowance{values: append([]string{}, config.Allow.Values...)}
	for _, pattern := range config.Allow.Regexes {
		re, err := regexp.Compile(pattern)
		if err != nil {
			return result, fmt.Errorf("compile allow regex %q: %w", pattern, err)
		}
		result.regexes = append(result.regexes, re)
	}
	return result, nil
}

func allow(allowlist allowance, value string) bool {
	for _, item := range allowlist.values {
		if item == "" {
			continue
		}
		if strings.Contains(value, item) {
			return true
		}
	}
	for _, re := range allowlist.regexes {
		if re.MatchString(value) {
			return true
		}
	}
	return false
}

func number(line model.DiffLine) int {
	if line.Kind == "deleted" {
		return line.OldNumber
	}
	return line.NewNumber
}

func identity(ruleID string, path string, line int, side string, snippet string) string {
	sum := sha256.Sum256([]byte(ruleID + "|" + path + "|" + fmt.Sprintf("%d", line) + "|" + side + "|" + snippet))
	return hex.EncodeToString(sum[:8])
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
