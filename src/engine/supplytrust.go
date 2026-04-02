package engine

import (
	"strings"

	"github.com/Pimatis/mavetis/src/analyze"
	"github.com/Pimatis/mavetis/src/model"
)

func supplyTrustFindings(diff model.Diff, config model.Config) []model.Finding {
	findings := make([]model.Finding, 0)
	lockTouched := lockfileTouched(diff)
	dependencyAdded := false
	lifecycleAdded := false
	for _, file := range diff.Files {
		if analyze.Fixture(file.Path) {
			continue
		}
		for _, hunk := range file.Hunks {
			text := strings.ToLower(join(hunk))
			if strings.HasSuffix(file.Path, "package.json") {
				if strings.Contains(text, "\"dependencies\"") || strings.Contains(text, "\"devdependencies\"") || strings.Contains(text, "\"optionaldependencies\"") {
					dependencyAdded = dependencyAdded || hasDependencyAdded(hunk)
				}
			}
			if strings.HasSuffix(file.Path, "package.json") {
				if strings.Contains(text, "postinstall") || strings.Contains(text, "preinstall") || strings.Contains(text, "prepare") {
					lifecycleAdded = lifecycleAdded || hasAddedLifecycle(hunk)
				}
			}
			findings = append(findings, registryTrustFindings(file.Path, hunk, config.Supply.TrustedRegistries)...)
			findings = append(findings, packagePolicyFindings(file.Path, hunk, config.Supply)...)
		}
	}
	if dependencyAdded && lifecycleAdded {
		findings = append(findings, crossFinding("supply.lifecycle.dependency", "Dependency addition paired with lifecycle script", "supply", "critical", "The branch adds package dependencies while also enabling install-time lifecycle scripts.", "Require explicit review for lifecycle scripts when dependency trust changes in the same branch."))
	}
	if dependencyAdded && !lockTouched {
		findings = append(findings, crossFinding("supply.lock.missing", "Dependency manifest changed without lockfile update", "supply", "high", "The branch changes dependency manifests without updating a matching lockfile.", "Update and review lockfiles together with dependency manifest changes so resolution remains deterministic."))
	}
	return findings
}

func registryTrustFindings(path string, hunk model.DiffHunk, trusted []string) []model.Finding {
	findings := make([]model.Finding, 0)
	deletedPrivate := false
	addedPublic := false
	addedUntrusted := ""
	for _, line := range hunk.Lines {
		lower := strings.ToLower(line.Text)
		if line.Kind == "deleted" && strings.Contains(lower, "registry") && strings.Contains(lower, "https://") && !strings.Contains(lower, "registry.npmjs.org") && !strings.Contains(lower, "pypi.org/simple") {
			deletedPrivate = true
		}
		if line.Kind == "added" && (strings.Contains(lower, "registry=https://registry.npmjs.org") || strings.Contains(lower, "index-url = https://pypi.org/simple")) {
			addedPublic = true
		}
		if line.Kind == "added" && strings.Contains(lower, "registry") && len(trusted) != 0 {
			trustedMatch := false
			for _, item := range trusted {
				if strings.Contains(lower, strings.ToLower(item)) {
					trustedMatch = true
				}
			}
			if !trustedMatch {
				addedUntrusted = strings.TrimSpace(line.Text)
			}
		}
	}
	if deletedPrivate && addedPublic {
		findings = append(findings, syntheticFinding("supply.registry.drift", "Private registry drifted to public registry", "supply", "critical", path, hunk, "The diff replaces a private or constrained package registry with a broad public registry.", "Keep dependency resolution pinned to reviewed registries and treat public-registry drift as a supply-chain event.", "a private registry entry was removed", "a public registry entry was added in the same hunk"))
	}
	if addedUntrusted != "" {
		findings = append(findings, syntheticFinding("supply.registry.untrusted", "Registry is outside trusted allowlist", "supply", "high", path, hunk, "The diff points dependency resolution to a registry that is outside the configured trusted registry set.", "Use only approved registries or expand the trust policy after explicit review.", "trusted registries are configured", "added registry line: "+addedUntrusted))
	}
	return findings
}

func packagePolicyFindings(path string, hunk model.DiffHunk, supply model.Supply) []model.Finding {
	findings := make([]model.Finding, 0)
	for _, line := range hunk.Lines {
		if line.Kind != "added" {
			continue
		}
		name := dependencyName(line.Text)
		if name == "" {
			continue
		}
		if matchesPackagePolicy(name, supply.DenyPackages) {
			findings = append(findings, syntheticFinding("supply.package.denied", "Dependency is denied by trust policy", "supply", "critical", path, hunk, "The diff adds a dependency that is explicitly denied by the configured package policy.", "Remove the denied dependency or update the trust policy only after security review.", "configured deny-packages policy matched", "added dependency: "+name))
			continue
		}
		if len(supply.AllowPackages) != 0 && !matchesPackagePolicy(name, supply.AllowPackages) {
			findings = append(findings, syntheticFinding("supply.package.untrusted", "Dependency falls outside package allowlist", "supply", "high", path, hunk, "The diff adds a dependency that is outside the configured allowlist policy.", "Add only approved packages or expand the allowlist after explicit trust review.", "configured allow-packages policy did not match", "added dependency: "+name))
		}
	}
	return findings
}

func lockfileTouched(diff model.Diff) bool {
	for _, file := range diff.Files {
		path := strings.ToLower(file.Path)
		if strings.HasSuffix(path, "package-lock.json") || strings.HasSuffix(path, "bun.lock") || strings.HasSuffix(path, "bun.lockb") || strings.HasSuffix(path, "pnpm-lock.yaml") || strings.HasSuffix(path, "yarn.lock") {
			return true
		}
	}
	return false
}

func hasDependencyAdded(hunk model.DiffHunk) bool {
	for _, line := range hunk.Lines {
		if line.Kind != "added" {
			continue
		}
		if dependencyName(line.Text) != "" {
			return true
		}
	}
	return false
}

func hasAddedLifecycle(hunk model.DiffHunk) bool {
	for _, line := range hunk.Lines {
		if line.Kind != "added" {
			continue
		}
		lower := strings.ToLower(line.Text)
		if strings.Contains(lower, "postinstall") || strings.Contains(lower, "preinstall") || strings.Contains(lower, "prepare") {
			return true
		}
	}
	return false
}

func dependencyName(line string) string {
	trimmed := strings.TrimSpace(line)
	if !strings.HasPrefix(trimmed, "\"") {
		return ""
	}
	parts := strings.SplitN(trimmed, ":", 2)
	if len(parts) != 2 {
		return ""
	}
	name := strings.Trim(parts[0], `" `)
	if name == "dependencies" || name == "devDependencies" || name == "optionalDependencies" || name == "peerDependencies" || name == "scripts" {
		return ""
	}
	return name
}

func matchesPackagePolicy(name string, patterns []string) bool {
	lower := strings.ToLower(name)
	for _, item := range patterns {
		pattern := strings.ToLower(strings.TrimSpace(item))
		if pattern == "" {
			continue
		}
		if pattern == lower {
			return true
		}
		if strings.HasSuffix(pattern, "*") && strings.HasPrefix(lower, strings.TrimSuffix(pattern, "*")) {
			return true
		}
	}
	return false
}
