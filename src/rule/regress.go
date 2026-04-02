package rule

import "github.com/Pimatis/mavetis/src/model"

func regress() []model.Rule {
	return []model.Rule{
		{
			ID:          "auth.mfa.disabled",
			Title:       "MFA requirement disabled or made optional",
			Message:     "The diff weakens multi-factor authentication by disabling it, making it optional, or adding an MFA bypass flag.",
			Remediation: "Keep MFA enforcement explicit on protected flows and remove optional or bypass behavior from production paths.",
			Category:    "auth",
			Severity:    "critical",
			Confidence:  "high",
			Target:      "added",
			Paths:       codeAndConfigFiles(),
			Require:     []string{`(?i)(requireMFA\s*[:=]\s*false|mfaRequired\s*[:=]\s*false|mfaEnforced\s*[:=]\s*false|enforceMFA\s*[:=]\s*false|skipMFA|disableMFA|mfaOptional|optionalMFA)`},
			Standards:   standard("OWASP-ASVS-V3.1", "OWASP-Authentication"),
		},
		{
			ID:          "auth.mfa.deleted",
			Title:       "MFA verification step removed",
			Message:     "The diff removes a probable MFA, OTP, TOTP, or WebAuthn verification step.",
			Remediation: "Restore the MFA challenge and verification path before allowing the protected flow to complete.",
			Category:    "auth",
			Severity:    "critical",
			Confidence:  "medium",
			Target:      "deleted",
			Paths:       codeFiles(),
			Require:     []string{`(?i)(mfa|otp|totp|webauthn|authenticator)`},
			Near:        []string{`(?i)(verify|challenge|require|validate|check)`},
			Standards:   standard("OWASP-ASVS-V3.1", "OWASP-Authentication"),
		},
		{
			ID:          "auth.ratelimit.deleted",
			Title:       "Authentication rate limiting removed",
			Message:     "The diff removes a probable limiter, threshold, or lockout guard tied to authentication flows.",
			Remediation: "Restore rate limiting or lockout controls around login, password reset, OTP, and MFA verification endpoints.",
			Category:    "auth",
			Severity:    "high",
			Confidence:  "medium",
			Target:      "deleted",
			Paths:       codeAndConfigFiles(),
			Require:     []string{`(?i)(rateLimit|rate_limit|limiter|throttle|maxAttempts|tooManyRequests|lockout)`},
			Near:        []string{`(?i)(login|signin|auth|password|reset|otp|mfa|verify)`},
			Standards:   standard("OWASP-ASVS-V7.2", "OWASP-Abuse-Prevention"),
		},
	}
}
