package rule

import (
	"fmt"
	"strings"

	"github.com/Pimatis/mavetis/src/model"
)

type ruleExample struct {
	Vulnerable string
	Safe       string
}

type syntheticDetail struct {
	Confidence      string
	Message         string
	Remediation     string
	Triggers        []string
	PositiveContext []string
	NegativeContext []string
	Example         ruleExample
}

func Explain(id string, rules []model.Rule) (model.RuleExplanation, bool) {
	id = strings.TrimSpace(id)
	if id == "" {
		return model.RuleExplanation{}, false
	}
	for _, item := range rules {
		if item.ID != id {
			continue
		}
		return explainRule(item), true
	}
	info, ok := syntheticInfo(id)
	if ok {
		return explainSynthetic(info), true
	}
	return model.RuleExplanation{}, false
}

func explainRule(item model.Rule) model.RuleExplanation {
	example := examplesFor(item.ID, item.Category)
	if item.VulnerableExample != "" {
		example.Vulnerable = item.VulnerableExample
	}
	if item.SafeExample != "" {
		example.Safe = item.SafeExample
	}
	return model.RuleExplanation{
		ID:                item.ID,
		Title:             item.Title,
		Severity:          item.Severity,
		Confidence:        item.Confidence,
		Category:          item.Category,
		Type:              item.Type,
		Target:            targetFor(item.Target),
		Engine:            engineFor(item),
		Message:           item.Message,
		Remediation:       item.Remediation,
		Standards:         mergeStandards(item.Standards, controlMap[item.ID]),
		Scope:             scopeFor(item),
		Triggers:          triggersFor(item),
		PositiveContext:   appendPatterns(nil, "near pattern", item.Near),
		NegativeContext:   negativeFor(item),
		VulnerableExample: example.Vulnerable,
		SafeExample:       example.Safe,
	}
}

func explainSynthetic(info model.RuleInfo) model.RuleExplanation {
	detail := detailFor(info)
	return model.RuleExplanation{
		ID:                info.ID,
		Title:             info.Title,
		Severity:          info.Severity,
		Confidence:        detail.Confidence,
		Category:          info.Category,
		Type:              "synthetic",
		Target:            "added",
		Engine:            "semantic correlation analyzer",
		Message:           detail.Message,
		Remediation:       detail.Remediation,
		Standards:         mergeStandards(info.Standards, controlMap[info.ID]),
		Scope:             []string{"target side: added", "source: diff hunk correlation"},
		Triggers:          append([]string{}, detail.Triggers...),
		PositiveContext:   append([]string{}, detail.PositiveContext...),
		NegativeContext:   append([]string{}, detail.NegativeContext...),
		VulnerableExample: detail.Example.Vulnerable,
		SafeExample:       detail.Example.Safe,
	}
}

func syntheticInfo(id string) (model.RuleInfo, bool) {
	for _, item := range SyntheticInfos() {
		if item.ID != id {
			continue
		}
		return item, true
	}
	return model.RuleInfo{}, false
}

func targetFor(value string) string {
	if value != "" {
		return value
	}
	return "added"
}

func engineFor(item model.Rule) string {
	if item.Type != "" {
		return "typed rule engine"
	}
	return "regex rule engine"
}

func scopeFor(item model.Rule) []string {
	scope := []string{"target side: " + targetFor(item.Target)}
	scope = appendPatterns(scope, "path scope", item.Paths)
	scope = appendPatterns(scope, "source path", item.FromPaths)
	scope = appendPatterns(scope, "forbidden path", item.ForbiddenPaths)
	return scope
}

func triggersFor(item model.Rule) []string {
	triggers := make([]string, 0)
	triggers = appendPatterns(triggers, "required pattern", item.Require)
	triggers = appendPatterns(triggers, "alternative pattern", item.Any)
	triggers = appendPatterns(triggers, "import pattern", item.Imports)
	triggers = appendPatterns(triggers, "required call", item.Calls)
	triggers = appendPatterns(triggers, "required middleware", item.Middleware)
	triggers = appendPatterns(triggers, "key pattern", item.Keys)
	triggers = appendPatterns(triggers, "allowed value", item.AllowedValues)
	triggers = appendPatterns(triggers, "forbidden value", item.ForbiddenValues)
	if item.ConstraintKey != "" {
		triggers = append(triggers, "constraint key: "+item.ConstraintKey)
	}
	if item.ConstraintPattern != "" {
		triggers = append(triggers, "constraint pattern: "+item.ConstraintPattern)
	}
	if item.Entropy > 0 {
		triggers = append(triggers, fmt.Sprintf("entropy threshold: %.2f", item.Entropy))
	}
	return triggers
}

func negativeFor(item model.Rule) []string {
	negative := make([]string, 0)
	negative = appendPatterns(negative, "absent guard", item.Absent)
	negative = appendPatterns(negative, "ignored path", item.Ignore)
	return negative
}

func appendPatterns(values []string, label string, patterns []string) []string {
	for _, pattern := range patterns {
		if pattern == "" {
			continue
		}
		values = append(values, label+": "+pattern)
	}
	return values
}

func detailFor(info model.RuleInfo) syntheticDetail {
	example := examplesFor(info.ID, info.Category)
	detail := syntheticDetail{
		Confidence:  "medium",
		Message:     "A local analyzer correlated multiple diff signals that are stronger together than a single regex match.",
		Remediation: "Follow the finding remediation emitted by review output and restore the missing security control.",
		Triggers:    []string{"synthetic analyzer: " + info.Title},
		Example:     example,
	}
	specific, ok := syntheticDetails[info.ID]
	if ok {
		detail = mergeDetail(detail, specific)
	}
	return detail
}

func mergeDetail(base syntheticDetail, specific syntheticDetail) syntheticDetail {
	if specific.Confidence != "" {
		base.Confidence = specific.Confidence
	}
	if specific.Message != "" {
		base.Message = specific.Message
	}
	if specific.Remediation != "" {
		base.Remediation = specific.Remediation
	}
	if len(specific.Triggers) != 0 {
		base.Triggers = append([]string{}, specific.Triggers...)
	}
	if len(specific.PositiveContext) != 0 {
		base.PositiveContext = append([]string{}, specific.PositiveContext...)
	}
	if len(specific.NegativeContext) != 0 {
		base.NegativeContext = append([]string{}, specific.NegativeContext...)
	}
	if specific.Example.Vulnerable != "" {
		base.Example.Vulnerable = specific.Example.Vulnerable
	}
	if specific.Example.Safe != "" {
		base.Example.Safe = specific.Example.Safe
	}
	return base
}

func examplesFor(id string, category string) ruleExample {
	example, ok := specificExamples[id]
	if ok {
		return example
	}
	example, ok = categoryExamples[category]
	if ok {
		return example
	}
	return ruleExample{
		Vulnerable: "Code that satisfies the listed trigger patterns without the listed guards.",
		Safe:       "Code that avoids the trigger pattern or implements the remediation before the sink executes.",
	}
}

var syntheticDetails = map[string]syntheticDetail{
	"semantic.ssrf.flow": {
		Message:     "Hunk-level taint analysis found request-controlled data reaching an outbound network sink without nearby allowlist or host validation.",
		Remediation: "Validate remote targets against an allowlist and block loopback, private, and metadata destinations.",
		Triggers: []string{
			"source: query, params, header, cookie, body, input, or request data",
			"sink: fetch, http.Get, http.Post, axios, requests.get/post, or urlopen",
			"tainted value appears in the same added hunk as the outbound sink",
		},
		NegativeContext: []string{"guard absent: authorize, permission, tenant, owner, scope, validateRedirect, allowlist, filepath.clean, or path.clean in the same hunk"},
	},
	"semantic.go.ssrf": {
		Message:     "Go AST flow analysis found request-derived data reaching http.Get inside the diff hunk.",
		Remediation: "Validate remote targets with an allowlist and reject private, loopback, and metadata destinations.",
		Triggers: []string{
			"source: Go call containing query, param, header, cookie, or body",
			"flow: assignment carries the request-derived value",
			"sink: http.Get consumes the tainted value",
		},
		NegativeContext: []string{"no regex absent guard: remove the tainted flow or validate before the http.Get call"},
		Example: ruleExample{
			Vulnerable: "target := r.URL.Query().Get(\"url\")\nhttp.Get(target)",
			Safe:       "target := r.URL.Query().Get(\"url\")\nif !allowedOutboundHost(target) {\n\treturn errBlockedHost\n}\nhttp.Get(target)",
		},
	},
}

var specificExamples = map[string]ruleExample{
	"inject.sql.raw": {
		Vulnerable: "query := \"SELECT * FROM users WHERE id = \" + userID\ndb.QueryContext(ctx, query)",
		Safe:       "db.QueryContext(ctx, \"SELECT * FROM users WHERE id = ?\", userID)",
	},
	"inject.ssrf.fetch": {
		Vulnerable: "target := r.URL.Query().Get(\"url\")\nhttp.Get(target)",
		Safe:       "target := r.URL.Query().Get(\"url\")\nif !allowedOutboundHost(target) {\n\treturn errBlockedHost\n}\nhttp.Get(target)",
	},
}

var categoryExamples = map[string]ruleExample{
	"ai":              {Vulnerable: "messages = append(messages, userPromptWithSecret)", Safe: "messages = append(messages, redactSecrets(userPrompt))"},
	"auth":            {Vulnerable: "if debug { return authenticatedUser }", Safe: "user, err := verifyCredentials(ctx, request)"},
	"authorization":   {Vulnerable: "user := users.Find(req.Param(\"id\"))", Safe: "user := users.FindForTenant(ctx, tenantID, req.Param(\"id\"))"},
	"boundary":        {Vulnerable: "import \"internal/admin\"", Safe: "callReviewedServerBoundary(ctx, request)"},
	"cloud":           {Vulnerable: "bucketAcl: public-read", Safe: "bucketAcl: private"},
	"config":          {Vulnerable: "NODE_ENV=development", Safe: "NODE_ENV=production"},
	"cors":            {Vulnerable: "AllowCredentials: true\nAllowedOrigins: []string{\"*\"}", Safe: "AllowCredentials: true\nAllowedOrigins: []string{\"https://app.example.com\"}"},
	"crypto":          {Vulnerable: "nonce := fixedNonce\ngcm.Seal(nil, nonce, data, nil)", Safe: "nonce := randomBytes(gcm.NonceSize())\ngcm.Seal(nil, nonce, data, nil)"},
	"deserialization": {Vulnerable: "pickle.loads(request.body)", Safe: "json.Unmarshal(body, &validatedDTO)"},
	"error":           {Vulnerable: "return err.StackTrace()", Safe: "return genericClientError"},
	"file":            {Vulnerable: "os.Open(filepath.Join(root, req.URL.Query().Get(\"path\")))", Safe: "path := safeJoin(root, req.URL.Query().Get(\"path\"))\nos.Open(path)"},
	"injection":       {Vulnerable: "exec.Command(\"sh\", \"-c\", userInput)", Safe: "exec.Command(binary, validatedArg)"},
	"intent":          {Vulnerable: "func verifyOwnership() bool { return true }", Safe: "func verifyOwnership(ctx context.Context, resource Resource) bool { return resource.OwnerID == ctxUserID(ctx) }"},
	"logging":         {Vulnerable: "logger.Info(\"token\", token)", Safe: "logger.Info(\"token accepted\")"},
	"logic":           {Vulnerable: "price := req.Body.Price", Safe: "price := catalog.PriceFor(productID)"},
	"memory":          {Vulnerable: "ptr := unsafe.Pointer(&value)", Safe: "use typed conversion without unsafe.Pointer"},
	"oauth":           {Vulnerable: "validateState = false", Safe: "validateState = true"},
	"privacy":         {Vulnerable: "response.Write(user.SSN)", Safe: "response.Write(maskPII(user.SSN))"},
	"secret":          {Vulnerable: "const apiKey = \"sk_live_example\"", Safe: "apiKey := os.Getenv(\"API_KEY\")"},
	"session":         {Vulnerable: "cookie.Secure = false", Safe: "cookie.Secure = true"},
	"ssrf":            {Vulnerable: "target := req.URL.Query().Get(\"url\")\nhttp.Get(target)", Safe: "target := req.URL.Query().Get(\"url\")\nif !allowedOutboundHost(target) {\n\treturn errBlockedHost\n}\nhttp.Get(target)"},
	"supply":          {Vulnerable: "\"left-pad\": \"latest\"", Safe: "\"@company/left-pad\": \"1.2.3\""},
	"template":        {Vulnerable: "template.New(\"page\").Parse(req.FormValue(\"tpl\"))", Safe: "template.ParseFS(templates, \"page.html\")"},
	"token":           {Vulnerable: "claims := jwt.Decode(token)", Safe: "claims, err := jwt.Verify(token, trustedKey)"},
	"transport":       {Vulnerable: "InsecureSkipVerify: true", Safe: "MinVersion: tls.VersionTLS12"},
	"webhook":         {Vulnerable: "payload := readBody(r)\nhandleWebhook(payload)", Safe: "payload := readRawBody(r)\nverifyWebhookSignature(payload, r.Header)"},
	"xss":             {Vulnerable: "element.innerHTML = request.query.html", Safe: "element.textContent = request.query.html"},
}
