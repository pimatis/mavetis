package cli

import (
	"fmt"

	"github.com/Pimatis/mavetis/src/config"
	"github.com/Pimatis/mavetis/src/model"
	"github.com/Pimatis/mavetis/src/rule"
)

func loadAllRules(data model.Config, path string) ([]model.Rule, error) {
	rules := rule.Builtins(data)
	custom, err := config.LoadRules(path)
	if err != nil {
		return nil, err
	}
	rules = append(rules, custom...)
	if err := validateRuleIDs(rules); err != nil {
		return nil, err
	}
	return rules, nil
}

func validateRuleIDs(rules []model.Rule) error {
	seen := map[string]struct{}{}
	for _, item := range rules {
		if _, ok := seen[item.ID]; ok {
			return fmt.Errorf("duplicate rule id: %s", item.ID)
		}
		seen[item.ID] = struct{}{}
	}
	return nil
}
