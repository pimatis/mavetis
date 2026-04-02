package model

type Diff struct {
	Files []DiffFile
	Meta  DiffMeta
}

type DiffMeta struct {
	Mode string
	Base string
	Head string
}

type DiffFile struct {
	Path    string
	OldPath string
	Change  string
	Hunks   []DiffHunk
}

type DiffHunk struct {
	OldStart int
	OldLines int
	NewStart int
	NewLines int
	Header   string
	Lines    []DiffLine
}

type DiffLine struct {
	Kind      string
	Text      string
	OldNumber int
	NewNumber int
}

func (file DiffFile) Flatten() []DiffLine {
	lines := make([]DiffLine, 0)
	for _, hunk := range file.Hunks {
		lines = append(lines, hunk.Lines...)
	}
	return lines
}
