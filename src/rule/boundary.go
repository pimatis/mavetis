package rule

import "github.com/Pimatis/mavetis/src/model"

func boundary() []model.Rule {
	return []model.Rule{
		{
			ID:          "boundary.admin.public",
			Type:        "pathBoundary",
			Title:       "Public route imports internal admin module",
			Message:     "The diff imports an internal or admin module from a public route or handler surface.",
			Remediation: "Keep privileged admin modules behind internal-only boundaries and expose reviewed adapters instead of direct imports.",
			Category:    "boundary",
			Severity:    "critical",
			Confidence:  "high",
			Target:      "added",
			Paths:       []string{"**/routes/**", "**/route/**", "**/api/**", "**/handlers/**", "**/public/**"},
			Imports:     []string{`(?i)(^|/)(internal/admin|admin/internal)(/|$)`},
			Standards:   standard("OWASP-ASVS-V4.1", "OWASP-Architecture"),
		},
		{
			ID:          "boundary.ui.auth",
			Type:        "pathBoundary",
			Title:       "UI layer imports privileged auth or security helper",
			Message:     "The diff imports an auth, security, or internal helper directly into a UI-facing surface.",
			Remediation: "Keep privileged auth and security helpers behind server-side boundaries and expose only reviewed client-safe interfaces.",
			Category:    "boundary",
			Severity:    "high",
			Confidence:  "high",
			Target:      "added",
			Paths:       []string{"**/ui/**", "**/views/**", "**/pages/**", "**/components/**", "**/*.tsx", "**/*.jsx"},
			Imports:     []string{`(?i)(^|/)(auth|security|internal|admin)(/|$)`},
			Standards:   standard("OWASP-ASVS-V4.1", "OWASP-Architecture"),
		},
		{
			ID:          "boundary.privileged.public",
			Type:        "pathBoundary",
			Title:       "Public surface imports privileged service",
			Message:     "The diff imports a privileged, internal, or backoffice service into a publicly reachable surface.",
			Remediation: "Keep privileged services isolated from public handlers and expose only narrow reviewed interfaces across boundaries.",
			Category:    "boundary",
			Severity:    "critical",
			Confidence:  "medium",
			Target:      "added",
			Paths:       []string{"**/routes/**", "**/route/**", "**/api/**", "**/public/**"},
			Imports:     []string{`(?i)(^|/)(privileged|backoffice|root|superuser|internal/service)(/|$)`},
			Standards:   standard("OWASP-ASVS-V4.1", "OWASP-Architecture"),
		},
	}
}
