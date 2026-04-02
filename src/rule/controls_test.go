package rule

import (
	"testing"

	"github.com/Pimatis/mavetis/src/model"
)

func TestBuiltinsIncludeExactASVSControls(t *testing.T) {
	rules := Builtins(model.Config{})
	for _, item := range rules {
		if item.ID != "auth.middleware.deleted" {
			continue
		}
		for _, standard := range item.Standards {
			if standard == "OWASP-ASVS-V3.1" {
				return
			}
		}
		t.Fatalf("expected exact ASVS control on %s: %#v", item.ID, item.Standards)
	}
	t.Fatal("expected builtin auth rule metadata")
}
