package rule

import "github.com/Pimatis/mavetis/src/model"

func client() []model.Rule {
	return []model.Rule{
		{
			ID:          "client.postmessage.origin",
			Title:       "postMessage without origin check",
			Message:     "The diff sends a postMessage without validating the target origin.",
			Remediation: "Always specify a target origin in postMessage and validate the origin of incoming messages.",
			Category:    "client",
			Severity:    "high",
			Confidence:  "high",
			Target:      "added",
			Paths:       webFiles(),
			Require:     []string{`(?i)(postMessage\()`},
			Absent:      []string{`(?i)(origin|event\.origin|\.origin)`},
			Standards:   standard("OWASP-ASVS-V3.2", "OWASP-Client"),
		},
		{
			ID:          "client.opener.blank",
			Title:       "target=_blank without rel=noopener",
			Message:     "The diff opens a link in a new tab without rel=noopener or noreferrer.",
			Remediation: "Add rel=noopener noreferrer to all links with target=_blank to prevent tabnabbing.",
			Category:    "client",
			Severity:    "medium",
			Confidence:  "high",
			Target:      "added",
			Paths:       webFiles(),
			Require:     []string{`(?i)(target\s*=\s*["\']_blank["\'])`},
			Absent:      []string{`(?i)(rel\s*=\s*["\'].*noopener|noreferrer)`},
			Standards:   standard("OWASP-ASVS-V3.2", "OWASP-Client"),
		},
		{
			ID:          "client.document.domain",
			Title:       "document.domain loosened",
			Message:     "The diff sets document.domain to a broader domain.",
			Remediation: "Avoid setting document.domain; use postMessage or CORS for cross-origin communication instead.",
			Category:    "client",
			Severity:    "medium",
			Confidence:  "high",
			Target:      "added",
			Paths:       webFiles(),
			Require:     []string{`(?i)(document\.domain\s*=)`},
			Standards:   standard("OWASP-ASVS-V3.2", "OWASP-Client"),
		},
	}
}
