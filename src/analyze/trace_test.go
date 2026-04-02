package analyze

import "testing"

func TestTrackPropagatesAssignments(t *testing.T) {
	steps := Track([]string{
		`target := ctx.Query("url")`,
		`copy = target`,
		`http.Get(copy)`,
	})
	if len(steps) != 3 {
		t.Fatalf("unexpected step count: %d", len(steps))
	}
	if len(steps[2].Taints) == 0 {
		t.Fatalf("expected propagated taints: %#v", steps[2])
	}
}
