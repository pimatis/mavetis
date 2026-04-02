package engine

import (
	"regexp"
	"strings"

	"github.com/Pimatis/mavetis/src/analyze"
	"github.com/Pimatis/mavetis/src/model"
)

var nonceAssign = regexp.MustCompile(`(?i)\b([A-Za-z_][A-Za-z0-9_]*)\b\s*(:=|=).*$`)
var cryptoUse = regexp.MustCompile(`(?i)(seal\(|open\(|encrypt\(|decrypt\(|gcm\.|chacha20|ctr\(|cbc\()`)

func nonceFindings(diff model.Diff) []model.Finding {
	findings := make([]model.Finding, 0)
	for _, file := range diff.Files {
		if !analyze.Executable(file.Path) {
			continue
		}
		for _, hunk := range file.Hunks {
			vars := nonceVars(hunk)
			if len(vars) == 0 {
				continue
			}
			for name, count := range vars {
				if count < 2 {
					continue
				}
				findings = append(findings, syntheticFinding("crypto.nonce.reuse", "Nonce or IV appears reused inside the diff hunk", "crypto", "critical", file.Path, hunk, "The same nonce or IV symbol appears in multiple cryptographic operations inside the same diff hunk.", "Use a fresh nonce or IV for each encryption operation and never reuse nonce material under the same key.", "a nonce or IV symbol was assigned in the hunk", "the same symbol was reused across multiple cryptographic operations", "matched symbol: "+name))
			}
		}
	}
	return findings
}

func nonceVars(hunk model.DiffHunk) map[string]int {
	assigned := map[string]struct{}{}
	counts := map[string]int{}
	for _, line := range hunk.Lines {
		if line.Kind != "added" {
			continue
		}
		match := nonceAssign.FindStringSubmatch(line.Text)
		if len(match) >= 2 {
			name := strings.ToLower(match[1])
			if strings.Contains(name, "nonce") || strings.Contains(name, "iv") {
				assigned[name] = struct{}{}
			}
			if strings.Contains(strings.ToLower(line.Text), "nonce") || strings.Contains(strings.ToLower(line.Text), " iv") {
				assigned[name] = struct{}{}
			}
		}
	}
	for _, line := range hunk.Lines {
		if line.Kind != "added" {
			continue
		}
		if !cryptoUse.MatchString(strings.ToLower(line.Text)) {
			continue
		}
		lower := strings.ToLower(line.Text)
		for name := range assigned {
			if strings.Contains(lower, name) {
				counts[name]++
			}
		}
	}
	return counts
}
