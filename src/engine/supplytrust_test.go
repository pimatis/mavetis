package engine

import (
	"testing"

	"github.com/Pimatis/mavetis/src/model"
)

func TestSupplyTrustFindingsDetectPolicyAndRegistryRisks(t *testing.T) {
	config := model.Config{Supply: model.Supply{AllowPackages: []string{"@company/*"}, DenyPackages: []string{"left-pad"}, TrustedRegistries: []string{"registry.company.local"}}}
	diff := model.Diff{Files: []model.DiffFile{
		{Path: ".npmrc", Hunks: []model.DiffHunk{{Lines: []model.DiffLine{{Kind: "deleted", Text: `registry=https://registry.company.local`, OldNumber: 1}, {Kind: "added", Text: `registry=https://registry.npmjs.org`, NewNumber: 1}}}}},
		{Path: "package.json", Hunks: []model.DiffHunk{{Lines: []model.DiffLine{{Kind: "added", Text: `"dependencies": {`, NewNumber: 2}, {Kind: "added", Text: `"left-pad": "1.3.0"`, NewNumber: 3}, {Kind: "added", Text: `"chalk": "5.0.0"`, NewNumber: 4}, {Kind: "added", Text: `"scripts": {`, NewNumber: 5}, {Kind: "added", Text: `"postinstall": "node setup.js"`, NewNumber: 6}}}}},
	}}
	findings := supplyTrustFindings(diff, config)
	expected := []string{"supply.registry.drift", "supply.registry.untrusted", "supply.package.denied", "supply.package.untrusted", "supply.lifecycle.dependency", "supply.lock.missing"}
	for _, id := range expected {
		found := false
		for _, item := range findings {
			if item.RuleID == id {
				found = true
			}
		}
		if !found {
			t.Fatalf("expected %s in %#v", id, findings)
		}
	}
}
