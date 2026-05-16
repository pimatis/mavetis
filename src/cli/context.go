package cli

import (
	"github.com/Pimatis/mavetis/src/model"
	"github.com/Pimatis/mavetis/src/resolve"
	"github.com/Pimatis/mavetis/src/scan"
)

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
