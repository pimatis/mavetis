package model

type Finding struct {
	ID              string   `json:"id"`
	RuleID          string   `json:"ruleId"`
	Title           string   `json:"title"`
	Category        string   `json:"category"`
	Severity        string   `json:"severity"`
	BaseSeverity    string   `json:"baseSeverity,omitempty"`
	Confidence      string   `json:"confidence"`
	Path            string   `json:"path"`
	Line            int      `json:"line"`
	Side            string   `json:"side"`
	Zone            string   `json:"zone,omitempty"`
	EffectiveFailOn string   `json:"effectiveFailOn,omitempty"`
	Message         string   `json:"message"`
	Snippet         string   `json:"snippet"`
	Remediation     string   `json:"remediation"`
	Reasons         []string `json:"reasons,omitempty"`
	Standards       []string `json:"standards"`
}

type Summary struct {
	Files    int `json:"files"`
	Findings int `json:"findings"`
	Low      int `json:"low"`
	Medium   int `json:"medium"`
	High     int `json:"high"`
	Critical int `json:"critical"`
}

func (summary *Summary) Add(finding Finding) {
	summary.Findings++
	if finding.Severity == "low" {
		summary.Low++
	}
	if finding.Severity == "medium" {
		summary.Medium++
	}
	if finding.Severity == "high" {
		summary.High++
	}
	if finding.Severity == "critical" {
		summary.Critical++
	}
}
