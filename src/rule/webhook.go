package rule

import "github.com/Pimatis/mavetis/src/model"

func webhook() []model.Rule {
	return []model.Rule{
		{
			ID:          "webhook.signature.missing",
			Title:       "Webhook handler without signature verification",
			Message:     "The diff introduces a webhook endpoint without nearby signature verification signals.",
			Remediation: "Verify webhook signatures with provider-specific secrets before parsing or acting on the payload.",
			Category:    "webhook",
			Severity:    "critical",
			Confidence:  "medium",
			Target:      "added",
			Paths:       codeFiles(),
			Require:     []string{`(?i)(webhook|stripe|github|shopify|slack|clerk|supabase).*(post\(|handle|route|callback|handler)`},
			Absent:      []string{`(?i)(signature|verify|hmac|svix|constructEvent|webhookSecret|SigningSecret|X-Hub-Signature|Stripe-Signature)`},
			Standards:   standard("OWASP-ASVS-V4.1", "OWASP-ASVS-V6.2", "OWASP-Webhooks"),
		},
		{
			ID:          "webhook.replay.window.missing",
			Title:       "Webhook replay window missing",
			Message:     "The diff verifies webhook material without nearby timestamp or replay-window validation.",
			Remediation: "Validate provider timestamps and reject stale webhook deliveries outside a short replay window.",
			Category:    "webhook",
			Severity:    "high",
			Confidence:  "medium",
			Target:      "added",
			Paths:       codeFiles(),
			Require:     []string{`(?i)(webhook|stripe|github|shopify|slack|clerk|supabase).*(signature|verify|hmac|constructEvent)`},
			Absent:      []string{`(?i)(timestamp|tolerance|replay|nonce|delivery|eventId|idempoten|X-Slack-Request-Timestamp|Stripe-Signature)`},
			Standards:   standard("OWASP-ASVS-V4.1", "OWASP-Webhooks"),
		},
		{
			ID:          "webhook.rawbody.missing",
			Title:       "Webhook signature verification after parsed body",
			Message:     "The diff parses webhook JSON before signature verification, which can invalidate provider signature checks.",
			Remediation: "Verify signatures against the raw request body before JSON parsing or mutation.",
			Category:    "webhook",
			Severity:    "high",
			Confidence:  "medium",
			Target:      "added",
			Paths:       codeFiles(),
			Require:     []string{`(?i)(webhook|stripe|github|shopify|slack|clerk|supabase).*(req\.json\(|request\.json\(|JSON\.parse|Decode\(|BindJSON|bodyParser\.json)`},
			Absent:      []string{`(?i)(rawBody|bodyParser\.raw|express\.raw|ReadAll|io\.ReadAll|text\(|arrayBuffer\()`},
			Standards:   standard("OWASP-ASVS-V4.1", "OWASP-Webhooks"),
		},
	}
}
