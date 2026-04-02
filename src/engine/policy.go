package engine

import (
	"strings"

	"github.com/Pimatis/mavetis/src/match"
	"github.com/Pimatis/mavetis/src/model"
)

type zonePolicy struct {
	name           string
	paths          []string
	severityOffset int
	failOn         string
}

type zoneMatch struct {
	name           string
	severityOffset int
	failOn         string
}

type profileSpec struct {
	categories map[string]struct{}
	prefixes   []string
}

var profiles = map[string]profileSpec{
	"auth": {
		categories: toSet("auth", "authorization", "session", "token", "oauth", "crypto"),
		prefixes:   []string{"observe.auth.", "downgrade.auth.", "downgrade.cookie.", "downgrade.crypto."},
	},
	"backend": {
		categories: toSet("auth", "authorization", "session", "token", "oauth", "crypto", "injection", "ssrf", "deserialization", "file", "transport", "logging", "error", "privacy", "supply", "config", "template", "cors"),
		prefixes:   []string{"observe.", "config.", "downgrade."},
	},
	"frontend": {
		categories: toSet("auth", "session", "token", "cors", "xss", "logging", "error", "privacy", "config"),
		prefixes:   []string{"observe.", "config.", "downgrade.cookie.", "downgrade.timeout"},
	},
}

func FilterRulesForProfile(rules []model.Rule, profile string) []model.Rule {
	if profile == "" || profile == "fintech" {
		return append([]model.Rule{}, rules...)
	}
	filtered := make([]model.Rule, 0, len(rules))
	for _, item := range rules {
		if !allowProfile(profile, item.ID, item.Category) {
			continue
		}
		filtered = append(filtered, item)
	}
	return filtered
}

func allowSynthetic(profile string, finding model.Finding) bool {
	return allowProfile(profile, finding.RuleID, finding.Category)
}

func syntheticInfosForProfile(profile string) []model.RuleInfo {
	items := syntheticInfos()
	if profile == "" || profile == "fintech" {
		return items
	}
	filtered := make([]model.RuleInfo, 0, len(items))
	for _, item := range items {
		if !allowProfile(profile, item.ID, item.Category) {
			continue
		}
		filtered = append(filtered, item)
	}
	return filtered
}

func policyReport(config model.Config) *model.Policy {
	policies := zonePolicies(config.Zones)
	if config.Profile == "" && len(policies) == 0 {
		return nil
	}
	policy := &model.Policy{Profile: config.Profile, FailOn: config.FailOn}
	for _, item := range policies {
		policy.Zones = append(policy.Zones, model.PolicyZone{Name: item.name, Paths: append([]string{}, item.paths...), SeverityOffset: item.severityOffset, FailOn: item.failOn})
	}
	return policy
}

func resolveZone(path string, config model.Config, cache map[string]zoneMatch) zoneMatch {
	if value, ok := cache[path]; ok {
		return value
	}
	best := zoneMatch{}
	for _, item := range zonePolicies(config.Zones) {
		if len(item.paths) == 0 {
			continue
		}
		if !match.Any(item.paths, path) {
			continue
		}
		if item.severityOffset > best.severityOffset {
			best = zoneMatch{name: item.name, severityOffset: item.severityOffset, failOn: item.failOn}
		}
	}
	cache[path] = best
	return best
}

func applyPolicy(finding model.Finding, zone zoneMatch, defaultFailOn string) model.Finding {
	if zone.name == "" {
		return finding
	}
	finding.Zone = zone.name
	finding.BaseSeverity = finding.Severity
	finding.EffectiveFailOn = stricterFailOn(defaultFailOn, zone.failOn)
	finding.Severity = raiseSeverity(finding.Severity, zone.severityOffset)
	return finding
}

func allowProfile(profile string, ruleID string, category string) bool {
	if profile == "" || profile == "fintech" {
		return true
	}
	spec, ok := profiles[profile]
	if !ok {
		return false
	}
	if _, ok := spec.categories[category]; ok {
		return true
	}
	for _, prefix := range spec.prefixes {
		if strings.HasPrefix(ruleID, prefix) {
			return true
		}
	}
	return false
}

func zonePolicies(zones model.Zones) []zonePolicy {
	items := make([]zonePolicy, 0, 2)
	if len(zones.Restricted) != 0 {
		items = append(items, zonePolicy{name: "restricted", paths: append([]string{}, zones.Restricted...), severityOffset: 1, failOn: "medium"})
	}
	if len(zones.Critical) != 0 {
		items = append(items, zonePolicy{name: "critical", paths: append([]string{}, zones.Critical...), severityOffset: 2, failOn: "low"})
	}
	return items
}

func raiseSeverity(value string, offset int) string {
	rank := model.SeverityRank(value)
	if rank == 0 {
		return value
	}
	rank += offset
	if rank > 4 {
		rank = 4
	}
	return severityName(rank)
}

func severityName(rank int) string {
	if rank == 4 {
		return "critical"
	}
	if rank == 3 {
		return "high"
	}
	if rank == 2 {
		return "medium"
	}
	return "low"
}

func stricterFailOn(current string, zone string) string {
	if current == "" {
		return zone
	}
	if zone == "" {
		return current
	}
	if model.SeverityRank(zone) < model.SeverityRank(current) {
		return zone
	}
	return current
}

func toSet(values ...string) map[string]struct{} {
	set := map[string]struct{}{}
	for _, item := range values {
		set[item] = struct{}{}
	}
	return set
}
