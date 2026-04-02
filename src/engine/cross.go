package engine

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"

	"github.com/Pimatis/mavetis/src/analyze"
	"github.com/Pimatis/mavetis/src/model"
)

type crossState struct {
	guardDeleted  bool
	routeAdded    bool
	scopeDeleted  bool
	lookupAdded   bool
	verifyDeleted bool
	decodeAdded   bool
}

func crossFindings(diff model.Diff) []model.Finding {
	state := crossState{}
	for _, file := range diff.Files {
		if !analyze.Executable(file.Path) {
			continue
		}
		for _, hunk := range file.Hunks {
			for _, line := range hunk.Lines {
				lower := strings.ToLower(line.Text)
				if line.Kind == "deleted" && crossGuard(lower) {
					state.guardDeleted = true
				}
				if line.Kind == "added" && crossRoute(lower) {
					state.routeAdded = true
				}
				if line.Kind == "deleted" && crossScope(lower) {
					state.scopeDeleted = true
				}
				if line.Kind == "added" && crossLookup(lower) {
					state.lookupAdded = true
				}
				if line.Kind == "deleted" && crossVerify(lower) {
					state.verifyDeleted = true
				}
				if line.Kind == "added" && crossDecode(lower) {
					state.decodeAdded = true
				}
			}
		}
	}
	findings := make([]model.Finding, 0)
	if state.guardDeleted && state.routeAdded {
		findings = append(findings, crossFinding("branch.guard.regression", "Branch diff removes a guard while exposing a route", "authorization", "critical", "The diff removes an authentication or authorization guard while also adding or changing a route or handler.", "Review the branch as a whole and restore explicit route protection before merge."))
	}
	if state.scopeDeleted && state.lookupAdded {
		findings = append(findings, crossFinding("branch.scope.regression", "Branch diff weakens resource scoping while adding direct lookups", "authorization", "critical", "The diff removes tenant or owner scoping while also adding direct object lookup behavior.", "Restore tenant or ownership filters and verify the new lookup path stays authorization-bound."))
	}
	if state.verifyDeleted && state.decodeAdded {
		findings = append(findings, crossFinding("branch.token.regression", "Branch diff removes verification while adding decode-like token usage", "token", "critical", "The diff removes a verification path and introduces decode or accept-invalid token behavior in the same branch.", "Restore verification and reject any decode-only or invalid-token acceptance flow before merge."))
	}
	return findings
}

func crossGuard(text string) bool {
	return strings.Contains(text, "authorize") || strings.Contains(text, "requireauth") || strings.Contains(text, "requirerole") || strings.Contains(text, "permission")
}

func crossRoute(text string) bool {
	return strings.Contains(text, "router.") || strings.Contains(text, "app.") || strings.Contains(text, "handlefunc") || strings.Contains(text, "get(") || strings.Contains(text, "post(")
}

func crossScope(text string) bool {
	return strings.Contains(text, "tenant") || strings.Contains(text, "owner") || strings.Contains(text, "workspace") || strings.Contains(text, "org_id")
}

func crossLookup(text string) bool {
	return strings.Contains(text, "find") || strings.Contains(text, "first") || strings.Contains(text, "get") || strings.Contains(text, "query")
}

func crossVerify(text string) bool {
	return strings.Contains(text, "verify") || strings.Contains(text, "parsewithclaims") || strings.Contains(text, "hmac.equal")
}

func crossDecode(text string) bool {
	return strings.Contains(text, "decode") || strings.Contains(text, "acceptinvalid") || strings.Contains(text, "allowunsigned")
}

func crossFinding(ruleID string, title string, category string, severity string, message string, remediation string) model.Finding {
	sum := sha256.Sum256([]byte(ruleID + title + message))
	return model.Finding{
		ID:          hex.EncodeToString(sum[:8]),
		RuleID:      ruleID,
		Title:       title,
		Category:    category,
		Severity:    severity,
		Confidence:  "medium",
		Path:        "<branch>",
		Line:        1,
		Side:        "branch",
		Message:     message,
		Snippet:     "branch-level regression correlation",
		Remediation: remediation,
		Reasons:     []string{"multiple files in the branch diff combine into a stronger regression signal"},
		Standards:   []string{"OWASP-ASVS", "OWASP-Authorization"},
	}
}
