package tui

import (
	"github.com/Pimatis/mavetis/src/baseline"
	"github.com/Pimatis/mavetis/src/config"
	"github.com/Pimatis/mavetis/src/diff"
	"github.com/Pimatis/mavetis/src/model"
	"github.com/Pimatis/mavetis/src/risk"
	"github.com/Pimatis/mavetis/src/rule"
	"github.com/Pimatis/mavetis/src/scan"
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
