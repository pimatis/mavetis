package model

type Rule struct {
	ID          string
	Title       string
	Message     string
	Remediation string
	Category    string
	Severity    string
	Confidence  string
	Target      string
	Paths       []string
	Ignore      []string
	Require     []string
	Any         []string
	Near        []string
	Absent      []string
	Entropy     float64
	Standards   []string
	Mask        bool
}

func SeverityRank(value string) int {
	if value == "critical" {
		return 4
	}
	if value == "high" {
		return 3
	}
	if value == "medium" {
		return 2
	}
	if value == "low" {
		return 1
	}
	return 0
}
