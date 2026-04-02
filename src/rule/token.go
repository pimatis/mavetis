package rule

import "github.com/Pimatis/mavetis/src/model"

func token() []model.Rule {
	return []model.Rule{
		{
			ID:          "token.claims.unchecked",
			Title:       "Token claims validation appears incomplete",
			Message:     "The diff validates or parses a token without nearby issuer, audience, or expiry validation signals.",
			Remediation: "Validate signature, issuer, audience, expiry, not-before, and token type claims before trusting the token.",
			Category:    "token",
			Severity:    "medium",
			Confidence:  "low",
			Target:      "added",
			Paths:       codeFiles(),
			Require:     []string{`(?i)(ParseWithClaims|verify\(|jwt\.Verify|token\.Verify|id token)`},
			Near:        []string{`(?i)(jwt|token|claims|oidc|oauth)`},
			Absent:      []string{`(?i)(issuer|audience|expires|expiry|exp|nbf|validateClaims|WithAudience|WithIssuer|NotBefore)`},
			Standards:   standard("OWASP-ASVS-V3.5", "OWASP-Authentication"),
		},
		{
			ID:          "token.refresh.rotation.deleted",
			Title:       "Refresh token rotation removed",
			Message:     "The diff removes a probable refresh token rotation or old-token revocation step.",
			Remediation: "Rotate refresh tokens on use and revoke the previous token to reduce replay risk.",
			Category:    "token",
			Severity:    "high",
			Confidence:  "medium",
			Target:      "deleted",
			Paths:       codeFiles(),
			Require:     []string{`(?i)(rotateRefresh|refresh.*rotate|revoke.*refresh|delete.*refresh|invalidate.*refresh)`},
			Near:        []string{`(?i)(refresh|token)`},
			Standards:   standard("OWASP-ASVS-V3.5", "OWASP-Authentication"),
		},
		{
			ID:          "token.binding.deleted",
			Title:       "Token binding or jti replay check removed",
			Message:     "The diff removes a probable jti, token binding, or proof-of-possession check.",
			Remediation: "Keep replay controls such as jti tracking, nonce validation, or proof-of-possession checks where the flow depends on them.",
			Category:    "token",
			Severity:    "high",
			Confidence:  "medium",
			Target:      "deleted",
			Paths:       codeFiles(),
			Require:     []string{`(?i)(jti|tokenBinding|proofOfPossession|cnf|nonce|replay)`},
			Near:        []string{`(?i)(token|jwt|oauth|oidc)`},
			Standards:   standard("OWASP-ASVS-V3.5", "OWASP-Authentication"),
		},
	}
}
