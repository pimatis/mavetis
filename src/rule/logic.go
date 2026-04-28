package rule

import "github.com/Pimatis/mavetis/src/model"

func logic() []model.Rule {
	return []model.Rule{
		{
			ID:          "logic.mass.assignment",
			Title:       "Mass assignment risk introduced",
			Message:     "The diff binds request-controlled input directly into a model or entity without field filtering.",
			Remediation: "Use explicit allowlists for assignable fields and avoid binding raw request bodies to internal models.",
			Category:    "logic",
			Severity:    "high",
			Confidence:  "medium",
			Target:      "added",
			Paths:       codeFiles(),
			Require:     []string{`(?i)(Object\.assign\(|\.assign\(|\.merge\(|\.set\(|create\(|update\(|save\(|insert\()`},
			Near:        []string{`(?i)(req\.body|request\.body|ctx\.request\.body|body|input|params|data)`},
			Absent:      []string{`(?i)(pick|select|whitelist|allowlist|permitted|fillable|guarded|validation|sanitize)`},
			Standards:   standard("OWASP-ASVS-V11.1", "OWASP-Business-Logic"),
		},
		{
			ID:          "logic.price.tampering",
			Title:       "Price or amount parameter tampering risk introduced",
			Message:     "The diff uses a request-controlled value for price, amount, or quantity without server-side verification.",
			Remediation: "Validate and recalculate price, total, and quantity on the server side and reject client-supplied monetary values.",
			Category:    "logic",
			Severity:    "critical",
			Confidence:  "medium",
			Target:      "added",
			Paths:       codeFiles(),
			Require:     []string{`(?i)(price|amount|total|quantity|cost|fee|discount|tax).{0,20}(=|:).{0,20}(req\.|request\.|ctx\.|params\.|query\.|body\.|input\.|userInput)`},
			Absent:      []string{`(?i)(serverSide|recalculate|verifyPrice|validateAmount|checkTotal|computedPrice|lookupPrice)`},
			Standards:   standard("OWASP-ASVS-V11.1", "OWASP-Business-Logic"),
		},
	}
}
