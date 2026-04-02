package diff

import (
	"github.com/Pimatis/mavetis/src/match"
	"github.com/Pimatis/mavetis/src/model"
)

func Filter(input model.Diff, pattern string) model.Diff {
	if pattern == "" {
		return input
	}
	result := model.Diff{Meta: input.Meta}
	for _, file := range input.Files {
		if !match.Glob(pattern, file.Path) {
			continue
		}
		result.Files = append(result.Files, file)
	}
	return result
}
