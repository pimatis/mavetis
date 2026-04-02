package diff

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/Pimatis/mavetis/src/model"
)

func Parse(input string, meta model.DiffMeta) (model.Diff, error) {
	result := model.Diff{Meta: meta}
	lines := strings.Split(strings.ReplaceAll(input, "\r\n", "\n"), "\n")
	file := model.DiffFile{}
	hunk := model.DiffHunk{}
	oldCursor := 0
	newCursor := 0
	hasFile := false
	hasHunk := false
	flushHunk := func() {
		if !hasHunk {
			return
		}
		file.Hunks = append(file.Hunks, hunk)
		hunk = model.DiffHunk{}
		hasHunk = false
	}
	flushFile := func() {
		flushHunk()
		if !hasFile {
			return
		}
		if file.Path != "" || len(file.Hunks) != 0 {
			result.Files = append(result.Files, file)
		}
		file = model.DiffFile{}
		hasFile = false
	}
	for _, value := range lines {
		if strings.HasPrefix(value, "diff --git ") {
			flushFile()
			parts := strings.Fields(value)
			if len(parts) < 4 {
				return result, fmt.Errorf("invalid diff header: %s", value)
			}
			file = model.DiffFile{OldPath: trim(parts[2]), Path: trim(parts[3]), Change: "modified"}
			hasFile = true
			continue
		}
		if !hasFile {
			continue
		}
		if strings.HasPrefix(value, "new file mode ") {
			file.Change = "added"
			continue
		}
		if strings.HasPrefix(value, "deleted file mode ") {
			file.Change = "deleted"
			continue
		}
		if strings.HasPrefix(value, "rename from ") {
			file.Change = "renamed"
			file.OldPath = strings.TrimSpace(strings.TrimPrefix(value, "rename from "))
			continue
		}
		if strings.HasPrefix(value, "rename to ") {
			file.Path = strings.TrimSpace(strings.TrimPrefix(value, "rename to "))
			continue
		}
		if strings.HasPrefix(value, "--- ") {
			file.OldPath = trim(strings.TrimSpace(strings.TrimPrefix(value, "--- ")))
			continue
		}
		if strings.HasPrefix(value, "+++ ") {
			file.Path = trim(strings.TrimSpace(strings.TrimPrefix(value, "+++ ")))
			continue
		}
		if strings.HasPrefix(value, "@@ ") {
			flushHunk()
			parsed, err := header(value)
			if err != nil {
				return result, err
			}
			hunk = parsed
			oldCursor = parsed.OldStart
			newCursor = parsed.NewStart
			hasHunk = true
			continue
		}
		if !hasHunk {
			continue
		}
		if value == "" {
			continue
		}
		if value == "\\ No newline at end of file" {
			continue
		}
		line, err := mapLine(value, &oldCursor, &newCursor)
		if err != nil {
			return result, err
		}
		hunk.Lines = append(hunk.Lines, line)
	}
	flushFile()
	return result, nil
}

func trim(value string) string {
	trimmed := strings.TrimPrefix(value, "a/")
	trimmed = strings.TrimPrefix(trimmed, "b/")
	if trimmed == "/dev/null" {
		return ""
	}
	return trimmed
}

func header(value string) (model.DiffHunk, error) {
	result := model.DiffHunk{Header: value}
	parts := strings.Fields(value)
	if len(parts) < 3 {
		return result, fmt.Errorf("invalid hunk header: %s", value)
	}
	oldStart, oldLines, err := rangev(strings.TrimPrefix(parts[1], "-"))
	if err != nil {
		return result, err
	}
	newStart, newLines, err := rangev(strings.TrimPrefix(parts[2], "+"))
	if err != nil {
		return result, err
	}
	result.OldStart = oldStart
	result.OldLines = oldLines
	result.NewStart = newStart
	result.NewLines = newLines
	return result, nil
}

func rangev(value string) (int, int, error) {
	parts := strings.Split(value, ",")
	start, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, 0, fmt.Errorf("invalid range: %s", value)
	}
	if len(parts) == 1 {
		return start, 1, nil
	}
	count, err := strconv.Atoi(parts[1])
	if err != nil {
		return 0, 0, fmt.Errorf("invalid range: %s", value)
	}
	return start, count, nil
}

func mapLine(value string, oldCursor *int, newCursor *int) (model.DiffLine, error) {
	result := model.DiffLine{}
	if value == "" {
		return result, fmt.Errorf("invalid diff line")
	}
	prefix := value[0]
	text := ""
	if len(value) > 1 {
		text = value[1:]
	}
	result.Text = text
	if prefix == ' ' {
		result.Kind = "context"
		result.OldNumber = *oldCursor
		result.NewNumber = *newCursor
		*oldCursor = *oldCursor + 1
		*newCursor = *newCursor + 1
		return result, nil
	}
	if prefix == '+' {
		result.Kind = "added"
		result.NewNumber = *newCursor
		*newCursor = *newCursor + 1
		return result, nil
	}
	if prefix == '-' {
		result.Kind = "deleted"
		result.OldNumber = *oldCursor
		*oldCursor = *oldCursor + 1
		return result, nil
	}
	return result, fmt.Errorf("invalid diff line prefix: %s", value)
}
