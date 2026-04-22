package baseline

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/Pimatis/mavetis/src/model"
	"github.com/Pimatis/mavetis/src/yaml"
)

type Entry struct {
	RuleID string `yaml:"rule"`
	Path   string `yaml:"path"`
	Line   int    `yaml:"line"`
}

type File struct {
	Entries []Entry `yaml:"baseline"`
}

func Load(path string) (File, error) {
	file := File{}
	if path == "" {
		return file, nil
	}
	content, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return file, nil
		}
		return file, fmt.Errorf("read baseline: %w", err)
	}
	value, err := yaml.Parse(string(content))
	if err != nil {
		return file, fmt.Errorf("parse baseline: %w", err)
	}
	mapped, err := yaml.Map(value)
	if err != nil {
		return file, fmt.Errorf("decode baseline: %w", err)
	}
	items, err := yaml.List(mapped["baseline"])
	if err != nil {
		return file, nil
	}
	for _, item := range items {
		entryMap, err := yaml.Map(item)
		if err != nil {
			continue
		}
		entry := Entry{}
		entry.RuleID, _ = yaml.String(entryMap["rule"])
		entry.Path, _ = yaml.String(entryMap["path"])
		line, ok := yaml.Float(entryMap["line"])
		if ok {
			entry.Line = int(line)
		}
		file.Entries = append(file.Entries, entry)
	}
	return file, nil
}

func Create(path string, report model.Report) error {
	entries := make([]Entry, 0, len(report.Findings))
	seen := map[string]struct{}{}
	for _, f := range report.Findings {
		key := fmt.Sprintf("%s|%s|%d", f.RuleID, f.Path, f.Line)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		entries = append(entries, Entry{
			RuleID: f.RuleID,
			Path:   f.Path,
			Line:   f.Line,
		})
	}
	sort.Slice(entries, func(i, j int) bool {
		if entries[i].Path != entries[j].Path {
			return entries[i].Path < entries[j].Path
		}
		if entries[i].Line != entries[j].Line {
			return entries[i].Line < entries[j].Line
		}
		return entries[i].RuleID < entries[j].RuleID
	})
	var b strings.Builder
	b.WriteString("# Mavetis baseline\n")
	b.WriteString("# Known findings suppressed in subsequent reviews\n\n")
	b.WriteString("baseline:\n")
	for _, e := range entries {
		b.WriteString(fmt.Sprintf("  - rule: %s\n", e.RuleID))
		b.WriteString(fmt.Sprintf("    path: %s\n", e.Path))
		b.WriteString(fmt.Sprintf("    line: %d\n", e.Line))
	}
	return os.WriteFile(path, []byte(b.String()), 0644)
}

func Filter(report model.Report, baseline File) model.Report {
	if len(baseline.Entries) == 0 {
		return report
	}
	seen := map[string]struct{}{}
	for _, e := range baseline.Entries {
		seen[fmt.Sprintf("%s|%s|%d", e.RuleID, e.Path, e.Line)] = struct{}{}
	}
	filtered := make([]model.Finding, 0, len(report.Findings))
	for _, f := range report.Findings {
		key := fmt.Sprintf("%s|%s|%d", f.RuleID, f.Path, f.Line)
		if _, ok := seen[key]; ok {
			continue
		}
		filtered = append(filtered, f)
	}
	report.Findings = filtered
	return report
}
