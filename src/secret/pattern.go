package secret

import "regexp"

type pattern struct {
	id          string
	title       string
	message     string
	severity    string
	confidence  string
	remediation string
	standards   []string
	re          *regexp.Regexp
	group       int
	entropy     float64
	path        *regexp.Regexp
}

func patterns() []pattern {
	return []pattern{
		{
			id:          "secret.scan.aws.access",
			title:       "AWS access key exposed",
			message:     "A value matching an AWS access key was found in local files.",
			severity:    "critical",
			confidence:  "high",
			remediation: "Remove the key from source control, rotate it, and load credentials from a managed secret source.",
			standards:   []string{"OWASP-ASVS-V8", "OWASP-Secrets", "CWE-798"},
			re:          regexp.MustCompile(`\b(AKIA|ASIA)[0-9A-Z]{16}\b`),
		},
		{
			id:          "secret.scan.github.token",
			title:       "GitHub token exposed",
			message:     "A value matching a GitHub token was found in local files.",
			severity:    "critical",
			confidence:  "high",
			remediation: "Revoke the token, remove it from source control, and use scoped CI or runtime secrets instead.",
			standards:   []string{"OWASP-ASVS-V8", "OWASP-Secrets", "CWE-798"},
			re:          regexp.MustCompile(`\b(gh[pousr]_[A-Za-z0-9_]{30,}|github_pat_[A-Za-z0-9_]{20,}_[A-Za-z0-9_]{20,})\b`),
			group:       1,
			entropy:     3.5,
		},
		{
			id:          "secret.scan.slack.token",
			title:       "Slack token exposed",
			message:     "A value matching a Slack token was found in local files.",
			severity:    "critical",
			confidence:  "high",
			remediation: "Revoke the token in Slack, remove it from source control, and inject it only at runtime.",
			standards:   []string{"OWASP-ASVS-V8", "OWASP-Secrets", "CWE-798"},
			re:          regexp.MustCompile(`\bxox[baprs]-[A-Za-z0-9-]{10,}\b`),
			entropy:     3.2,
		},
		{
			id:          "secret.scan.stripe.key",
			title:       "Stripe secret key exposed",
			message:     "A value matching a Stripe secret key was found in local files.",
			severity:    "critical",
			confidence:  "high",
			remediation: "Rotate the Stripe key, remove it from source control, and load it from a secure runtime source.",
			standards:   []string{"OWASP-ASVS-V8", "OWASP-Secrets", "CWE-798"},
			re:          regexp.MustCompile(`\bsk_(live|test)_[A-Za-z0-9]{16,}\b`),
			entropy:     3.2,
		},
		{
			id:          "secret.scan.google.api",
			title:       "Google API key exposed",
			message:     "A value matching a Google API key was found in local files.",
			severity:    "high",
			confidence:  "high",
			remediation: "Restrict and rotate the API key, remove it from source control, and load it from a managed secret source.",
			standards:   []string{"OWASP-ASVS-V8", "OWASP-Secrets", "CWE-798"},
			re:          regexp.MustCompile(`\bAIza[0-9A-Za-z\-_]{35}\b`),
			entropy:     3.2,
		},
		{
			id:          "secret.scan.privatekey",
			title:       "Private key material exposed",
			message:     "Private key material was found in local files.",
			severity:    "critical",
			confidence:  "high",
			remediation: "Remove the private key from source control and rotate every certificate, SSH key, or token derived from it.",
			standards:   []string{"OWASP-ASVS-V8", "OWASP-Secrets", "CWE-321"},
			re:          regexp.MustCompile(`-----BEGIN (RSA |EC |DSA |OPENSSH |PGP )?PRIVATE KEY-----`),
		},
		{
			id:          "secret.scan.dotenv",
			title:       "Dotenv secret value exposed",
			message:     "A private environment file contains a likely secret assignment.",
			severity:    "high",
			confidence:  "medium",
			remediation: "Remove the dotenv secret, rotate the value, and keep environment secrets outside source control.",
			standards:   []string{"OWASP-ASVS-V8", "OWASP-Secrets", "CWE-798"},
			re:          regexp.MustCompile(`(?i)^\s*([A-Z0-9_]*(SECRET|TOKEN|KEY|PASSWORD|PASSWD|PRIVATE)[A-Z0-9_]*)\s*=\s*['\"]?([^'\"\s#]{12,})`),
			group:       3,
			entropy:     3.0,
			path:        regexp.MustCompile(`(^|/)\.env(\.|$)`),
		},
		{
			id:          "secret.scan.generic",
			title:       "High entropy secret candidate exposed",
			message:     "A high-entropy value assigned to a secret-like name was found in local files.",
			severity:    "high",
			confidence:  "medium",
			remediation: "Review the value, remove it from source control if sensitive, and rotate it before reuse.",
			standards:   []string{"OWASP-ASVS-V8", "OWASP-Secrets", "CWE-798"},
			re:          regexp.MustCompile(`(?i)(secret|token|api[_-]?key|client[_-]?secret|private[_-]?key|password|passwd)\s*[:=]\s*['\"]?([A-Za-z0-9_./+=\-]{16,})`),
			group:       2,
			entropy:     3.6,
		},
	}
}
