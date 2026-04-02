package engine

import (
	"strings"

	"github.com/Pimatis/mavetis/src/analyze"
	"github.com/Pimatis/mavetis/src/match"
	"github.com/Pimatis/mavetis/src/model"
)

func intentFindings(diff model.Diff) []model.Finding {
	findings := make([]model.Finding, 0)
	for _, file := range diff.Files {
		if !analyze.Executable(file.Path) {
			continue
		}
		if analyze.Fixture(file.Path) {
			continue
		}
		for _, hunk := range file.Hunks {
			text := join(hunk)
			anchors := uniqueAnchors(analyze.SecurityAnchorsFromText(text))
			if len(anchors) == 0 {
				continue
			}
			for _, anchor := range anchors {
				deletedText := strings.ToLower(stripAnchorLines(joinDeleted(hunk), anchor.Name))
				addedText := strings.ToLower(stripAnchorLines(joinAdded(hunk), anchor.Name))
				if !strings.Contains(strings.ToLower(text), strings.ToLower(anchor.Name)) {
					continue
				}
				if !intentMismatch(anchor, deletedText, addedText) {
					continue
				}
				findings = append(findings, syntheticFinding("intent.mismatch", "Security intent no longer matches implementation", anchor.Category, anchor.Severity, file.Path, hunk, "The diff changes a security-named function while removing or weakening the expected protective behavior.", "Restore the missing security behavior or rename the function so its intent stays honest and reviewable.", "security-sensitive function anchor: "+anchor.Name, "expected security tokens disappeared or bypass-like terms appeared in the replacement"))
			}
		}
	}
	return findings
}

func snapshotFindings(diff model.Diff, snapshots []model.Snapshot) []model.Finding {
	if len(snapshots) == 0 {
		return nil
	}
	findings := make([]model.Finding, 0)
	for _, file := range diff.Files {
		for _, snapshot := range snapshots {
			if snapshot.Path != file.Path && !match.Any([]string{snapshot.Path}, file.Path) {
				continue
			}
			for _, hunk := range file.Hunks {
				text := strings.ToLower(join(hunk))
				if !strings.Contains(text, strings.ToLower(snapshot.Anchor)) {
					continue
				}
				deleted := strings.ToLower(stripAnchorLines(joinDeleted(hunk), snapshot.Anchor))
				added := strings.ToLower(stripAnchorLines(joinAdded(hunk), snapshot.Anchor))
				if !sharesTokens(deleted, snapshot.Require) {
					continue
				}
				if sharesTokens(added, snapshot.Require) {
					continue
				}
				findings = append(findings, model.Finding{
					ID:          identity(snapshot.ID, file.Path, hunkLine(hunk), "added", snapshot.Anchor),
					RuleID:      snapshot.ID,
					Title:       "Repository security snapshot regressed",
					Category:    snapshot.Category,
					Severity:    snapshot.Severity,
					Confidence:  "high",
					Path:        file.Path,
					Line:        hunkLine(hunk),
					Side:        "added",
					Message:     snapshot.Message,
					Snippet:     "snapshot anchor: " + snapshot.Anchor,
					Remediation: snapshot.Remediation,
					Reasons:     []string{"snapshot anchor matched the changed hunk", "required baseline tokens disappeared: " + strings.Join(snapshot.Require, ", ")},
					Standards:   append([]string{}, snapshot.Standards...),
				})
				break
			}
		}
	}
	return findings
}

func intentMismatch(anchor analyze.Anchor, deletedText string, addedText string) bool {
	if !sharesTokens(deletedText, anchor.Expected) {
		return false
	}
	if sharesTokens(addedText, anchor.Expected) {
		if !sharesTokens(addedText, anchor.Bypass) {
			return false
		}
	}
	if sharesTokens(addedText, anchor.Bypass) {
		return true
	}
	return !sharesTokens(addedText, anchor.Expected)
}

func sharesTokens(text string, wants []string) bool {
	for _, item := range wants {
		if strings.Contains(text, strings.ToLower(item)) {
			return true
		}
	}
	return false
}

func joinDeleted(hunk model.DiffHunk) string {
	parts := make([]string, 0, len(hunk.Lines))
	for _, line := range hunk.Lines {
		if line.Kind != "deleted" {
			continue
		}
		parts = append(parts, line.Text)
	}
	return strings.Join(parts, "\n")
}

func joinAdded(hunk model.DiffHunk) string {
	parts := make([]string, 0, len(hunk.Lines))
	for _, line := range hunk.Lines {
		if line.Kind != "added" {
			continue
		}
		parts = append(parts, line.Text)
	}
	return strings.Join(parts, "\n")
}

func hunkLine(hunk model.DiffHunk) int {
	for _, line := range hunk.Lines {
		if line.Kind == "added" || line.Kind == "deleted" {
			return number(line)
		}
	}
	return 1
}

func uniqueAnchors(items []analyze.Anchor) []analyze.Anchor {
	seen := map[string]struct{}{}
	result := make([]analyze.Anchor, 0, len(items))
	for _, item := range items {
		key := strings.ToLower(item.Name)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		result = append(result, item)
	}
	return result
}

func stripAnchorLines(text string, anchor string) string {
	lines := strings.Split(text, "\n")
	filtered := make([]string, 0, len(lines))
	lower := strings.ToLower(anchor)
	for _, line := range lines {
		if strings.Contains(strings.ToLower(line), lower) {
			continue
		}
		filtered = append(filtered, line)
	}
	return strings.Join(filtered, "\n")
}
