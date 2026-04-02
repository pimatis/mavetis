package rule

import "github.com/Pimatis/mavetis/src/model"

func Builtins(config model.Config) []model.Rule {
	rules := make([]model.Rule, 0)
	rules = append(rules, secrets(config)...)
	rules = append(rules, authn()...)
	rules = append(rules, session()...)
	rules = append(rules, authorize()...)
	rules = append(rules, oauth()...)
	rules = append(rules, token()...)
	rules = append(rules, crypto()...)
	rules = append(rules, inject()...)
	rules = append(rules, template()...)
	rules = append(rules, supply()...)
	return enrichControls(rules)
}

func standard(values ...string) []string {
	return values
}
