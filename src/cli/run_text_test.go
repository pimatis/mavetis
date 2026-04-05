package cli

import (
	"testing"

	"github.com/Pimatis/mavetis/src/model"
)

func TestSuggestedCommandBuildsReviewCommand(t *testing.T) {
	spec := model.Review{Files: []string{"src/cli/run.go"}, WithSuggested: false}
	command := suggestedCommand(spec)
	if command != "mavetis review src/cli/run.go --with-suggested" {
		t.Fatalf("unexpected command: %s", command)
	}
}

func TestSuggestedCommandPreservesRelevantFlags(t *testing.T) {
	spec := model.Review{
		Files:      []string{"src/cli/run.go", "src/output/text.go"},
		Path:       "src/**",
		Profile:    "backend",
		Severity:   "high",
		FailOn:     "medium",
		Format:     "json",
		Explain:    true,
		ConfigPath: ".mavetis.yaml",
		RulesPath:  "rules.yaml",
	}
	command := suggestedCommand(spec)
	want := "mavetis review src/cli/run.go src/output/text.go --path \"src/**\" --profile backend --severity high --fail-on medium --config .mavetis.yaml --rules rules.yaml --format json --explain --with-suggested"
	if command != want {
		t.Fatalf("unexpected command: %s", command)
	}
}
