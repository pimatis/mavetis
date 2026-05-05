package output

import (
	"strings"

	"github.com/Pimatis/mavetis/src/model"
)

func RuleExplanation(data model.RuleExplanation) string {
	builder := strings.Builder{}
	writeField(&builder, "Rule", data.ID)
	writeField(&builder, "Title", data.Title)
	writeField(&builder, "Severity", data.Severity)
	writeField(&builder, "Confidence", data.Confidence)
	writeField(&builder, "Category", data.Category)
	writeField(&builder, "Type", data.Type)
	writeField(&builder, "Target", data.Target)
	writeField(&builder, "Engine", data.Engine)
	writeList(&builder, "ASVS mappings", asvsMappings(data.Standards))
	writeList(&builder, "Standards", nonASVSStandards(data.Standards))
	writeField(&builder, "Why it fires", data.Message)
	writeList(&builder, "Scope", data.Scope)
	writeList(&builder, "Trigger patterns", data.Triggers)
	writeList(&builder, "Positive context", data.PositiveContext)
	writeList(&builder, "Negative context / absent guards", data.NegativeContext)
	writeBlock(&builder, "Example vulnerable snippet", data.VulnerableExample)
	writeBlock(&builder, "Example safe pattern", data.SafeExample)
	writeField(&builder, "Remediation", data.Remediation)
	return builder.String()
}

func writeField(builder *strings.Builder, label string, value string) {
	value = strings.TrimSpace(value)
	if value == "" {
		value = "none"
	}
	builder.WriteString(label)
	builder.WriteString(": ")
	builder.WriteString(value)
	builder.WriteString("\n")
}

func writeList(builder *strings.Builder, label string, values []string) {
	builder.WriteString(label)
	builder.WriteString(":")
	builder.WriteString("\n")
	if len(values) == 0 {
		builder.WriteString("  - none\n")
		return
	}
	for _, item := range values {
		value := strings.TrimSpace(item)
		if value == "" {
			continue
		}
		builder.WriteString("  - ")
		builder.WriteString(value)
		builder.WriteString("\n")
	}
}

func writeBlock(builder *strings.Builder, label string, value string) {
	value = strings.TrimSpace(value)
	builder.WriteString(label)
	builder.WriteString(":")
	builder.WriteString("\n")
	if value == "" {
		builder.WriteString("  none\n")
		return
	}
	for _, line := range strings.Split(value, "\n") {
		builder.WriteString("  ")
		builder.WriteString(line)
		builder.WriteString("\n")
	}
}

func asvsMappings(values []string) []string {
	result := make([]string, 0, len(values))
	for _, item := range values {
		if !strings.HasPrefix(item, "OWASP-ASVS-") {
			continue
		}
		result = append(result, item)
	}
	return result
}

func nonASVSStandards(values []string) []string {
	result := make([]string, 0, len(values))
	for _, item := range values {
		if strings.HasPrefix(item, "OWASP-ASVS-") {
			continue
		}
		result = append(result, item)
	}
	return result
}
