package wizard

import (
	"strings"
)

type ConfigTemplate struct {
	Profile    string
	Severity   string
	FailOn     string
	Output     string
	Ignore     []string
	Critical   []string
	Restricted []string
}

func Generate(template ConfigTemplate) string {
	var b strings.Builder
	b.WriteString("# Mavetis security review configuration\n")
	b.WriteString("# https://github.com/pimatis/mavetis\n\n")
	b.WriteString("profile: " + template.Profile + "\n")
	b.WriteString("severity: " + template.Severity + "\n")
	b.WriteString("fail-on: " + template.FailOn + "\n")
	b.WriteString("output: " + template.Output + "\n")
	if len(template.Ignore) > 0 {
		b.WriteString("\nignore:\n")
		for _, item := range template.Ignore {
			b.WriteString("  - " + item + "\n")
		}
	}
	if len(template.Critical) > 0 || len(template.Restricted) > 0 {
		b.WriteString("\nzones:\n")
		if len(template.Critical) > 0 {
			b.WriteString("  critical:\n")
			for _, item := range template.Critical {
				b.WriteString("    - " + item + "\n")
			}
		}
		if len(template.Restricted) > 0 {
			b.WriteString("  restricted:\n")
			for _, item := range template.Restricted {
				b.WriteString("    - " + item + "\n")
			}
		}
	}
	return b.String()
}
