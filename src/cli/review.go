package cli

import (
	"encoding/json"

	"github.com/Pimatis/mavetis/src/baseline"
	"github.com/Pimatis/mavetis/src/cache"
	"github.com/Pimatis/mavetis/src/config"
	"github.com/Pimatis/mavetis/src/diff"
	"github.com/Pimatis/mavetis/src/engine"
	"github.com/Pimatis/mavetis/src/git"
	"github.com/Pimatis/mavetis/src/model"
	"github.com/Pimatis/mavetis/src/risk"
	"github.com/Pimatis/mavetis/src/scan"
)

func runReview(arguments []string, ci bool) int {
	spec, cfg, rules, err := prepareReview(arguments, ci)
	if err != nil {
		return fail(err)
	}
	report, err := buildReport(spec, cfg, rules)
	if err != nil {
		return fail(err)
	}
	report = applyBaseline(report, cfg, spec)
	report.Score = riskScore(report.Summary)
	if err := render(report, cfg.Output, spec.Explain); err != nil {
		return fail(err)
	}
	if blocked(report, cfg.FailOn) {
		return 1
	}
	return 0
}

func prepareReview(arguments []string, ci bool) (model.Review, model.Config, []model.Rule, error) {
	spec, err := parseReview(arguments, ci)
	if err != nil {
		return model.Review{}, model.Config{}, nil, err
	}
	cfg, err := config.Load(spec.ConfigPath)
	if err != nil {
		return model.Review{}, model.Config{}, nil, err
	}
	merge(&cfg, spec)
	rules, err := loadAllRules(cfg, spec.RulesPath)
	if err != nil {
		return model.Review{}, model.Config{}, nil, err
	}
	return spec, cfg, rules, nil
}

func buildReport(spec model.Review, cfg model.Config, rules []model.Rule) (model.Report, error) {
	if spec.All {
		return buildAllReport(spec, cfg, rules)
	}
	if len(spec.Files) != 0 {
		return buildFileReport(spec, cfg, rules)
	}
	raw, meta, err := git.Review(spec)
	if err != nil {
		return model.Report{}, err
	}
	parsed, err := diff.Parse(raw, meta)
	if err != nil {
		return model.Report{}, err
	}
	parsed = diff.Filter(parsed, spec.Path)
	suggestions := make([]model.Suggestion, 0)
	if spec.WithContext {
		var contextErr error
		parsed, suggestions, contextErr = withChangedContext(parsed)
		if contextErr != nil {
			return model.Report{}, contextErr
		}
	}
	report, err := engine.Review(parsed, cfg, rules)
	if err != nil {
		return model.Report{}, err
	}
	if spec.Path != "" {
		report.Meta.Mode = meta.Mode + ":" + spec.Path
	}
	if len(suggestions) != 0 {
		report.Suggestions = suggestions
	}
	return report, nil
}

func reviewScannedFiles(root string, files []scan.ScannedFile, spec model.Review, cfg model.Config, rules []model.Rule) (model.Report, error) {
	if spec.NoCache {
		return engine.Review(scan.FromFiles(files), cfg, rules)
	}
	cacheKey, err := reviewCacheKey(cfg, rules)
	if err != nil {
		return model.Report{}, err
	}
	cachePath, cacheData, err := cache.Load(root, "review", spec.CachePath, cacheKey)
	if err != nil {
		return model.Report{}, err
	}
	cacheFiles := make([]cache.File, 0, len(files))
	for _, file := range files {
		cacheFiles = append(cacheFiles, cache.File{Path: file.Path, Size: file.Size, ModTime: file.ModTime})
	}
	cache.Prune(cacheData, cacheFiles)
	report := model.Report{Meta: model.DiffMeta{Mode: "file"}}
	report.Policy = &model.Policy{Profile: cfg.Profile, FailOn: cfg.FailOn}
	seen := map[string]struct{}{}
	findings := make([]model.Finding, 0)
	for _, file := range files {
		cacheFile := cache.File{Path: file.Path, Size: file.Size, ModTime: file.ModTime}
		fileFindings, ok := cache.Findings(cacheData, cacheFile)
		if !ok {
			fileReport, err := engine.Review(scan.FromFiles([]scan.ScannedFile{file}), cfg, rules)
			if err != nil {
				return report, err
			}
			fileFindings = fileReport.Findings
			cache.Put(cacheData, cacheFile, fileFindings)
			report.Rules = fileReport.Rules
			if report.Policy == nil {
				report.Policy = fileReport.Policy
			}
		}
		for _, finding := range fileFindings {
			if _, exists := seen[finding.ID]; exists {
				continue
			}
			seen[finding.ID] = struct{}{}
			findings = append(findings, finding)
		}
	}
	if len(report.Rules) == 0 {
		emptyReport, err := engine.Review(scan.FromFiles(nil), cfg, rules)
		if err != nil {
			return report, err
		}
		report.Rules = emptyReport.Rules
		report.Policy = emptyReport.Policy
	}
	sortFindings(findings)
	report.Findings = findings
	report.Summary.Files = len(files)
	for _, finding := range findings {
		report.Summary.Add(finding)
	}
	if err := cache.Save(cachePath, cacheData); err != nil {
		return report, err
	}
	return report, nil
}

func reviewCacheKey(cfg model.Config, rules []model.Rule) (string, error) {
	payload := struct {
		Severity string
		FailOn   string
		Profile  string
		Allow    model.Allow
		Zones    model.Zones
		Rules    []model.Rule
	}{
		Severity: cfg.Severity,
		FailOn:   cfg.FailOn,
		Profile:  cfg.Profile,
		Allow:    cfg.Allow,
		Zones:    cfg.Zones,
		Rules:    rules,
	}
	content, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	return cache.Key(string(content)), nil
}

func applyBaseline(report model.Report, cfg model.Config, spec model.Review) model.Report {
	path := spec.BaselinePath
	if path == "" {
		path = cfg.Baseline.Path
	}
	if path == "" {
		return report
	}
	bl, err := baseline.Load(path)
	if err != nil {
		return report
	}
	return baseline.Filter(report, bl)
}

func riskScore(summary model.Summary) *model.Score {
	score := risk.Calculate(summary)
	return &model.Score{Value: score.Value, Rating: score.Rating}
}
