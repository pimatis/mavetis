package engine

import (
	"crypto/sha256"
	"encoding/hex"
	"regexp"
	"strings"

	"github.com/Pimatis/mavetis/src/analyze"
	"github.com/Pimatis/mavetis/src/model"
)

var fetchSink = regexp.MustCompile(`(?i)(fetch\(|http\.Get\(|http\.Post\(|axios\.|requests\.(get|post)|urlopen\()`)
var execSink = regexp.MustCompile(`(?i)(exec\(|spawn\(|system\(|subprocess\.|sh\s+-c|bash\s+-c)`)
var pathSink = regexp.MustCompile(`(?i)(ReadFile|WriteFile|os\.Open|os\.OpenFile|filepath\.Join|path\.Join|os\.Create)`)
var sqlSink = regexp.MustCompile(`(?i)(select\b.+\bfrom\b|insert\b.+\binto\b|update\b.+\bset\b|delete\b.+\bfrom\b|fmt\.sprint|sprintf|\+)`)
var lookupSink = regexp.MustCompile(`(?i)(find|first|getby|get\(|load|select|delete|update)`)
var templateSink = regexp.MustCompile(`(?i)(template\.New|template\.Parse|ParseFiles\(|ParseGlob\(|ParseFS\(|render_template_string|Handlebars\.compile|ejs\.render|eval\(|new Function\()`)

func semanticFindings(diff model.Diff) []model.Finding {
	findings := make([]model.Finding, 0)
	for _, file := range diff.Files {
		if !analyze.Executable(file.Path) {
			continue
		}
		if analyze.ReviewArtifact(file.Path) {
			continue
		}
		for _, hunk := range file.Hunks {
			text := join(hunk)
			vars := analyze.Tainted(text)
			guarded := analyze.Guarded(text)
			calls := goCalls(file.Path, text)
			steps := analyze.Track(lines(hunk))
			stepIndex := 0
			for _, line := range hunk.Lines {
				if line.Kind != "added" {
					continue
				}
				current := vars
				if stepIndex < len(steps) {
					current = mergeVars(vars, steps[stepIndex].Taints)
					stepIndex++
				}
				findings = append(findings, semanticLine(file.Path, line, text, current, guarded, calls)...)
			}
		}
	}
	return findings
}

func semanticLine(path string, line model.DiffLine, hunkText string, vars []string, guarded bool, calls []string) []model.Finding {
	findings := make([]model.Finding, 0)
	if fetchSink.MatchString(line.Text) && analyze.TaintedUse(line.Text, vars) && !guarded {
		reasons := append([]string{}, "request-controlled value was detected in the same hunk", "outbound network sink uses that value without nearby mitigation")
		reasons = append(reasons, astReason(calls, "http.Get", "Go AST confirmed an outbound HTTP call in the hunk")...)
		findings = append(findings, semanticFinding("semantic.ssrf.flow", "Request-controlled URL reaches outbound fetch", "ssrf", "critical", path, line, "Outbound fetch appears to use a request-controlled value without nearby allowlist or host validation.", "Validate remote targets against an allowlist and block loopback, private, and metadata destinations.", reasons...))
	}
	if execSink.MatchString(line.Text) && analyze.TaintedUse(line.Text, vars) && !guarded {
		reasons := append([]string{}, "request-controlled value was detected in the same hunk", "command execution sink uses that value")
		reasons = append(reasons, astReason(calls, "exec.Command", "Go AST confirmed command execution in the hunk")...)
		findings = append(findings, semanticFinding("semantic.command.flow", "Request-controlled value reaches command execution", "injection", "critical", path, line, "Command execution appears to consume request-controlled input in the same hunk.", "Remove shell execution or separate arguments from untrusted input with strict allowlists.", reasons...))
	}
	if pathSink.MatchString(line.Text) && analyze.TaintedUse(line.Text, vars) && !guarded {
		reasons := append([]string{}, "request-controlled path-like value was detected in the same hunk", "filesystem sink uses that value without nearby path normalization")
		reasons = append(reasons, astReason(calls, "os.Open", "Go AST confirmed filesystem access in the hunk")...)
		findings = append(findings, semanticFinding("semantic.traversal.flow", "Request-controlled value reaches filesystem access", "file", "high", path, line, "Filesystem access appears to consume request-controlled path fragments without nearby normalization controls.", "Normalize with clean and boundary checks before opening or writing paths derived from user input.", reasons...))
	}
	if sqlSink.MatchString(line.Text) && analyze.TaintedUse(line.Text, vars) {
		findings = append(findings, semanticFinding("semantic.sql.flow", "Request-controlled value participates in SQL construction", "injection", "high", path, line, "SQL construction appears to mix request-controlled input into a dynamic query path.", "Use parameterized queries and avoid building SQL text with request-controlled values.", "request-controlled value was detected in the same hunk", "query-building sink appears in the same line"))
	}
	if lookupSink.MatchString(line.Text) && !fetchSink.MatchString(line.Text) && analyze.TaintedUse(line.Text, vars) && !guarded {
		findings = append(findings, semanticFinding("semantic.idor.flow", "Request-controlled identifier reaches object lookup", "authorization", "high", path, line, "A request-controlled identifier appears to flow into a lookup or mutation without nearby ownership or permission checks.", "Bind lookups to tenant, owner, or permission context before loading or mutating resources.", "request-controlled identifier was detected in the same hunk", "resource lookup or mutation occurs without nearby guard signals"))
	}
	if templateSink.MatchString(line.Text) && analyze.TaintedUse(hunkText, vars) {
		reasons := append([]string{}, "template or eval sink appears in the hunk", "request-controlled data is present near the sink")
		reasons = append(reasons, astReason(calls, "template.New", "Go AST confirmed template construction in the hunk")...)
		findings = append(findings, semanticFinding("semantic.template.flow", "Request-controlled template or eval path introduced", "template", "high", path, line, "Template parsing or dynamic evaluation appears near request-controlled content in the same hunk.", "Keep template source static and remove eval-like execution paths from untrusted input flows.", reasons...))
	}
	return findings
}

func lines(hunk model.DiffHunk) []string {
	values := make([]string, 0, len(hunk.Lines))
	for _, line := range hunk.Lines {
		if line.Kind != "added" {
			continue
		}
		values = append(values, line.Text)
	}
	return values
}

func mergeVars(left []string, right []string) []string {
	set := map[string]struct{}{}
	values := make([]string, 0, len(left)+len(right))
	for _, item := range left {
		if _, ok := set[item]; ok {
			continue
		}
		set[item] = struct{}{}
		values = append(values, item)
	}
	for _, item := range right {
		if _, ok := set[item]; ok {
			continue
		}
		set[item] = struct{}{}
		values = append(values, item)
	}
	return values
}

func goCalls(path string, text string) []string {
	if analyze.Language(path) != "go" {
		return nil
	}
	return analyze.GoCalls(text)
}

func astReason(calls []string, want string, reason string) []string {
	for _, item := range calls {
		if item != want {
			continue
		}
		return []string{reason}
	}
	return nil
}

func semanticFinding(ruleID string, title string, category string, severity string, path string, line model.DiffLine, message string, remediation string, reasons ...string) model.Finding {
	snippet := line.Text
	sum := sha256.Sum256([]byte(ruleID + "|" + path + "|" + line.Text))
	return model.Finding{
		ID:          hex.EncodeToString(sum[:8]),
		RuleID:      ruleID,
		Title:       title,
		Category:    category,
		Severity:    severity,
		Confidence:  "medium",
		Path:        path,
		Line:        number(line),
		Side:        line.Kind,
		Message:     message,
		Snippet:     strings.TrimSpace(snippet),
		Remediation: remediation,
		Reasons:     reasons,
		Standards:   []string{"OWASP-ASVS", "OWASP-Secure-Coding"},
	}
}
