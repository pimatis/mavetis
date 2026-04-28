package rule

import (
	"fmt"
	"strings"

	"github.com/Pimatis/mavetis/src/model"
)

func secrets(config model.Config) []model.Rule {
	rules := []model.Rule{
		{
			ID:          "secret.aws.access",
			Title:       "AWS access key exposed",
			Message:     "An AWS access key pattern was added to the diff.",
			Remediation: "Remove the key from code, rotate it, and load credentials from a secure runtime source.",
			Category:    "secret",
			Severity:    "critical",
			Confidence:  "high",
			Target:      "added",
			Require:     []string{`\bAKIA[0-9A-Z]{16}\b`},
			Standards:   standard("OWASP-ASVS-V8", "OWASP-Secrets"),
			Mask:        true,
		},
		{
			ID:          "secret.aws.secret",
			Title:       "AWS secret key exposed",
			Message:     "A probable AWS secret key was added to the diff.",
			Remediation: "Remove the secret, rotate it, and replace it with environment-backed credential loading.",
			Category:    "secret",
			Severity:    "critical",
			Confidence:  "medium",
			Target:      "added",
			Require:     []string{`(?i)aws(.{0,20})?(secret|access).{0,10}[=:].{0,10}[A-Za-z0-9/+=]{40}`},
			Entropy:     3.8,
			Standards:   standard("OWASP-ASVS-V8", "OWASP-Secrets"),
			Mask:        true,
		},
		{
			ID:          "secret.stripe",
			Title:       "Stripe secret key exposed",
			Message:     "A Stripe secret key was added to the diff.",
			Remediation: "Remove the key, rotate it, and inject it through a secure secret source.",
			Category:    "secret",
			Severity:    "critical",
			Confidence:  "high",
			Target:      "added",
			Require:     []string{`\bsk_(live|test)_[A-Za-z0-9]{16,}\b`},
			Standards:   standard("OWASP-ASVS-V8", "OWASP-Secrets"),
			Mask:        true,
		},
		{
			ID:          "secret.supabase",
			Title:       "Supabase service key exposed",
			Message:     "A probable Supabase service role key was added to the diff.",
			Remediation: "Keep service role keys server-side only and rotate the exposed credential.",
			Category:    "secret",
			Severity:    "critical",
			Confidence:  "medium",
			Target:      "added",
			Require:     []string{`(?i)(supabase|service_role|service-key).{0,20}[=:].{0,5}[A-Za-z0-9\-_\.]{20,}`},
			Entropy:     3.5,
			Standards:   standard("OWASP-ASVS-V8", "OWASP-Secrets"),
			Mask:        true,
		},
		{
			ID:          "secret.privatekey",
			Title:       "Private key material exposed",
			Message:     "Private key material or PEM content was added to the diff.",
			Remediation: "Remove the private key from source control and rotate any affected certificates or tokens.",
			Category:    "secret",
			Severity:    "critical",
			Confidence:  "high",
			Target:      "added",
			Require:     []string{`-----BEGIN (RSA |EC |DSA |OPENSSH |PGP )?PRIVATE KEY-----`},
			Standards:   standard("OWASP-ASVS-V8", "OWASP-Crypto"),
			Mask:        true,
		},
		{
			ID:          "secret.dotenv",
			Title:       "Environment secret file exposed",
			Message:     "A diff adds environment secret content or a private dotenv file.",
			Remediation: "Keep environment secrets outside version control and replace committed values with placeholders.",
			Category:    "secret",
			Severity:    "high",
			Confidence:  "medium",
			Target:      "added",
			Paths:       []string{"**/.env", "**/.env.*"},
			Require:     []string{`.+`},
			Standards:   standard("OWASP-ASVS-V8", "OWASP-Secrets"),
			Mask:        true,
		},
		{
			ID:          "secret.jwt",
			Title:       "JWT secret exposed",
			Message:     "A JWT secret or signing key was added to the diff.",
			Remediation: "Store signing keys in a secret manager or environment source and rotate the exposed key.",
			Category:    "secret",
			Severity:    "critical",
			Confidence:  "medium",
			Target:      "added",
			Require:     []string{`(?i)(jwt|token|signing).{0,20}(secret|key).{0,10}[=:].{0,5}[A-Za-z0-9\-_./+=]{16,}`},
			Entropy:     3.2,
			Standards:   standard("OWASP-ASVS-V3", "OWASP-ASVS-V8"),
			Mask:        true,
		},
		{
			ID:          "secret.generic",
			Title:       "High entropy secret candidate exposed",
			Message:     "A probable high-entropy application secret was added to the diff.",
			Remediation: "Review the added value, remove sensitive material from source control, and rotate it if it is valid.",
			Category:    "secret",
			Severity:    "high",
			Confidence:  "low",
			Target:      "added",
			Require:     []string{`(?i)(secret|token|apikey|api_key|passwd|password|client_secret|private_key).{0,10}[=:].{0,10}[A-Za-z0-9\-_./+=]{16,}`},
			Entropy:     3.7,
			Standards:   standard("OWASP-ASVS-V8", "OWASP-Secrets"),
			Mask:        true,
		},
		{
			ID:          "secret.pii.exposed",
			Title:       "PII or sensitive personal data exposed",
			Message:     "The diff adds a value that resembles a national identifier, credit card, or other personal data pattern.",
			Remediation: "Remove personal data from source control and replace it with tokens or synthetic data.",
			Category:    "secret",
			Severity:    "critical",
			Confidence:  "medium",
			Target:      "added",
			Require:     []string{`\b\d{3}-\d{2}-\d{4}\b|\b\d{4}[\s-]?\d{4}[\s-]?\d{4}[\s-]?\d{4}\b|\b[A-Z]{1,2}\d[A-Z\d]?\s?\d[A-Z]{2}\b|\b\d{3}-\d{3}-\d{4}\b|\b[A-Z]{2}\d{6,9}\b`},
			Standards:   standard("OWASP-ASVS-V8.3", "OWASP-Privacy"),
			Mask:        true,
		},
	}
	for _, prefix := range config.Company.Prefixes {
		clean := strings.TrimSpace(prefix)
		if clean == "" {
			continue
		}
		rules = append(rules, model.Rule{
			ID:          fmt.Sprintf("secret.company.%s", sanitize(clean)),
			Title:       "Company secret prefix exposed",
			Message:     "A configured company-specific secret prefix was added to the diff.",
			Remediation: "Remove the leaked value, rotate it, and review downstream access tied to the exposed credential.",
			Category:    "secret",
			Severity:    "critical",
			Confidence:  "medium",
			Target:      "added",
			Require:     []string{fmt.Sprintf(`%s[A-Za-z0-9\-_]{6,}`, escape(clean))},
			Standards:   standard("OWASP-ASVS-V8", "OWASP-Secrets"),
			Mask:        true,
		})
	}
	return rules
}

func sanitize(value string) string {
	builder := strings.Builder{}
	for _, char := range value {
		if char >= 'a' && char <= 'z' || char >= 'A' && char <= 'Z' || char >= '0' && char <= '9' {
			builder.WriteRune(char)
			continue
		}
		builder.WriteRune('-')
	}
	return strings.Trim(builder.String(), "-")
}

func escape(value string) string {
	replacer := strings.NewReplacer(
		"\\", `\\`,
		".", `\.`,
		"+", `\+`,
		"*", `\*`,
		"?", `\?`,
		"[", `\[`,
		"]", `\]`,
		"(", `\(`,
		")", `\)`,
		"{", `\{`,
		"}", `\}`,
		"^", `\^`,
		"$", `\$`,
		"|", `\|`,
	)
	return replacer.Replace(value)
}
