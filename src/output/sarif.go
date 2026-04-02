package output

import (
	"encoding/json"

	"github.com/Pimatis/mavetis/src/model"
)

type sarif struct {
	Version string     `json:"version"`
	Schema  string     `json:"$schema"`
	Runs    []sarifRun `json:"runs"`
}

type sarifRun struct {
	Tool    sarifTool     `json:"tool"`
	Results []sarifResult `json:"results"`
}

type sarifTool struct {
	Driver sarifDriver `json:"driver"`
}

type sarifDriver struct {
	Name           string      `json:"name"`
	Version        string      `json:"version"`
	Rules          []sarifRule `json:"rules"`
	InformationURI string      `json:"informationUri"`
}

type sarifRule struct {
	ID               string          `json:"id"`
	Name             string          `json:"name"`
	ShortDescription sarifMessage    `json:"shortDescription"`
	Properties       sarifProperties `json:"properties"`
}

type sarifProperties struct {
	Severity  string   `json:"security-severity"`
	Category  string   `json:"category"`
	Standards []string `json:"standards"`
}

type sarifResult struct {
	RuleID    string          `json:"ruleId"`
	Level     string          `json:"level"`
	Message   sarifMessage    `json:"message"`
	Locations []sarifLocation `json:"locations"`
}

type sarifMessage struct {
	Text string `json:"text"`
}

type sarifLocation struct {
	PhysicalLocation sarifPhysical `json:"physicalLocation"`
}

type sarifPhysical struct {
	ArtifactLocation sarifArtifact `json:"artifactLocation"`
	Region           sarifRegion   `json:"region"`
}

type sarifArtifact struct {
	URI string `json:"uri"`
}

type sarifRegion struct {
	StartLine int `json:"startLine"`
}

func SARIF(report model.Report) (string, error) {
	rules := make([]sarifRule, 0, len(report.Rules))
	for _, rule := range report.Rules {
		rules = append(rules, sarifRule{
			ID:               rule.ID,
			Name:             rule.Title,
			ShortDescription: sarifMessage{Text: rule.Title},
			Properties:       sarifProperties{Severity: rule.Severity, Category: rule.Category, Standards: rule.Standards},
		})
	}
	results := make([]sarifResult, 0, len(report.Findings))
	for _, finding := range report.Findings {
		results = append(results, sarifResult{
			RuleID:  finding.RuleID,
			Level:   level(finding.Severity),
			Message: sarifMessage{Text: finding.Message},
			Locations: []sarifLocation{{
				PhysicalLocation: sarifPhysical{
					ArtifactLocation: sarifArtifact{URI: finding.Path},
					Region:           sarifRegion{StartLine: finding.Line},
				},
			}},
		})
	}
	document := sarif{
		Version: "2.1.0",
		Schema:  "https://json.schemastore.org/sarif-2.1.0.json",
		Runs: []sarifRun{{
			Tool:    sarifTool{Driver: sarifDriver{Name: model.Name, Version: model.Version, InformationURI: model.Repository, Rules: rules}},
			Results: results,
		}},
	}
	buffer, err := json.MarshalIndent(document, "", "  ")
	if err != nil {
		return "", err
	}
	return string(buffer), nil
}

func level(value string) string {
	if value == "critical" || value == "high" {
		return "error"
	}
	if value == "medium" {
		return "warning"
	}
	return "note"
}
