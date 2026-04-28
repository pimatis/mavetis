package rule

import "github.com/Pimatis/mavetis/src/model"

var controlMap = map[string][]string{
	"auth.middleware.deleted":         {"OWASP-ASVS-V3.1"},
	"session.storage.local":           {"OWASP-ASVS-V3.4"},
	"token.decode.only":               {"OWASP-ASVS-V3.5"},
	"auth.bypass.debug":               {"OWASP-ASVS-V3.1"},
	"auth.mfa.disabled":               {"OWASP-ASVS-V3.1"},
	"auth.mfa.deleted":                {"OWASP-ASVS-V3.1"},
	"auth.ratelimit.deleted":          {"OWASP-ASVS-V7.2"},
	"boundary.admin.public":           {"OWASP-ASVS-V4.1"},
	"boundary.ui.auth":                {"OWASP-ASVS-V4.1"},
	"boundary.privileged.public":      {"OWASP-ASVS-V4.1"},
	"session.cookie.secure":           {"OWASP-ASVS-V3.4"},
	"session.cookie.httponly":         {"OWASP-ASVS-V3.4"},
	"session.cookie.samesite":         {"OWASP-ASVS-V3.4"},
	"session.invalidation.deleted":    {"OWASP-ASVS-V3.3"},
	"session.regenerate.deleted":      {"OWASP-ASVS-V3.3"},
	"session.fixation.input":          {"OWASP-ASVS-V3.3"},
	"session.timeout.deleted":         {"OWASP-ASVS-V3.3"},
	"session.token.singleuse.deleted": {"OWASP-ASVS-V3.5"},
	"auth.role.deleted":               {"OWASP-ASVS-V4.1"},
	"auth.owner.deleted":              {"OWASP-ASVS-V4.1"},
	"authorization.scope.deleted":     {"OWASP-ASVS-V4.1"},
	"authorization.bypass.added":      {"OWASP-ASVS-V4.1"},
	"authorization.idor.lookup":       {"OWASP-ASVS-V4.1"},
	"authorization.operation.deleted": {"OWASP-ASVS-V4.1"},
	"auth.redirect.untrusted":         {"OWASP-ASVS-V4.1"},
	"oauth.state.disabled":            {"OWASP-ASVS-V4.1"},
	"oauth.nonce.disabled":            {"OWASP-ASVS-V4.1"},
	"oauth.pkce.disabled":             {"OWASP-ASVS-V4.1"},
	"oauth.replay.deleted":            {"OWASP-ASVS-V4.1"},
	"token.claims.unchecked":          {"OWASP-ASVS-V3.5"},
	"token.refresh.rotation.deleted":  {"OWASP-ASVS-V3.5"},
	"token.binding.deleted":           {"OWASP-ASVS-V3.5"},
	"secret.aws.access":               {"OWASP-ASVS-V8.1"},
	"secret.aws.secret":               {"OWASP-ASVS-V8.1"},
	"secret.stripe":                   {"OWASP-ASVS-V8.1"},
	"secret.supabase":                 {"OWASP-ASVS-V8.1"},
	"secret.privatekey":               {"OWASP-ASVS-V6.4", "OWASP-ASVS-V8.1"},
	"secret.dotenv":                   {"OWASP-ASVS-V8.1"},
	"secret.jwt":                      {"OWASP-ASVS-V3.5", "OWASP-ASVS-V8.1"},
	"secret.generic":                  {"OWASP-ASVS-V8.1"},
	"crypto.mathrandom":               {"OWASP-ASVS-V6.2"},
	"crypto.staticiv":                 {"OWASP-ASVS-V6.2"},
	"crypto.nonce.reuse":              {"OWASP-ASVS-V6.2"},
	"crypto.weakhash":                 {"OWASP-ASVS-V6.2"},
	"crypto.weakcipher":               {"OWASP-ASVS-V6.2"},
	"crypto.custom":                   {"OWASP-ASVS-V6.2"},
	"crypto.compare.missing":          {"OWASP-ASVS-V6.2"},
	"crypto.key.exposed":              {"OWASP-ASVS-V6.4"},
	"crypto.alg.none":                 {"OWASP-ASVS-V3.5", "OWASP-ASVS-V6.2"},
	"crypto.alg.trusted":              {"OWASP-ASVS-V3.5", "OWASP-ASVS-V6.2"},
	"crypto.kid.trusted":              {"OWASP-ASVS-V3.5", "OWASP-ASVS-V6.2"},
	"crypto.jku.remote":               {"OWASP-ASVS-V3.5", "OWASP-ASVS-V6.2"},
	"crypto.key.confusion":            {"OWASP-ASVS-V3.5", "OWASP-ASVS-V6.2"},
	"crypto.verify.skip":              {"OWASP-ASVS-V6.2"},
	"crypto.verify.deleted":           {"OWASP-ASVS-V6.2"},
	"inject.sql.raw":                  {"OWASP-ASVS-V1.2"},
	"inject.command.exec":             {"OWASP-ASVS-V1.2"},
	"inject.ssrf.fetch":               {"OWASP-ASVS-V4.3"},
	"inject.xss.innerhtml":            {"OWASP-ASVS-V3.2"},
	"inject.deserialize":              {"OWASP-ASVS-V1.5"},
	"inject.traversal":                {"OWASP-ASVS-V5.4"},
	"inject.upload.validation":        {"OWASP-ASVS-V5.2"},
	"inject.tls.disable":              {"OWASP-ASVS-V9.1"},
	"inject.cors.wildcard":            {"OWASP-ASVS-V3.4"},
	"inject.logging.secret":           {"OWASP-ASVS-V8.1"},
	"inject.stacktrace":               {"OWASP-ASVS-V10.3"},
	"template.ssti.dynamic":           {"OWASP-ASVS-V1.2"},
	"template.eval.dynamic":           {"OWASP-ASVS-V1.2"},
	"file.archive.zipslip":            {"OWASP-ASVS-V5.4"},
	"file.upload.validation.deleted":  {"OWASP-ASVS-V5.2"},
	"supply.workflow.secret":          {"OWASP-ASVS-V14.2"},
	"supply.postinstall":              {"OWASP-ASVS-V14.2"},
	"supply.remote.dependency":        {"OWASP-ASVS-V14.2"},
	"supply.version.floating":         {"OWASP-ASVS-V14.2"},
	"supply.registry.public":          {"OWASP-ASVS-V14.2"},
	"supply.replace.remote":           {"OWASP-ASVS-V14.2"},
	"supply.typosquat":                {"OWASP-ASVS-V14.2"},
	"supply.lock.deleted":             {"OWASP-ASVS-V14.2"},
	"supply.integrity.deleted":        {"OWASP-ASVS-V14.2"},
	"supply.exec.workflow":            {"OWASP-ASVS-V14.2"},
	"supply.action.unpinned":          {"OWASP-ASVS-V14.2"},
	"supply.workflow.permissions":     {"OWASP-ASVS-V14.2"},
	"supply.workflow.pulltarget":      {"OWASP-ASVS-V14.2"},
	"config.debug.enabled":            {"OWASP-ASVS-V1.14"},
	"config.env.production":           {"OWASP-ASVS-V1.14"},
	"config.cors.wildcard":            {"OWASP-ASVS-V14.4"},
	"config.csp.disabled":             {"OWASP-ASVS-V14.4"},
	"config.tls.legacy":               {"OWASP-ASVS-V9.1"},
	"config.container.privileged":     {"OWASP-ASVS-V1.14"},
	"observe.request.body":            {"OWASP-ASVS-V8.1"},
	"observe.auth.material":           {"OWASP-ASVS-V8.1"},
	"observe.pii":                     {"OWASP-ASVS-V8.1"},
	"observe.error.stringify":         {"OWASP-ASVS-V10.3"},
	"observe.trace.sensitive":         {"OWASP-ASVS-V8.1"},
	"semantic.ssrf.flow":              {"OWASP-ASVS-V4.3"},
	"semantic.command.flow":           {"OWASP-ASVS-V1.2"},
	"semantic.traversal.flow":         {"OWASP-ASVS-V5.4"},
	"semantic.sql.flow":               {"OWASP-ASVS-V1.2"},
	"semantic.idor.flow":              {"OWASP-ASVS-V4.1"},
	"semantic.template.flow":          {"OWASP-ASVS-V1.2"},
	"semantic.go.ssrf":                {"OWASP-ASVS-V4.3"},
	"semantic.go.exec":                {"OWASP-ASVS-V1.2"},
	"semantic.go.path":                {"OWASP-ASVS-V5.4"},
	"semantic.go.template":            {"OWASP-ASVS-V1.2"},
	"branch.guard.regression":         {"OWASP-ASVS-V4.1"},
	"branch.scope.regression":         {"OWASP-ASVS-V4.1"},
	"branch.token.regression":         {"OWASP-ASVS-V3.5"},
	"downgrade.cookie.samesite":       {"OWASP-ASVS-V3.4"},
	"downgrade.cookie.lifetime":       {"OWASP-ASVS-V3.3"},
	"downgrade.crypto.bcrypt":         {"OWASP-ASVS-V6.2"},
	"downgrade.auth.ratelimit":        {"OWASP-ASVS-V7.2"},
	"downgrade.timeout":               {"OWASP-ASVS-V3.3", "OWASP-ASVS-V9.1"},
	"downgrade.auth.mfa":              {"OWASP-ASVS-V3.1"},
	"supply.lifecycle.dependency":     {"OWASP-ASVS-V14.2"},
	"supply.lock.missing":             {"OWASP-ASVS-V14.2"},
	"supply.registry.drift":           {"OWASP-ASVS-V14.2"},
	"supply.registry.untrusted":       {"OWASP-ASVS-V14.2"},
	"supply.package.denied":           {"OWASP-ASVS-V14.2"},
	"supply.package.untrusted":        {"OWASP-ASVS-V14.2"},
	"intent.mismatch":                 {"OWASP-ASVS-V1.14"},
	"auth.password.plaintext":         {"OWASP-ASVS-V2.4"},
	"auth.password.weakhash":          {"OWASP-ASVS-V2.4", "OWASP-ASVS-V6.2"},
	"crypto.rsa.keysize":              {"OWASP-ASVS-V6.2"},
	"inject.redos":                    {"OWASP-ASVS-V5.1"},
	"inject.xml.xxee":                 {"OWASP-ASVS-V5.4"},
	"inject.openredirect":             {"OWASP-ASVS-V5.1"},
	"inject.lfi":                      {"OWASP-ASVS-V12.3"},
	"inject.rfi":                      {"OWASP-ASVS-V12.3"},
	"file.download.traversal":         {"OWASP-ASVS-V5.4", "OWASP-ASVS-V12.4"},
	"config.hsts.missing":             {"OWASP-ASVS-V9.1"},
	"config.xframe.missing":           {"OWASP-ASVS-V14.4"},
	"config.xcontenttype.missing":     {"OWASP-ASVS-V14.4"},
	"logic.mass.assignment":           {"OWASP-ASVS-V11.1"},
	"logic.price.tampering":           {"OWASP-ASVS-V11.1"},
	"secret.pii.exposed":              {"OWASP-ASVS-V8.3"},
	"observe.healthdata":              {"OWASP-ASVS-V8.3"},
	"semantic.go.unsafe":              {"OWASP-ASVS-V5.5"},
}

func enrichControls(rules []model.Rule) []model.Rule {
	for index, item := range rules {
		controls := controlMap[item.ID]
		if len(controls) == 0 {
			continue
		}
		rules[index].Standards = mergeStandards(item.Standards, controls)
	}
	return rules
}

func mergeStandards(values []string, extras []string) []string {
	set := map[string]struct{}{}
	merged := make([]string, 0, len(values)+len(extras))
	for _, item := range values {
		if _, ok := set[item]; ok {
			continue
		}
		set[item] = struct{}{}
		merged = append(merged, item)
	}
	for _, item := range extras {
		if _, ok := set[item]; ok {
			continue
		}
		set[item] = struct{}{}
		merged = append(merged, item)
	}
	return merged
}
