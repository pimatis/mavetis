package rule

import (
	"testing"

	"github.com/Pimatis/mavetis/src/model"
)

func TestBuiltinsExposeStandardsForEveryRule(t *testing.T) {
	rules := Builtins(model.Config{})
	if len(rules) == 0 {
		t.Fatal("expected builtin rules")
	}
	for _, item := range rules {
		if item.ID == "" {
			t.Fatal("expected rule id")
		}
		if item.Title == "" {
			t.Fatalf("expected title for %s", item.ID)
		}
		if len(item.Standards) == 0 {
			t.Fatalf("expected standards for %s", item.ID)
		}
	}
}

func TestBuiltinsExposeExpandedFamilies(t *testing.T) {
	rules := Builtins(model.Config{})
	expected := []string{
		"session.fixation.input",
		"authorization.scope.deleted",
		"oauth.state.disabled",
		"crypto.verify.deleted",
		"supply.remote.dependency",
	}
	for _, id := range expected {
		if !contains(rules, id) {
			t.Fatalf("expected builtin rule %s", id)
		}
	}
}

func contains(rules []model.Rule, id string) bool {
	for _, item := range rules {
		if item.ID == id {
			return true
		}
	}
	return false
}
