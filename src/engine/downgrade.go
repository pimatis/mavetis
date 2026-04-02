package engine

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/Pimatis/mavetis/src/model"
)

type downgradeSignal struct {
	ruleID      string
	title       string
	category    string
	severity    string
	compare     string
	key         string
	value       int64
	display     string
	message     string
	remediation string
	standards   []string
	line        model.DiffLine
}

var sameSiteFieldPattern = regexp.MustCompile(`(?i)(sameSite|samesite)\s*[:=]\s*["']?(strict|lax|none)`)
var sameSiteHeaderPattern = regexp.MustCompile(`(?i)samesite=(strict|lax|none)`)
var durationPattern = regexp.MustCompile(`(?i)(maxAge|max-age|expiresIn|expiration|expiry|ttl|idleTimeout|sessionTimeout|requestTimeout|readTimeout|writeTimeout|headerTimeout|timeout)\s*[:=]?\s*["']?([0-9]+)\s*(ms|s|m|h|d)?`)
var bcryptCallPattern = regexp.MustCompile(`(?i)bcrypt\.[A-Za-z0-9_]+\([^\n,]+,\s*(\d{1,3})\)`)
var bcryptFieldPattern = regexp.MustCompile(`(?i)(bcrypt(?:cost|rounds)?|password(?:cost|rounds)?|salt(?:rounds)?|workfactor)\s*[:=]\s*(\d{1,3})`)
var rateLimitPattern = regexp.MustCompile(`(?i)(rateLimit|rate_limit|maxAttempts|maxLoginAttempts|maxRequests|requestsPerMinute|requestsPerSecond|burst|throttle|limitPerIP|limitPerUser|tooManyRequestsAfter)\s*[:=]\s*(\d{1,7})`)
var mfaBoolPattern = regexp.MustCompile(`(?i)(requireMFA|mfaRequired|mfaEnforced|enforceMFA|requireOtp|otpRequired)\s*[:=]\s*(true|false)`)

func downgradeFindings(diff model.Diff) []model.Finding {
	findings := make([]model.Finding, 0)
	for _, file := range diff.Files {
		for _, hunk := range file.Hunks {
			added := map[string][]downgradeSignal{}
			deleted := map[string][]downgradeSignal{}
			for _, line := range hunk.Lines {
				signals := parseDowngradeSignals(line)
				for _, signal := range signals {
					key := signal.ruleID + "|" + signal.key
					if line.Kind == "added" {
						added[key] = append(added[key], signal)
					}
					if line.Kind == "deleted" {
						deleted[key] = append(deleted[key], signal)
					}
				}
			}
			for key, deletedSignals := range deleted {
				addedSignals := added[key]
				if len(addedSignals) == 0 {
					continue
				}
				finding, ok := compareDowngrade(file.Path, deletedSignals, addedSignals)
				if !ok {
					continue
				}
				findings = append(findings, finding)
			}
		}
	}
	return findings
}

func compareDowngrade(path string, deletedSignals []downgradeSignal, addedSignals []downgradeSignal) (model.Finding, bool) {
	sample := deletedSignals[0]
	deletedSignal := deletedSignals[0]
	addedSignal := addedSignals[0]
	if sample.compare == "lower" {
		for _, signal := range deletedSignals[1:] {
			if signal.value > deletedSignal.value {
				deletedSignal = signal
			}
		}
		for _, signal := range addedSignals[1:] {
			if signal.value < addedSignal.value {
				addedSignal = signal
			}
		}
		if addedSignal.value >= deletedSignal.value {
			return model.Finding{}, false
		}
	}
	if sample.compare == "higher" {
		for _, signal := range deletedSignals[1:] {
			if signal.value < deletedSignal.value {
				deletedSignal = signal
			}
		}
		for _, signal := range addedSignals[1:] {
			if signal.value > addedSignal.value {
				addedSignal = signal
			}
		}
		if addedSignal.value <= deletedSignal.value {
			return model.Finding{}, false
		}
	}
	reasons := []string{
		"deleted value: " + deletedSignal.display,
		"added value: " + addedSignal.display,
		"the diff weakens an existing security control instead of introducing a fresh control",
	}
	return model.Finding{
		ID:          identity(sample.ruleID, path, number(addedSignal.line), "added", addedSignal.display),
		RuleID:      sample.ruleID,
		Title:       sample.title,
		Category:    sample.category,
		Severity:    sample.severity,
		Confidence:  "high",
		Path:        path,
		Line:        number(addedSignal.line),
		Side:        "added",
		Message:     sample.message,
		Snippet:     strings.TrimSpace(addedSignal.line.Text),
		Remediation: sample.remediation,
		Reasons:     reasons,
		Standards:   append([]string{}, sample.standards...),
	}, true
}

func parseDowngradeSignals(line model.DiffLine) []downgradeSignal {
	signals := make([]downgradeSignal, 0, 2)
	if signal, ok := parseSameSite(line); ok {
		signals = append(signals, signal)
	}
	if signal, ok := parseCookieLifetime(line); ok {
		signals = append(signals, signal)
	}
	if signal, ok := parseBcrypt(line); ok {
		signals = append(signals, signal)
	}
	if signal, ok := parseRateLimit(line); ok {
		signals = append(signals, signal)
	}
	if signal, ok := parseTimeout(line); ok {
		signals = append(signals, signal)
	}
	if signal, ok := parseMFA(line); ok {
		signals = append(signals, signal)
	}
	return signals
}

func parseSameSite(line model.DiffLine) (downgradeSignal, bool) {
	match := sameSiteFieldPattern.FindStringSubmatch(line.Text)
	if len(match) >= 3 {
		value, ok := sameSiteRank(match[2])
		if ok {
			return downgradeSignal{ruleID: "downgrade.cookie.samesite", title: "SameSite policy weakened", category: "session", severity: "high", compare: "lower", key: "samesite", value: value, display: strings.ToLower(match[2]), message: "The diff weakens an existing SameSite cookie policy.", remediation: "Keep SameSite protections at their stronger setting unless a reviewed cross-site flow requires a documented exception.", standards: []string{"OWASP-ASVS-V3.4"}, line: line}, true
		}
	}
	match = sameSiteHeaderPattern.FindStringSubmatch(line.Text)
	if len(match) >= 2 {
		value, ok := sameSiteRank(match[1])
		if ok {
			return downgradeSignal{ruleID: "downgrade.cookie.samesite", title: "SameSite policy weakened", category: "session", severity: "high", compare: "lower", key: "samesite", value: value, display: strings.ToLower(match[1]), message: "The diff weakens an existing SameSite cookie policy.", remediation: "Keep SameSite protections at their stronger setting unless a reviewed cross-site flow requires a documented exception.", standards: []string{"OWASP-ASVS-V3.4"}, line: line}, true
		}
	}
	return downgradeSignal{}, false
}

func parseCookieLifetime(line model.DiffLine) (downgradeSignal, bool) {
	lower := strings.ToLower(line.Text)
	if strings.Contains(lower, "cache-control") || strings.Contains(lower, "cache-control:") {
		return downgradeSignal{}, false
	}
	match := durationPattern.FindStringSubmatch(line.Text)
	if len(match) < 3 {
		return downgradeSignal{}, false
	}
	key := normalizeKey(match[1])
	if key != "maxage" && key != "max-age" && key != "expiresin" && key != "expiration" && key != "expiry" && key != "ttl" {
		return downgradeSignal{}, false
	}
	if !strings.Contains(lower, "cookie") && !strings.Contains(lower, "session") && !strings.Contains(lower, "token") && !strings.Contains(lower, "auth") && !strings.Contains(lower, "set-cookie") {
		return downgradeSignal{}, false
	}
	value, ok := parseScaled(match[2], match[3])
	if !ok {
		return downgradeSignal{}, false
	}
	return downgradeSignal{ruleID: "downgrade.cookie.lifetime", title: "Cookie or token lifetime increased", category: "session", severity: "high", compare: "higher", key: "lifetime", value: value, display: match[2] + match[3], message: "The diff increases a cookie, session, or token lifetime and can prolong credential exposure.", remediation: "Keep session and token lifetimes short and extend them only with explicit security review.", standards: []string{"OWASP-ASVS-V3.3"}, line: line}, true
}

func parseBcrypt(line model.DiffLine) (downgradeSignal, bool) {
	match := bcryptCallPattern.FindStringSubmatch(line.Text)
	if len(match) >= 2 {
		value, err := strconv.ParseInt(match[1], 10, 64)
		if err == nil {
			return downgradeSignal{ruleID: "downgrade.crypto.bcrypt", title: "bcrypt cost factor reduced", category: "crypto", severity: "high", compare: "lower", key: "bcrypt", value: value, display: match[1], message: "The diff lowers a bcrypt cost factor and weakens password hashing resistance.", remediation: "Keep password hashing work factors at their stronger reviewed setting and raise them gradually when capacity allows.", standards: []string{"OWASP-ASVS-V6.2"}, line: line}, true
		}
	}
	match = bcryptFieldPattern.FindStringSubmatch(line.Text)
	if len(match) >= 3 {
		value, err := strconv.ParseInt(match[2], 10, 64)
		if err == nil {
			return downgradeSignal{ruleID: "downgrade.crypto.bcrypt", title: "bcrypt cost factor reduced", category: "crypto", severity: "high", compare: "lower", key: "bcrypt", value: value, display: match[2], message: "The diff lowers a bcrypt cost factor and weakens password hashing resistance.", remediation: "Keep password hashing work factors at their stronger reviewed setting and raise them gradually when capacity allows.", standards: []string{"OWASP-ASVS-V6.2"}, line: line}, true
		}
	}
	return downgradeSignal{}, false
}

func parseRateLimit(line model.DiffLine) (downgradeSignal, bool) {
	match := rateLimitPattern.FindStringSubmatch(line.Text)
	if len(match) < 3 {
		return downgradeSignal{}, false
	}
	value, err := strconv.ParseInt(match[2], 10, 64)
	if err != nil {
		return downgradeSignal{}, false
	}
	return downgradeSignal{ruleID: "downgrade.auth.ratelimit", title: "Rate limit threshold increased", category: "auth", severity: "high", compare: "higher", key: normalizeKey(match[1]), value: value, display: match[2], message: "The diff increases an authentication or abuse-prevention threshold and can weaken brute-force resistance.", remediation: "Keep rate limits tight on authentication and recovery flows, and review threshold increases as a security change.", standards: []string{"OWASP-ASVS-V7.2"}, line: line}, true
}

func parseTimeout(line model.DiffLine) (downgradeSignal, bool) {
	match := durationPattern.FindStringSubmatch(line.Text)
	if len(match) < 3 {
		return downgradeSignal{}, false
	}
	key := normalizeKey(match[1])
	if key == "maxage" || key == "max-age" || key == "expiresin" || key == "expiration" || key == "expiry" || key == "ttl" {
		return downgradeSignal{}, false
	}
	lower := strings.ToLower(line.Text)
	if strings.Contains(lower, "cache") {
		return downgradeSignal{}, false
	}
	value, ok := parseScaled(match[2], match[3])
	if !ok {
		return downgradeSignal{}, false
	}
	return downgradeSignal{ruleID: "downgrade.timeout", title: "Security-relevant timeout increased", category: "config", severity: "medium", compare: "higher", key: key, value: value, display: match[2] + match[3], message: "The diff increases a timeout value and can extend exposure to stalled or long-lived security-sensitive operations.", remediation: "Review timeout increases carefully and keep authentication, session, and service timeouts aligned with hardened production expectations.", standards: []string{"OWASP-ASVS-V3.3", "OWASP-ASVS-V9.1"}, line: line}, true
}

func parseMFA(line model.DiffLine) (downgradeSignal, bool) {
	match := mfaBoolPattern.FindStringSubmatch(line.Text)
	if len(match) >= 3 {
		value := int64(0)
		if strings.EqualFold(match[2], "true") {
			value = 1
		}
		return downgradeSignal{ruleID: "downgrade.auth.mfa", title: "MFA requirement weakened", category: "auth", severity: "critical", compare: "lower", key: "mfa", value: value, display: strings.ToLower(match[2]), message: "The diff weakens a previously enforced MFA requirement.", remediation: "Keep MFA mandatory on the protected flow and route exceptions through explicit reviewed policy.", standards: []string{"OWASP-ASVS-V3.1"}, line: line}, true
	}
	lower := strings.ToLower(line.Text)
	if strings.Contains(lower, "skipmfa") || strings.Contains(lower, "disablemfa") || strings.Contains(lower, "mfaoptional") || strings.Contains(lower, "optionalmfa") {
		return downgradeSignal{ruleID: "downgrade.auth.mfa", title: "MFA requirement weakened", category: "auth", severity: "critical", compare: "lower", key: "mfa", value: 0, display: "optional", message: "The diff weakens a previously enforced MFA requirement.", remediation: "Keep MFA mandatory on the protected flow and route exceptions through explicit reviewed policy.", standards: []string{"OWASP-ASVS-V3.1"}, line: line}, true
	}
	return downgradeSignal{}, false
}

func sameSiteRank(value string) (int64, bool) {
	lower := strings.ToLower(value)
	if lower == "strict" {
		return 3, true
	}
	if lower == "lax" {
		return 2, true
	}
	if lower == "none" {
		return 1, true
	}
	return 0, false
}

func parseScaled(raw string, unit string) (int64, bool) {
	value, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return 0, false
	}
	lower := strings.ToLower(unit)
	if lower == "ms" || lower == "" {
		return value, true
	}
	if lower == "s" {
		return value * 1000, true
	}
	if lower == "m" {
		return value * 60 * 1000, true
	}
	if lower == "h" {
		return value * 60 * 60 * 1000, true
	}
	if lower == "d" {
		return value * 24 * 60 * 60 * 1000, true
	}
	return 0, false
}

func normalizeKey(value string) string {
	lower := strings.ToLower(value)
	lower = strings.ReplaceAll(lower, "_", "")
	lower = strings.ReplaceAll(lower, "-", "")
	lower = strings.ReplaceAll(lower, ".", "")
	return lower
}
