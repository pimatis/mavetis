package engine

import "github.com/Pimatis/mavetis/src/model"

func MatrixInfos(rules []model.Rule) []model.RuleInfo {
	items := infos(rules)
	items = append(items, syntheticInfosForProfile("")...)
	return items
}

func infos(rules []model.Rule) []model.RuleInfo {
	items := make([]model.RuleInfo, 0, len(rules))
	for _, rule := range rules {
		items = append(items, model.RuleInfo{ID: rule.ID, Title: rule.Title, Category: rule.Category, Severity: rule.Severity, Standards: rule.Standards})
	}
	return items
}

func snapshotInfos(snapshots []model.Snapshot) []model.RuleInfo {
	items := make([]model.RuleInfo, 0, len(snapshots))
	for _, snapshot := range snapshots {
		items = append(items, model.RuleInfo{ID: snapshot.ID, Title: "Repository security snapshot regressed", Category: snapshot.Category, Severity: snapshot.Severity, Standards: snapshot.Standards})
	}
	return items
}

func syntheticInfos() []model.RuleInfo {
	return []model.RuleInfo{
		{ID: "semantic.ssrf.flow", Title: "Request-controlled URL reaches outbound fetch", Category: "ssrf", Severity: "critical", Standards: []string{"OWASP-ASVS-V4.3"}},
		{ID: "semantic.command.flow", Title: "Request-controlled value reaches command execution", Category: "injection", Severity: "critical", Standards: []string{"OWASP-ASVS-V1.2"}},
		{ID: "semantic.traversal.flow", Title: "Request-controlled value reaches filesystem access", Category: "file", Severity: "high", Standards: []string{"OWASP-ASVS-V5.4"}},
		{ID: "semantic.sql.flow", Title: "Request-controlled value participates in SQL construction", Category: "injection", Severity: "high", Standards: []string{"OWASP-ASVS-V1.2"}},
		{ID: "semantic.idor.flow", Title: "Request-controlled identifier reaches object lookup", Category: "authorization", Severity: "high", Standards: []string{"OWASP-ASVS-V4.1"}},
		{ID: "semantic.template.flow", Title: "Request-controlled template or eval path introduced", Category: "template", Severity: "high", Standards: []string{"OWASP-ASVS-V1.2"}},
		{ID: "semantic.go.ssrf", Title: "Go AST found tainted flow into http.Get", Category: "ssrf", Severity: "critical", Standards: []string{"OWASP-ASVS-V4.3"}},
		{ID: "semantic.go.exec", Title: "Go AST found tainted flow into exec.Command", Category: "injection", Severity: "critical", Standards: []string{"OWASP-ASVS-V1.2"}},
		{ID: "semantic.go.path", Title: "Go AST found tainted flow into os.Open", Category: "file", Severity: "high", Standards: []string{"OWASP-ASVS-V5.4"}},
		{ID: "semantic.go.template", Title: "Go AST found tainted flow into template construction", Category: "template", Severity: "high", Standards: []string{"OWASP-ASVS-V1.2"}},
		{ID: "crypto.nonce.reuse", Title: "Nonce or IV appears reused inside the diff hunk", Category: "crypto", Severity: "critical", Standards: []string{"OWASP-ASVS-V6.2"}},
		{ID: "supply.registry.public", Title: "Registry source points to a broad public index", Category: "supply", Severity: "high", Standards: []string{"OWASP-ASVS-V14.2"}},
		{ID: "supply.version.floating", Title: "Floating dependency version introduced", Category: "supply", Severity: "medium", Standards: []string{"OWASP-ASVS-V14.2"}},
		{ID: "supply.replace.remote", Title: "Remote Go module replacement introduced", Category: "supply", Severity: "high", Standards: []string{"OWASP-ASVS-V14.2"}},
		{ID: "supply.typosquat", Title: "Package name resembles a likely typosquat", Category: "supply", Severity: "high", Standards: []string{"OWASP-ASVS-V14.2"}},
		{ID: "crypto.alg.trusted", Title: "Verification algorithm taken from untrusted token header", Category: "crypto", Severity: "critical", Standards: []string{"OWASP-ASVS-V3.5", "OWASP-ASVS-V6.2"}},
		{ID: "crypto.kid.trusted", Title: "Key selection appears to trust unvalidated kid header data", Category: "crypto", Severity: "high", Standards: []string{"OWASP-ASVS-V3.5", "OWASP-ASVS-V6.2"}},
		{ID: "crypto.jku.remote", Title: "Verification keys fetched from token-controlled metadata", Category: "crypto", Severity: "critical", Standards: []string{"OWASP-ASVS-V3.5", "OWASP-ASVS-V6.2"}},
		{ID: "crypto.key.confusion", Title: "Potential HMAC and public-key confusion introduced", Category: "crypto", Severity: "critical", Standards: []string{"OWASP-ASVS-V3.5", "OWASP-ASVS-V6.2"}},
		{ID: "downgrade.cookie.samesite", Title: "SameSite policy weakened", Category: "session", Severity: "high", Standards: []string{"OWASP-ASVS-V3.4"}},
		{ID: "downgrade.cookie.lifetime", Title: "Cookie or token lifetime increased", Category: "session", Severity: "high", Standards: []string{"OWASP-ASVS-V3.3"}},
		{ID: "downgrade.crypto.bcrypt", Title: "bcrypt cost factor reduced", Category: "crypto", Severity: "high", Standards: []string{"OWASP-ASVS-V6.2"}},
		{ID: "downgrade.auth.ratelimit", Title: "Rate limit threshold increased", Category: "auth", Severity: "high", Standards: []string{"OWASP-ASVS-V7.2"}},
		{ID: "downgrade.timeout", Title: "Security-relevant timeout increased", Category: "config", Severity: "medium", Standards: []string{"OWASP-ASVS-V3.3", "OWASP-ASVS-V9.1"}},
		{ID: "downgrade.auth.mfa", Title: "MFA requirement weakened", Category: "auth", Severity: "critical", Standards: []string{"OWASP-ASVS-V3.1"}},
		{ID: "supply.lifecycle.dependency", Title: "Dependency addition paired with lifecycle script", Category: "supply", Severity: "critical", Standards: []string{"OWASP-ASVS-V14.2"}},
		{ID: "supply.lock.missing", Title: "Dependency manifest changed without lockfile update", Category: "supply", Severity: "high", Standards: []string{"OWASP-ASVS-V14.2"}},
		{ID: "supply.registry.drift", Title: "Private registry drifted to public registry", Category: "supply", Severity: "critical", Standards: []string{"OWASP-ASVS-V14.2"}},
		{ID: "supply.registry.untrusted", Title: "Registry is outside trusted allowlist", Category: "supply", Severity: "high", Standards: []string{"OWASP-ASVS-V14.2"}},
		{ID: "supply.package.denied", Title: "Dependency is denied by trust policy", Category: "supply", Severity: "critical", Standards: []string{"OWASP-ASVS-V14.2"}},
		{ID: "supply.package.untrusted", Title: "Dependency falls outside package allowlist", Category: "supply", Severity: "high", Standards: []string{"OWASP-ASVS-V14.2"}},
		{ID: "intent.mismatch", Title: "Security intent no longer matches implementation", Category: "intent", Severity: "high", Standards: []string{"OWASP-ASVS-V1.14"}},
		{ID: "branch.guard.regression", Title: "Branch diff removes a guard while exposing a route", Category: "authorization", Severity: "critical", Standards: []string{"OWASP-ASVS-V4.1"}},
		{ID: "branch.scope.regression", Title: "Branch diff weakens resource scoping while adding direct lookups", Category: "authorization", Severity: "critical", Standards: []string{"OWASP-ASVS-V4.1"}},
		{ID: "branch.token.regression", Title: "Branch diff removes verification while adding decode-like token usage", Category: "token", Severity: "critical", Standards: []string{"OWASP-ASVS-V3.5"}},
	}
}

func appendFindings(target *[]model.Finding, seen map[string]struct{}, additions []model.Finding, config model.Config, zoneCache map[string]zoneMatch) {
	for _, finding := range additions {
		if !allowSynthetic(config.Profile, finding) {
			continue
		}
		if finding.Path != "" && finding.Path != "<branch>" {
			zone := resolveZone(finding.Path, config, zoneCache)
			finding = applyPolicy(finding, zone, config.FailOn)
		}
		if model.SeverityRank(finding.Severity) < model.SeverityRank(config.Severity) {
			continue
		}
		if _, ok := seen[finding.ID]; ok {
			continue
		}
		seen[finding.ID] = struct{}{}
		*target = append(*target, finding)
	}
}
