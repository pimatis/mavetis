package cli

import (
	"strconv"
	"strings"

	"github.com/Pimatis/mavetis/src/match"
	"github.com/Pimatis/mavetis/src/model"
	"github.com/Pimatis/mavetis/src/resolve"
	"github.com/Pimatis/mavetis/src/scan"
)

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

func buildAllReport(spec model.Review, cfg model.Config, rules []model.Rule) (model.Report, error) {
	root, err := scan.Root()
	if err != nil {
		return model.Report{}, err
	}
	files, err := scan.LoadAllFiles(root)
	if err != nil {
		return model.Report{}, err
	}
	files = filterScannedFiles(files, spec.Path)
	report, err := reviewScannedFiles(root, files, spec, cfg, rules)
	if err != nil {
		return model.Report{}, err
	}
	report.Meta.Mode = "file:all"
	if spec.Path != "" {
		report.Meta.Mode = report.Meta.Mode + ":" + spec.Path
	}
	return report, nil
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
