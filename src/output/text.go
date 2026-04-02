package output

import (
	"fmt"
	"strings"

	"github.com/Pimatis/mavetis/src/model"
)

func Text(report model.Report) string {
	return TextExplain(report, false)
}

func TextExplain(report model.Report, explain bool) string {
	builder := strings.Builder{}
	tone := colors()
	builder.WriteString(line("Mode", report.Meta.Mode, tone))
	if report.Meta.Base != "" {
		builder.WriteString(line("Base", report.Meta.Base, tone))
	}
	if report.Meta.Head != "" {
		builder.WriteString(line("Head", report.Meta.Head, tone))
	}
	builder.WriteString(line("Files", fmt.Sprintf("%d", report.Summary.Files), tone))
	if report.Policy != nil {
		if report.Policy.Profile != "" {
			builder.WriteString(line("Profile", report.Policy.Profile, tone))
		}
		if report.Policy.FailOn != "" {
			builder.WriteString(line("FailOn", report.Policy.FailOn, tone))
		}
		for _, zone := range report.Policy.Zones {
			builder.WriteString(line("Zone", fmt.Sprintf("%s severity+%d fail-on=%s", zone.Name, zone.SeverityOffset, zone.FailOn), tone))
		}
	}
	builder.WriteString(summary(report.Summary, tone))
	for _, finding := range report.Findings {
		builder.WriteString("\n")
		builder.WriteString(title(finding, tone))
		builder.WriteString(line("Rule", finding.RuleID, tone))
		builder.WriteString(line("File", fmt.Sprintf("%s:%d", finding.Path, finding.Line), tone))
		builder.WriteString(line("Side", finding.Side, tone))
		builder.WriteString(line("Confidence", finding.Confidence, tone))
		if finding.Zone != "" {
			builder.WriteString(line("Zone", finding.Zone, tone))
		}
		if finding.BaseSeverity != "" {
			builder.WriteString(line("BaseSeverity", finding.BaseSeverity, tone))
		}
		if finding.EffectiveFailOn != "" {
			builder.WriteString(line("EffectiveFailOn", finding.EffectiveFailOn, tone))
		}
		builder.WriteString(line("Message", finding.Message, tone))
		builder.WriteString(line("Snippet", strings.TrimSpace(finding.Snippet), tone))
		builder.WriteString(line("Fix", finding.Remediation, tone))
		if explain {
			builder.WriteString(reasons(finding, tone))
		}
	}
	return builder.String()
}

func line(label string, value string, tone palette) string {
	return fmt.Sprintf("%s: %s\n", paint(tone.label, label, tone), value)
}

func summary(data model.Summary, tone palette) string {
	parts := []string{
		fmt.Sprintf("%s=%d", paint(tone.critical, "critical", tone), data.Critical),
		fmt.Sprintf("%s=%d", paint(tone.high, "high", tone), data.High),
		fmt.Sprintf("%s=%d", paint(tone.medium, "medium", tone), data.Medium),
		fmt.Sprintf("%s=%d", paint(tone.low, "low", tone), data.Low),
	}
	value := fmt.Sprintf("%d (%s)", data.Findings, strings.Join(parts, " "))
	return line("Findings", value, tone)
}

func reasons(finding model.Finding, tone palette) string {
	if len(finding.Reasons) == 0 {
		return ""
	}
	builder := strings.Builder{}
	builder.WriteString(line("Why", finding.Reasons[0], tone))
	for index := 1; index < len(finding.Reasons); index++ {
		builder.WriteString(line("Why", finding.Reasons[index], tone))
	}
	return builder.String()
}

func title(finding model.Finding, tone palette) string {
	severity := strings.ToUpper(finding.Severity)
	badge := "[" + severity + "]"
	return fmt.Sprintf("%s %s\n", paint(severityColor(finding.Severity, tone), badge, tone), finding.Title)
}
