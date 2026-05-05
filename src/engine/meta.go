package engine

import (
	"github.com/Pimatis/mavetis/src/model"
	"github.com/Pimatis/mavetis/src/rule"
)

func MatrixInfos(rules []model.Rule) []model.RuleInfo {
	items := infos(rules)
	items = append(items, syntheticInfosForProfile("")...)
	return items
}

func infos(rules []model.Rule) []model.RuleInfo {
	items := make([]model.RuleInfo, 0, len(rules))
	for _, rule := range rules {
		items = append(items, model.RuleInfo{ID: rule.ID, Title: rule.Title, Category: rule.Category, Severity: rule.Severity, Standards: rule.Standards})
	}
	return items
}

func snapshotInfos(snapshots []model.Snapshot) []model.RuleInfo {
	items := make([]model.RuleInfo, 0, len(snapshots))
	for _, snapshot := range snapshots {
		items = append(items, model.RuleInfo{ID: snapshot.ID, Title: "Repository security snapshot regressed", Category: snapshot.Category, Severity: snapshot.Severity, Standards: snapshot.Standards})
	}
	return items
}

func syntheticInfos() []model.RuleInfo {
	return rule.SyntheticInfos()
}

func appendFindings(target *[]model.Finding, seen map[string]struct{}, additions []model.Finding, config model.Config, zoneCache map[string]zoneMatch) {
	for _, finding := range additions {
		if !allowSynthetic(config.Profile, finding) {
			continue
		}
		if finding.Path != "" && finding.Path != "<branch>" {
			zone := resolveZone(finding.Path, config, zoneCache)
			finding = applyPolicy(finding, zone, config.FailOn)
		}
		if model.SeverityRank(finding.Severity) < model.SeverityRank(config.Severity) {
			continue
		}
		if _, ok := seen[finding.ID]; ok {
			continue
		}
		seen[finding.ID] = struct{}{}
		*target = append(*target, finding)
	}
}
