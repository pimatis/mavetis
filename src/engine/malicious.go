package engine

import (
	"strings"

	"github.com/Pimatis/mavetis/src/analyze"
	"github.com/Pimatis/mavetis/src/model"
)

func suspiciousName(hunk model.DiffHunk) string {
	for _, line := range hunk.Lines {
		if line.Kind != "added" {
			continue
		}
		text := strings.TrimSpace(line.Text)
		parts := strings.FieldsFunc(text, func(char rune) bool {
			return char == '"' || char == ':' || char == ' ' || char == '=' || char == '>' || char == '<'
		})
		for _, part := range parts {
			if part == "" {
				continue
			}
			if suspect := analyze.SuspiciousPackage(part); suspect != "" {
				return suspect
			}
		}
	}
	return ""
}

func manifestFindings(diff model.Diff) []model.Finding {
	findings := make([]model.Finding, 0)
	for _, file := range diff.Files {
		if analyze.Fixture(file.Path) {
			continue
		}
		language := analyze.Language(file.Path)
		isManifest := strings.HasSuffix(file.Path, "package.json") || strings.HasSuffix(file.Path, "go.mod") || strings.HasSuffix(file.Path, ".npmrc") || strings.HasSuffix(file.Path, "pip.conf") || strings.HasSuffix(file.Path, "poetry.toml")
		if language != "config" && !isManifest {
			continue
		}
		for _, hunk := range file.Hunks {
			text := strings.ToLower(join(hunk))
			if strings.Contains(text, "registry=https://registry.npmjs.org") || strings.Contains(text, "index-url = https://pypi.org/simple") {
				findings = append(findings, syntheticFinding("supply.registry.public", "Registry source points to a broad public index", "supply", "high", file.Path, hunk, "The diff points dependency resolution to a broad public registry or index source.", "Review whether this change weakens dependency trust boundaries or private package resolution rules.", "manifest or config changed package registry settings", "public index source appeared in the same hunk"))
			}
			if strings.Contains(text, `"latest"`) || strings.Contains(text, `"*"`) || strings.Contains(text, `"next"`) {
				findings = append(findings, syntheticFinding("supply.version.floating", "Floating dependency version introduced", "supply", "medium", file.Path, hunk, "The diff introduces a floating package version that can reduce build determinism and reviewability.", "Pin dependency versions and keep lockfiles committed so upgrades stay explicit.", "dependency version is not pinned to a stable release", "floating version text appeared in the same hunk"))
			}
			if strings.Contains(text, "replace ") && strings.Contains(text, "github.com") {
				findings = append(findings, syntheticFinding("supply.replace.remote", "Remote Go module replacement introduced", "supply", "high", file.Path, hunk, "The diff introduces a Go module replace directive that points to a remote source.", "Review the replacement source carefully and prefer pinned trusted origins for release builds.", "go.mod replacement logic appeared in the same hunk", "remote source indicator appeared in the replacement"))
			}
			if suspect := suspiciousName(hunk); suspect != "" {
				findings = append(findings, syntheticFinding("supply.typosquat", "Package name resembles a likely typosquat", "supply", "high", file.Path, hunk, "The diff introduces a package name that is one edit away from a widely used dependency name.", "Review the package source carefully and verify that the name is intentional before merge.", "package name appears unusually similar to a popular dependency", "closest known package: "+suspect))
			}
		}
	}
	return findings
}
