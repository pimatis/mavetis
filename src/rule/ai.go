package rule

import "github.com/Pimatis/mavetis/src/model"

func ai() []model.Rule {
	return []model.Rule{
		{
			ID:          "ai.prompt.secret.exposure",
			Title:       "Secret material added to AI prompt",
			Message:     "The diff appears to include secrets, tokens, or credentials in an AI prompt or model message.",
			Remediation: "Never send secrets to LLM prompts; pass only minimum necessary context after redaction.",
			Category:    "ai",
			Severity:    "critical",
			Confidence:  "high",
			Target:      "added",
			Paths:       codeFiles(),
			Require:     []string{`(?i)(\bprompt\b|\bmessages\b|\bsystem\b|\buser\b|\bassistant\b|chat\.completions|generateContent).*(api[_-]?key|secret|password|token|authorization|cookie|privateKey|process\.env)`},
			Standards:   standard("OWASP-ASVS-V8.1", "OWASP-LLM01", "OWASP-LLM06"),
			Mask:        true,
		},
		{
			ID:          "ai.prompt.user.system.mix",
			Title:       "User input assigned to system prompt",
			Message:     "The diff appears to place request-controlled input into a privileged system prompt role.",
			Remediation: "Keep system prompts static and place untrusted user content only in user-scoped message fields with clear delimiters.",
			Category:    "ai",
			Severity:    "high",
			Confidence:  "medium",
			Target:      "added",
			Paths:       codeFiles(),
			Require:     []string{`(?i)(role\s*[:=]\s*["']system["']|systemPrompt|system_instruction).*(req\.|request\.|ctx\.|params|query|body|input|userInput)`},
			Standards:   standard("OWASP-LLM01", "OWASP-LLM04"),
		},
		{
			ID:          "ai.tool.untrusted.input",
			Title:       "AI tool execution from untrusted model output",
			Message:     "The diff executes tools or functions from model output without visible allowlist or schema validation.",
			Remediation: "Validate model-selected tools against an allowlist and strict schema before executing side-effecting operations.",
			Category:    "ai",
			Severity:    "critical",
			Confidence:  "medium",
			Target:      "added",
			Paths:       codeFiles(),
			Require:     []string{`(?i)(tool_calls|function_call|modelOutput|assistantMessage|llmResponse).*(execute|invoke|call|dispatch|run)`},
			Absent:      []string{`(?i)(allowlist|allowedTools|schema|validate|zod|jsonschema|safeTool|permission)`},
			Standards:   standard("OWASP-LLM05", "OWASP-LLM06"),
		},
	}
}
