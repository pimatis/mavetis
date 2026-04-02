package analyze

import "testing"

func TestGoCalls(t *testing.T) {
	calls := GoCalls("http.Get(target)\nlog.Print(value)")
	if len(calls) < 2 {
		t.Fatalf("expected parsed calls, got %#v", calls)
	}
	if calls[0] != "http.Get" {
		t.Fatalf("unexpected first call: %#v", calls)
	}
}
