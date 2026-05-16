package rule

import "github.com/Pimatis/mavetis/src/model"

func file() []model.Rule {
	return []model.Rule{
		{
			ID:          "file.sourcemap.exposed",
			Title:       "Source map exposed in production",
			Message:     "The diff deploys a source map file to a public path.",
			Remediation: "Remove source maps from production builds or restrict them to authenticated internal access.",
			Category:    "file",
			Severity:    "medium",
			Confidence:  "high",
			Target:      "added",
			Paths:       []string{"**/*.map", "**/dist/**", "**/build/**", "**/public/**", "**/static/**"},
			Require:     []string{`(?i)(//# sourceMappingURL=|\.js\.map|\.css\.map)`},
			Standards:   standard("OWASP-ASVS-V14.4", "OWASP-File"),
		},
		{
			ID:          "file.git.exposed",
			Title:       ".git directory exposed",
			Message:     "The diff exposes the .git directory in a public path.",
			Remediation: "Block access to .git and other version control metadata in the web server configuration.",
			Category:    "file",
			Severity:    "high",
			Confidence:  "high",
			Target:      "added",
			Paths:       []string{"**/.git/**", "**/public/**", "**/static/**", "**/dist/**", "**/build/**"},
			Require:     []string{`(?i)(\.git/|\.git/config)`},
			Standards:   standard("OWASP-ASVS-V14.4", "OWASP-File"),
		},
		{
			ID:          "file.backup.exposed",
			Title:       "Backup or temporary file exposed",
			Message:     "The diff exposes a backup or temporary file in a public path.",
			Remediation: "Remove backup and temporary files from production deployments.",
			Category:    "file",
			Severity:    "medium",
			Confidence:  "high",
			Target:      "added",
			Paths:       []string{"**/public/**", "**/static/**", "**/dist/**", "**/build/**"},
			Require:     []string{`(?i)(\.bak|\.old|\.swp|\.tmp|\.backup|~)$`},
			Standards:   standard("OWASP-ASVS-V14.4", "OWASP-File"),
		},
	}
}
