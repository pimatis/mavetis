package model

type Report struct {
	Meta     DiffMeta   `json:"meta"`
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
