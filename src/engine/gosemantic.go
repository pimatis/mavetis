package engine

import (
	"strings"

	"github.com/Pimatis/mavetis/src/analyze"
	"github.com/Pimatis/mavetis/src/model"
)

func goSemanticFindings(diff model.Diff) []model.Finding {
	findings := make([]model.Finding, 0)
	for _, file := range diff.Files {
		if analyze.Language(file.Path) != "go" {
			continue
		}
		if analyze.ReviewArtifact(file.Path) {
			continue
		}
		for _, hunk := range file.Hunks {
			flows := analyze.GoFlows(join(hunk))
			for _, flow := range flows {
				if strings.EqualFold(flow.Sink, "http.Get") {
					findings = append(findings, syntheticFinding("semantic.go.ssrf", "Go AST found tainted flow into http.Get", "ssrf", "critical", file.Path, hunk, "Go AST flow analysis found request-derived data reaching http.Get inside the diff hunk.", "Validate remote targets with an allowlist and reject private, loopback, and metadata destinations.", "Go AST tracked request-derived values through assignments", "http.Get consumed the tainted value"))
				}
				if strings.EqualFold(flow.Sink, "exec.Command") {
					findings = append(findings, syntheticFinding("semantic.go.exec", "Go AST found tainted flow into exec.Command", "injection", "critical", file.Path, hunk, "Go AST flow analysis found request-derived data reaching exec.Command inside the diff hunk.", "Avoid command execution on untrusted input and strictly separate arguments from user-controlled data.", "Go AST tracked request-derived values through assignments", "exec.Command consumed the tainted value"))
				}
				if strings.EqualFold(flow.Sink, "os.Open") {
					findings = append(findings, syntheticFinding("semantic.go.path", "Go AST found tainted flow into os.Open", "file", "high", file.Path, hunk, "Go AST flow analysis found request-derived data reaching os.Open inside the diff hunk.", "Normalize and bound filesystem paths before opening user-influenced locations.", "Go AST tracked request-derived values through assignments", "os.Open consumed the tainted value"))
				}
				if strings.EqualFold(flow.Sink, "template.New") {
					findings = append(findings, syntheticFinding("semantic.go.template", "Go AST found tainted flow into template construction", "template", "high", file.Path, hunk, "Go AST flow analysis found request-derived data reaching template construction inside the diff hunk.", "Keep template definitions static and avoid constructing templates from user-controlled input.", "Go AST tracked request-derived values through assignments", "template construction consumed the tainted value"))
				}
			}
		}
	}
	return findings
}

func syntheticFinding(ruleID string, title string, category string, severity string, path string, hunk model.DiffHunk, message string, remediation string, reasons ...string) model.Finding {
	line := 1
	if len(hunk.Lines) != 0 {
		line = number(hunk.Lines[0])
	}
	return model.Finding{
		ID:          identity(ruleID, path, line, "added", title),
		RuleID:      ruleID,
		Title:       title,
		Category:    category,
		Severity:    severity,
		Confidence:  "medium",
		Path:        path,
		Line:        line,
		Side:        "added",
		Message:     message,
		Snippet:     "semantic correlation inside diff hunk",
		Remediation: remediation,
		Reasons:     reasons,
		Standards:   []string{"OWASP-ASVS", "OWASP-Secure-Coding"},
	}
}
