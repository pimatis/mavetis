package output

import (
	"encoding/json"

	"github.com/Pimatis/mavetis/src/model"
)

func JSON(report model.Report) (string, error) {
	document := report
	document.Rules = matchedRules(report)
	buffer, err := json.MarshalIndent(document, "", "  ")
	if err != nil {
		return "", err
	}
	return string(buffer), nil
}

func matchedRules(report model.Report) []model.RuleInfo {
	if len(report.Findings) == 0 {
		return nil
	}
	matched := map[string]struct{}{}
	for _, finding := range report.Findings {
		matched[finding.RuleID] = struct{}{}
	}
	rules := make([]model.RuleInfo, 0, len(matched))
	for _, rule := range report.Rules {
		if _, ok := matched[rule.ID]; !ok {
			continue
		}
		rules = append(rules, rule)
	}
	return rules
}
