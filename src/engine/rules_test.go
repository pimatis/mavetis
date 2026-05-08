package engine

import (
	"testing"

	"github.com/Pimatis/mavetis/src/model"
	"github.com/Pimatis/mavetis/src/rule"
)

type ruleCase struct {
	name     string
	path     string
	line     model.DiffLine
	hunkText string
	expect   string
}

func TestBuiltinsCoverExpandedSecurityRules(t *testing.T) {
	cases := []ruleCase{
		{
			name:     "session fixation heuristic",
			path:     "auth/session.go",
			line:     model.DiffLine{Kind: "added", Text: `sessionID := ctx.Query("session")`, NewNumber: 10},
			hunkText: "login handler\nsessionID := ctx.Query(\"session\")\ncreate session",
			expect:   "session.fixation.input",
		},
		{
			name:     "authorization scope deletion",
			path:     "repository/user.go",
			line:     model.DiffLine{Kind: "deleted", Text: `query = query.Where("tenant_id = ?", tenantID)`, OldNumber: 42},
			hunkText: "find user\nquery = query.Where(\"tenant_id = ?\", tenantID)\nupdate",
			expect:   "authorization.scope.deleted",
		},
		{
			name:     "oauth state disabled",
			path:     "auth/oauth.ts",
			line:     model.DiffLine{Kind: "added", Text: `validateState = false`, NewNumber: 8},
			hunkText: "oauth callback\nvalidateState = false\nauthorize",
			expect:   "oauth.state.disabled",
		},
		{
			name:     "crypto verification removed",
			path:     "auth/token.go",
			line:     model.DiffLine{Kind: "deleted", Text: `claims, err := jwt.ParseWithClaims(token, claims, keyFunc)`, OldNumber: 33},
			hunkText: "token verify\nclaims, err := jwt.ParseWithClaims(token, claims, keyFunc)\nreturn claims",
			expect:   "crypto.verify.deleted",
		},
		{
			name:     "remote dependency source",
			path:     "package.json",
			line:     model.DiffLine{Kind: "added", Text: `"lib": "git+https://github.com/example/lib.git#main"`, NewNumber: 5},
			hunkText: `"dependencies": { "lib": "git+https://github.com/example/lib.git#main" }`,
			expect:   "supply.remote.dependency",
		},
		{
			name:     "webhook signature missing",
			path:     "api/webhook.ts",
			line:     model.DiffLine{Kind: "added", Text: `stripeWebhookHandler = app.post("/webhook", handler)`, NewNumber: 6},
			hunkText: `stripeWebhookHandler = app.post("/webhook", handler)`,
			expect:   "webhook.signature.missing",
		},
		{
			name:     "webhook raw body missing",
			path:     "api/webhook.ts",
			line:     model.DiffLine{Kind: "added", Text: `stripeWebhookHandler = await request.json()`, NewNumber: 7},
			hunkText: `stripeWebhookHandler = await request.json()`,
			expect:   "webhook.rawbody.missing",
		},
		{
			name:     "tenant lookup missing",
			path:     "repository/invoice.ts",
			line:     model.DiffLine{Kind: "added", Text: `const invoice = await db.invoice.findUnique({ where: { invoiceId: req.params.id } })`, NewNumber: 14},
			hunkText: `const invoice = await db.invoice.findUnique({ where: { invoiceId: req.params.id } })`,
			expect:   "authorization.tenant.lookup.missing",
		},
		{
			name:     "password reset token logged",
			path:     "auth/reset.ts",
			line:     model.DiffLine{Kind: "added", Text: `logger.info("reset token", resetToken)`, NewNumber: 22},
			hunkText: `logger.info("reset token", resetToken)`,
			expect:   "auth.reset.token.logged",
		},
		{
			name:     "public storage read",
			path:     "infra/bucket.yaml",
			line:     model.DiffLine{Kind: "added", Text: `bucketAcl: public-read`, NewNumber: 4},
			hunkText: `s3 bucket bucketAcl: public-read`,
			expect:   "cloud.storage.public.read",
		},
		{
			name:     "wildcard iam policy",
			path:     "infra/policy.json",
			line:     model.DiffLine{Kind: "added", Text: `"Action": "*"`, NewNumber: 9},
			hunkText: `{"Effect":"Allow","Action":"*","Resource":"arn:aws:s3:::example"}`,
			expect:   "iac.iam.policy.wildcard",
		},
		{
			name:     "ai prompt secret exposure",
			path:     "src/ai.ts",
			line:     model.DiffLine{Kind: "added", Text: `messages.push({ role: "user", content: "api_key=" + process.env.API_KEY })`, NewNumber: 30},
			hunkText: `messages.push({ role: "user", content: "api_key=" + process.env.API_KEY })`,
			expect:   "ai.prompt.secret.exposure",
		},
		{
			name:     "ai tool untrusted input",
			path:     "src/agent.ts",
			line:     model.DiffLine{Kind: "added", Text: `tool_calls.forEach(call => dispatch(call.name, call.arguments))`, NewNumber: 31},
			hunkText: `tool_calls.forEach(call => dispatch(call.name, call.arguments))`,
			expect:   "ai.tool.untrusted.input",
		},
		{
			name:     "websocket origin missing",
			path:     "api/ws.go",
			line:     model.DiffLine{Kind: "added", Text: `var upgrader = websocket.Upgrader{}`, NewNumber: 12},
			hunkText: `var upgrader = websocket.Upgrader{}`,
			expect:   "websocket.origin.missing",
		},
		{
			name:     "websocket downgrade to plain ws",
			path:     "frontend/app.ts",
			line:     model.DiffLine{Kind: "added", Text: `const socket = new WebSocket("ws://api.example.com/chat")`, NewNumber: 8},
			hunkText: `const socket = new WebSocket("ws://api.example.com/chat")`,
			expect:   "websocket.downgrade.missing",
		},
		{
			name:     "websocket message validation missing",
			path:     "server/chat.go",
			line:     model.DiffLine{Kind: "added", Text: `conn.ReadMessage()`, NewNumber: 22},
			hunkText: `conn.ReadMessage()`,
			expect:   "websocket.message.validation.missing",
		},
		{
			name:     "websocket auth missing",
			path:     "server/socket.ts",
			line:     model.DiffLine{Kind: "added", Text: `const io = new socketIo.Server(httpServer, {})`, NewNumber: 5},
			hunkText: `const io = new socketIo.Server(httpServer, {})`,
			expect:   "websocket.auth.missing",
		},
		{
			name:     "websocket ratelimit missing",
			path:     "server/stream.go",
			line:     model.DiffLine{Kind: "added", Text: `ws.Handle("/stream", streamHandler)`, NewNumber: 15},
			hunkText: `ws.Handle("/stream", streamHandler)`,
			expect:   "websocket.ratelimit.missing",
		},
		{
			name:     "race file toctou stat then write",
			path:     "storage/upload.go",
			line:     model.DiffLine{Kind: "added", Text: `if _, err := os.Stat(filePath); os.IsNotExist(err) {`, NewNumber: 18},
			hunkText: `if _, err := os.Stat(filePath); os.IsNotExist(err) {\nos.WriteFile(filePath, data, 0644)\n}`,
			expect:   "race.file.toctou",
		},
		{
			name:     "race file link toctou open in temp",
			path:     "util/download.go",
			line:     model.DiffLine{Kind: "added", Text: `f, err := os.OpenFile(path.Join(tmpDir, name), os.O_WRONLY|os.O_CREATE, 0600)`, NewNumber: 25},
			hunkText: `f, err := os.OpenFile(path.Join(tmpDir, name), os.O_WRONLY|os.O_CREATE, 0600)`,
			expect:   "race.file.link.toctou",
		},
		{
			name:     "race db concurrent without locking",
			path:     "wallet/service.ts",
			line:     model.DiffLine{Kind: "added", Text: `const balance = await db.wallet.findUnique({ where: { id } })\nawait db.wallet.update({ where: { id }, data: { balance: balance + amount } })`, NewNumber: 30},
			hunkText: `const balance = await db.wallet.findUnique({ where: { id } })\nawait db.wallet.update({ where: { id }, data: { balance: balance + amount } })`,
			expect:   "race.db.concurrent",
		},
		{
			name:     "race counter increment non-atomic",
			path:     "stats/visit.go",
			line:     model.DiffLine{Kind: "added", Text: `page.Views++`, NewNumber: 10},
			hunkText: `page, _ := repo.Find(id)\npage.Views++\nrepo.Save(page)`,
			expect:   "race.counter.increment",
		},
	}
	for _, item := range cases {
		t.Run(item.name, func(t *testing.T) {
			report := runBuiltin(t, item.path, item.line, item.hunkText)
			if !hasRule(report, item.expect) {
				t.Fatalf("expected rule %s, got %#v", item.expect, report.Findings)
			}
		})
	}
}

func TestBuiltinsSuppressWhenMitigationExists(t *testing.T) {
	report := runBuiltin(t, "auth/oauth.ts", model.DiffLine{Kind: "added", Text: `returnTo = request.query.redirect`, NewNumber: 12}, "oauth redirect\nreturnTo = request.query.redirect\nvalidateRedirect(returnTo)")
	if hasRule(report, "auth.redirect.untrusted") {
		t.Fatalf("expected mitigation-aware suppression, got %#v", report.Findings)
	}
}

func TestWebsocketSuppressWhenOriginCheckPresent(t *testing.T) {
	report := runBuiltin(t, "api/ws.go", model.DiffLine{Kind: "added", Text: `var upgrader = websocket.Upgrader{CheckOrigin: allowOrigin}`, NewNumber: 12}, "var upgrader = websocket.Upgrader{CheckOrigin: allowOrigin}")
	if hasRule(report, "websocket.origin.missing") {
		t.Fatalf("expected origin check suppression, got %#v", report.Findings)
	}
}

func TestRaceSuppressWhenAtomicGuardPresent(t *testing.T) {
	report := runBuiltin(t, "storage/upload.go", model.DiffLine{Kind: "added", Text: `if _, err := os.Stat(filePath); os.IsNotExist(err) {`, NewNumber: 18}, "if _, err := os.Stat(filePath); os.IsNotExist(err) {\nf, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)\n}")
	if hasRule(report, "race.file.toctou") {
		t.Fatalf("expected O_EXCL suppression, got %#v", report.Findings)
	}
}

func TestRaceCounterSuppressWhenAtomicPresent(t *testing.T) {
	report := runBuiltin(t, "stats/visit.go", model.DiffLine{Kind: "added", Text: `mu.Lock()`, NewNumber: 9}, "mu.Lock()\npage.Views++\nmu.Unlock()")
	if hasRule(report, "race.counter.increment") {
		t.Fatalf("expected mutex suppression, got %#v", report.Findings)
	}
}

func runBuiltin(t *testing.T, path string, line model.DiffLine, hunkText string) model.Report {
	t.Helper()
	config := model.Config{Severity: "low"}
	hunk := model.DiffHunk{Lines: []model.DiffLine{line}}
	if hunkText != "" {
		hunk.Lines = []model.DiffLine{{Kind: line.Kind, Text: hunkText, OldNumber: line.OldNumber, NewNumber: line.NewNumber}, line}
	}
	diff := model.Diff{Files: []model.DiffFile{{Path: path, Hunks: []model.DiffHunk{hunk}}}}
	report, err := Review(diff, config, rule.Builtins(config))
	if err != nil {
		t.Fatalf("review failed: %v", err)
	}
	return report
}

func hasRule(report model.Report, ruleID string) bool {
	for _, finding := range report.Findings {
		if finding.RuleID == ruleID {
			return true
		}
	}
	return false
}
