package engine

import (
	"strings"

	"github.com/Pimatis/mavetis/src/analyze"
	"github.com/Pimatis/mavetis/src/model"
)

func signatureFindings(diff model.Diff) []model.Finding {
	findings := make([]model.Finding, 0)
	for _, file := range diff.Files {
		if !analyze.Executable(file.Path) {
			continue
		}
		for _, hunk := range file.Hunks {
			text := strings.ToLower(join(hunk))
			if strings.Contains(text, "header[\"alg\"]") || strings.Contains(text, "token.header.alg") {
				findings = append(findings, syntheticFinding("crypto.alg.trusted", "Verification algorithm taken from untrusted token header", "crypto", "critical", file.Path, hunk, "The diff appears to trust the algorithm value from an untrusted token header.", "Enforce a fixed allowlist of accepted algorithms and never trust the token header to select verification behavior.", "algorithm selection logic appeared in the same hunk", "the source appears to come from token-controlled header data"))
			}
			if strings.Contains(text, "header[\"kid\"]") || strings.Contains(text, "token.header.kid") {
				findings = append(findings, syntheticFinding("crypto.kid.trusted", "Key selection appears to trust unvalidated kid header data", "crypto", "high", file.Path, hunk, "The diff appears to resolve verification keys directly from untrusted kid header content.", "Resolve kid values only against a strict trusted key set and reject unrecognized or unsafe key selectors.", "kid-based key selection logic appeared in the same hunk", "the source appears to come from token-controlled header data"))
			}
			if strings.Contains(text, "jku") || strings.Contains(text, "jwk") || strings.Contains(text, "x5u") {
				if strings.Contains(text, "http.get") || strings.Contains(text, "fetch(") || strings.Contains(text, "axios") {
					findings = append(findings, syntheticFinding("crypto.jku.remote", "Verification keys fetched from token-controlled metadata", "crypto", "critical", file.Path, hunk, "The diff appears to fetch JKU, JWK, or X5U material from token-controlled metadata.", "Do not fetch verification keys from token-controlled URLs. Use trusted preconfigured key sources only.", "JWK or X5U metadata appeared in the same hunk", "network retrieval logic appeared near the metadata usage"))
				}
			}
			if strings.Contains(text, "publickey") && strings.Contains(text, "hmac") {
				findings = append(findings, syntheticFinding("crypto.key.confusion", "Potential HMAC and public-key confusion introduced", "crypto", "critical", file.Path, hunk, "The diff appears to mix public-key material into an HMAC or symmetric verification path.", "Keep symmetric and asymmetric verification paths fully separated and use explicit key typing.", "public-key and HMAC terms appeared together in the same hunk", "this pattern can indicate algorithm or key confusion"))
			}
		}
	}
	return findings
}
