package analyze

import "testing"

func TestGoFlows(t *testing.T) {
	body := `target := ctx.Query("url")
copy := target
http.Get(copy)`
	flows := GoFlows(body)
	if len(flows) != 1 {
		t.Fatalf("unexpected flow count: %#v", flows)
	}
	if flows[0].Sink != "http.Get" {
		t.Fatalf("unexpected sink: %#v", flows[0])
	}
	if len(flows[0].Taints) == 0 {
		t.Fatalf("expected taints: %#v", flows[0])
	}
}
