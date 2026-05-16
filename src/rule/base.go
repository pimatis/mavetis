package rule

import "github.com/Pimatis/mavetis/src/model"

func Builtins(config model.Config) []model.Rule {
	rules := make([]model.Rule, 0)
	rules = append(rules, secrets(config)...)
	rules = append(rules, authn()...)
	rules = append(rules, regress()...)
	rules = append(rules, session()...)
	rules = append(rules, authorize()...)
	rules = append(rules, oauth()...)
	rules = append(rules, webhook()...)
	rules = append(rules, token()...)
	rules = append(rules, crypto()...)
	rules = append(rules, inject()...)
	rules = append(rules, ai()...)
	rules = append(rules, template()...)
	rules = append(rules, boundary()...)
	rules = append(rules, cloud()...)
	rules = append(rules, supply()...)
	rules = append(rules, drift()...)
	rules = append(rules, observe()...)
	rules = append(rules, logic()...)
	rules = append(rules, websocket()...)
	rules = append(rules, race()...)
	rules = append(rules, graphql()...)
	rules = append(rules, gospecific()...)
	rules = append(rules, client()...)
	rules = append(rules, file()...)
	rules = append(rules, grpc()...)
	return enrichControls(rules)
}

func standard(values ...string) []string {
	return values
}
