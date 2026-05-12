package tui

import (
	"github.com/Pimatis/mavetis/src/baseline"
	"github.com/Pimatis/mavetis/src/config"
	"github.com/Pimatis/mavetis/src/diff"
	"github.com/Pimatis/mavetis/src/model"
	"github.com/Pimatis/mavetis/src/output"
	"github.com/Pimatis/mavetis/src/risk"
	"github.com/Pimatis/mavetis/src/rule"
	"github.com/Pimatis/mavetis/src/scan"
	"github.com/Pimatis/mavetis/src/secret"
	"github.com/Pimatis/mavetis/src/wizard"
)

func diffParse(raw string, meta model.DiffMeta) (model.Diff, error) {
	return diff.Parse(raw, meta)
}

func loadConfig(path string) (model.Config, error) {
	return config.Load(path)
}

func allRulesFor(config model.Config) []model.Rule {
	return rule.Builtins(config)
}

func baselineLoad(path string) (baseline.File, error) {
	return baseline.Load(path)
}

func baselineFilter(report model.Report, file baseline.File) model.Report {
	return baseline.Filter(report, file)
}

func riskCalculate(summary model.Summary) risk.Score {
	return risk.Calculate(summary)
}

func scanRoot() (string, error) {
	return scan.Root()
}

func loadAllFiles(root string) ([]scan.ScannedFile, error) {
	return scan.LoadAllFiles(root)
}

func fromFiles(files []scan.ScannedFile) model.Diff {
	return scan.FromFiles(files)
}

func secretScan(root string, cfg model.Config) (model.Report, error) {
	return secret.Scan(root, cfg, secret.Options{})
}

func baselineCreate(path string, report model.Report) error {
	return baseline.Create(path, report)
}

func ruleExplain(id string, rules []model.Rule) (model.RuleExplanation, bool) {
	return rule.Explain(id, rules)
}

func ruleList(cfg model.Config) []model.RuleInfo {
	items := rule.Builtins(cfg)
	infos := make([]model.RuleInfo, 0, len(items))
	for _, item := range items {
		infos = append(infos, model.RuleInfo{
			ID:       item.ID,
			Title:    item.Title,
			Category: item.Category,
			Severity: item.Severity,
			Standards: append([]string{}, item.Standards...),
		})
	}
	return infos
}

func wizardDetect(root string) wizard.Project {
	return wizard.Detect(root)
}

func wizardGenerate(template wizard.ConfigTemplate) string {
	return wizard.Generate(template)
}

func appendGitignore(root string, entry string) error {
	return wizard.AppendGitignore(root, entry)
}

func outputExplain(data model.RuleExplanation) string {
	return output.RuleExplanation(data)
}
