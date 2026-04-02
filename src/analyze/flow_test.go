package analyze

import "testing"

func TestFlowHelpers(t *testing.T) {
	vars := Tainted(`target := ctx.Query("url")`)
	if len(vars) != 1 || vars[0] != "target" {
		t.Fatalf("unexpected tainted vars: %#v", vars)
	}
	if !TaintedUse(`http.Get(target)`, vars) {
		t.Fatal("expected tainted use")
	}
	if !Guarded(`validateRedirect(returnTo)`) {
		t.Fatal("expected guarded text")
	}
}
