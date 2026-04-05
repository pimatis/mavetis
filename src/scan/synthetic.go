package scan

import (
	"strings"

	"github.com/Pimatis/mavetis/src/model"
)

func FromFiles(files []ScannedFile) model.Diff {
	result := model.Diff{Meta: model.DiffMeta{Mode: "file"}}
	for _, file := range files {
		lines := splitLines(file.Content)
		diffLines := make([]model.DiffLine, 0, len(lines))
		for index, line := range lines {
			diffLines = append(diffLines, model.DiffLine{Kind: "added", Text: line, NewNumber: index + 1})
		}
		result.Files = append(result.Files, model.DiffFile{
			Path:   file.Path,
			Change: "modified",
			Hunks: []model.DiffHunk{{
				OldStart: 0,
				OldLines: 0,
				NewStart: 1,
				NewLines: len(diffLines),
				Header:   "@@ file review @@",
				Lines:    diffLines,
			}},
		})
	}
	return result
}

func splitLines(content string) []string {
	normalized := strings.ReplaceAll(content, "\r\n", "\n")
	lines := strings.Split(normalized, "\n")
	if len(lines) == 0 {
		return nil
	}
	if lines[len(lines)-1] == "" {
		return lines[:len(lines)-1]
	}
	return lines
}
