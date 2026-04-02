package rule

import "github.com/Pimatis/mavetis/src/model"

func authn() []model.Rule {
	return []model.Rule{
		{
			ID:          "auth.middleware.deleted",
			Title:       "Authentication middleware removed",
			Message:     "A probable authentication middleware or guard was removed from the diff.",
			Remediation: "Restore the guard or replace it with an equivalent access control enforcement path.",
			Category:    "auth",
			Severity:    "critical",
			Confidence:  "medium",
			Target:      "deleted",
			Paths:       codeFiles(),
			Require:     []string{`(?i)(auth|authenticate|middleware|guard|requireAuth|verifySession|isAuthenticated)`},
			Standards:   standard("OWASP-ASVS-V3.1", "OWASP-Authentication"),
		},
		{
			ID:          "session.storage.local",
			Title:       "Token persisted in local storage",
			Message:     "The diff stores authentication material in browser local storage.",
			Remediation: "Prefer HttpOnly secure cookies or platform-backed secure storage for bearer tokens.",
			Category:    "auth",
			Severity:    "high",
			Confidence:  "high",
			Target:      "added",
			Paths:       webFiles(),
			Require:     []string{`(?i)localStorage\.(setItem|getItem).*?(token|jwt|session|auth)`},
			Standards:   standard("OWASP-ASVS-V3.4", "OWASP-Authentication"),
		},
		{
			ID:          "token.decode.only",
			Title:       "JWT decode used without verification",
			Message:     "The diff introduces token decoding instead of signature verification.",
			Remediation: "Verify token signatures and claims before use instead of decoding untrusted tokens directly.",
			Category:    "token",
			Severity:    "high",
			Confidence:  "high",
			Target:      "added",
			Paths:       codeFiles(),
			Require:     []string{`(?i)(jwt\.|token\.)decode\(`},
			Absent:      []string{`(?i)(verify|parseWithClaims|validate|CheckSignature)`},
			Standards:   standard("OWASP-ASVS-V3.5", "OWASP-Authentication"),
		},
		{
			ID:          "auth.bypass.debug",
			Title:       "Authentication bypass flag introduced",
			Message:     "The diff introduces a probable authentication bypass or anonymous override flag.",
			Remediation: "Remove the bypass and keep authentication decisions explicit, deny-by-default, and environment-independent.",
			Category:    "auth",
			Severity:    "critical",
			Confidence:  "high",
			Target:      "added",
			Paths:       codeAndConfigFiles(),
			Require:     []string{`(?i)(skipAuth|disableAuth|authDisabled|allowAnonymous|anonymousAccess|requireAuth\s*[:=]\s*false)`},
			Standards:   standard("OWASP-ASVS-V3.1", "OWASP-Authentication"),
		},
	}
}
