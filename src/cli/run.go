package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/Pimatis/mavetis/src/baseline"
	"github.com/Pimatis/mavetis/src/cache"
	"github.com/Pimatis/mavetis/src/config"
	"github.com/Pimatis/mavetis/src/diff"
	"github.com/Pimatis/mavetis/src/engine"
	"github.com/Pimatis/mavetis/src/git"
	"github.com/Pimatis/mavetis/src/hook"
	"github.com/Pimatis/mavetis/src/match"
	"github.com/Pimatis/mavetis/src/model"
	"github.com/Pimatis/mavetis/src/output"
	"github.com/Pimatis/mavetis/src/resolve"
	"github.com/Pimatis/mavetis/src/risk"
	"github.com/Pimatis/mavetis/src/scan"
	"github.com/Pimatis/mavetis/src/tui"
)

func Execute(arguments []string) int {
	if len(arguments) == 0 {
		if err := tui.Run(); err != nil {
			if errors.Is(err, tui.ErrNoTerminal) {
				usage()
				return 0
			}
			return fail(err)
		}
		return 0
	}
	command := arguments[0]
	if command == "help" || command == "-h" || command == "--help" {
		usage()
		return 0
	}
	if command == "review" {
		return runReview(arguments[1:], false)
	}
	if command == "ci" {
		return runReview(arguments[1:], true)
	}
	if command == "hooks" {
		return runHooks(arguments[1:])
	}
	if command == "rules" {
		return runRules(arguments[1:])
	}
	if command == "explain" {
		return runExplain(arguments[1:])
	}
	if command == "update" {
		return runUpdate(arguments[1:])
	}
	if command == "shell" {
		return runShell(arguments[1:])
	}
	if command == "init" {
		return runInit(arguments[1:])
	}
	if command == "baseline" {
		return runBaseline(arguments[1:])
	}
	if command == "secrets" {
		return runSecrets(arguments[1:])
	}
	if command == "version" || command == "-v" || command == "--version" {
		fmt.Printf("%s %s\n", model.Name, model.Version)
		return 0
	}
	usage()
	return 1
}

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

func buildFileReport(spec model.Review, cfg model.Config, rules []model.Rule) (model.Report, error) {
	root, err := scan.Root()
	if err != nil {
		return model.Report{}, err
	}
	files, err := scan.LoadFiles(root, spec.Files)
	if err != nil {
		return model.Report{}, err
	}
	files = filterScannedFiles(files, spec.Path)
	reviewFiles := append([]scan.ScannedFile{}, files...)
	suggestions := make([]model.Suggestion, 0)
	if spec.WithSuggested {
		discovered, additions, discoverErr := resolve.Discover(root, files, resolve.DefaultLimits())
		if discoverErr != nil {
			return model.Report{}, discoverErr
		}
		reviewFiles = appendUniqueFiles(reviewFiles, discovered)
		suggestions = markReviewedSuggestions(additions)
	}
	if !spec.WithSuggested {
		additions, suggestErr := resolve.Suggest(root, files, resolve.DefaultLimits())
		if suggestErr != nil {
			return model.Report{}, suggestErr
		}
		suggestions = additions
	}
	report, err := reviewScannedFiles(root, reviewFiles, spec, cfg, rules)
	if err != nil {
		return model.Report{}, err
	}
	report.Meta.Mode = "file"
	if spec.Path != "" {
		report.Meta.Mode = report.Meta.Mode + ":" + spec.Path
	}
	report.Suggestions = suggestions
	if len(suggestions) != 0 && !spec.WithSuggested {
		report.SuggestedCommand = suggestedCommand(spec)
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

func filterScannedFiles(files []scan.ScannedFile, pattern string) []scan.ScannedFile {
	if pattern == "" {
		return files
	}
	filtered := make([]scan.ScannedFile, 0, len(files))
	for _, file := range files {
		if !match.Glob(pattern, file.Path) {
			continue
		}
		filtered = append(filtered, file)
	}
	return filtered
}

func appendUniqueFiles(current []scan.ScannedFile, additions []scan.ScannedFile) []scan.ScannedFile {
	seen := map[string]struct{}{}
	for _, file := range current {
		seen[file.Path] = struct{}{}
	}
	for _, file := range additions {
		if _, ok := seen[file.Path]; ok {
			continue
		}
		seen[file.Path] = struct{}{}
		current = append(current, file)
	}
	return current
}

func markReviewedSuggestions(suggestions []model.Suggestion) []model.Suggestion {
	for index := range suggestions {
		suggestions[index].Reviewed = true
	}
	return suggestions
}

func withChangedContext(parsed model.Diff) (model.Diff, []model.Suggestion, error) {
	root, err := scan.Root()
	if err != nil {
		return parsed, nil, err
	}
	seeds, err := scan.LoadExistingFiles(root, changedContextSeedPaths(parsed))
	if err != nil {
		return parsed, nil, err
	}
	discovered, additions, err := resolve.Discover(root, seeds, resolve.DefaultLimits())
	if err != nil {
		return parsed, nil, err
	}
	if len(discovered) == 0 {
		return parsed, nil, nil
	}
	return appendContextFiles(parsed, discovered), markReviewedSuggestions(additions), nil
}

func changedContextSeedPaths(parsed model.Diff) []string {
	seen := map[string]struct{}{}
	paths := make([]string, 0, len(parsed.Files))
	for _, file := range parsed.Files {
		if file.Path == "" {
			continue
		}
		if file.Change == "deleted" {
			continue
		}
		if _, ok := seen[file.Path]; ok {
			continue
		}
		seen[file.Path] = struct{}{}
		paths = append(paths, file.Path)
	}
	return paths
}

func appendContextFiles(parsed model.Diff, files []scan.ScannedFile) model.Diff {
	seen := map[string]struct{}{}
	for _, file := range parsed.Files {
		if file.Path == "" {
			continue
		}
		seen[file.Path] = struct{}{}
	}
	contextDiff := scan.FromFiles(files)
	for _, file := range contextDiff.Files {
		if _, ok := seen[file.Path]; ok {
			continue
		}
		file.Change = "context"
		for hunkIndex := range file.Hunks {
			file.Hunks[hunkIndex].Header = "@@ changed context @@"
		}
		seen[file.Path] = struct{}{}
		parsed.Files = append(parsed.Files, file)
	}
	if parsed.Meta.Mode != "" {
		parsed.Meta.Mode = parsed.Meta.Mode + "+context"
	}
	return parsed
}

func suggestedCommand(spec model.Review) string {
	parts := []string{"mavetis", "review"}
	for _, file := range spec.Files {
		parts = append(parts, shellPart(file))
	}
	if spec.Path != "" {
		parts = append(parts, "--path", shellPart(spec.Path))
	}
	if spec.Profile != "" {
		parts = append(parts, "--profile", shellPart(spec.Profile))
	}
	if spec.Severity != "" {
		parts = append(parts, "--severity", shellPart(spec.Severity))
	}
	if spec.FailOn != "" {
		parts = append(parts, "--fail-on", shellPart(spec.FailOn))
	}
	if spec.ConfigPath != "" {
		parts = append(parts, "--config", shellPart(spec.ConfigPath))
	}
	if spec.RulesPath != "" {
		parts = append(parts, "--rules", shellPart(spec.RulesPath))
	}
	if spec.Format != "" {
		parts = append(parts, "--format", shellPart(spec.Format))
	}
	if spec.Explain {
		parts = append(parts, "--explain")
	}
	parts = append(parts, "--with-suggested")
	return strings.Join(parts, " ")
}

func shellPart(value string) string {
	if value == "" {
		return `""`
	}
	if strings.ContainsAny(value, " \t\n\r'\"\\$&;|<>*?()[]{}") {
		return strconv.Quote(value)
	}
	return value
}

func runHooks(arguments []string) int {
	if len(arguments) == 0 {
		return fail(errors.New("hooks command requires install or uninstall"))
	}
	root, err := git.Root()
	if err != nil {
		return fail(err)
	}
	if arguments[0] == "install" {
		if err := hook.Install(root); err != nil {
			return fail(err)
		}
		fmt.Println("hooks installed")
		return 0
	}
	if arguments[0] == "uninstall" {
		if err := hook.Uninstall(root); err != nil {
			return fail(err)
		}
		fmt.Println("hooks removed")
		return 0
	}
	return fail(errors.New("hooks command requires install or uninstall"))
}

func merge(config *model.Config, spec model.Review) {
	if spec.Severity != "" {
		config.Severity = spec.Severity
	}
	if spec.FailOn != "" {
		config.FailOn = spec.FailOn
	}
	if spec.Format != "" {
		config.Output = spec.Format
	}
	if spec.Profile != "" {
		config.Profile = spec.Profile
	}
}

func render(report model.Report, format string, explain bool) error {
	if format == "json" {
		body, err := output.JSON(report)
		if err != nil {
			return err
		}
		fmt.Println(body)
		return nil
	}
	if format == "sarif" {
		body, err := output.SARIF(report)
		if err != nil {
			return err
		}
		fmt.Println(body)
		return nil
	}
	fmt.Print(output.TextExplain(report, explain))
	return nil
}

func blocked(report model.Report, threshold string) bool {
	for _, finding := range report.Findings {
		effectiveThreshold := threshold
		if finding.EffectiveFailOn != "" {
			effectiveThreshold = finding.EffectiveFailOn
		}
		if model.SeverityRank(finding.Severity) >= model.SeverityRank(effectiveThreshold) {
			return true
		}
	}
	return false
}

func fail(err error) int {
	fmt.Fprintln(os.Stderr, err.Error())
	return 2
}

func riskScore(summary model.Summary) *model.Score {
	score := risk.Calculate(summary)
	return &model.Score{Value: score.Value, Rating: score.Rating}
}
