package model

type Report struct {
	Meta     DiffMeta   `json:"meta"`
	Policy   *Policy    `json:"policy,omitempty"`
	Summary  Summary    `json:"summary"`
	Findings []Finding  `json:"findings"`
	Rules    []RuleInfo `json:"rules"`
}

type RuleInfo struct {
	ID        string   `json:"id"`
	Title     string   `json:"title"`
	Category  string   `json:"category"`
	Severity  string   `json:"severity"`
	Standards []string `json:"standards"`
}

type Policy struct {
	Profile string       `json:"profile,omitempty"`
	FailOn  string       `json:"failOn,omitempty"`
	Zones   []PolicyZone `json:"zones,omitempty"`
}

type PolicyZone struct {
	Name           string   `json:"name"`
	Paths          []string `json:"paths,omitempty"`
	SeverityOffset int      `json:"severityOffset"`
	FailOn         string   `json:"failOn"`
}
