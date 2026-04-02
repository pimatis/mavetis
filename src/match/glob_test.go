package match

import "testing"

func TestGlob(t *testing.T) {
	cases := []struct {
		pattern string
		value   string
		match   bool
	}{
		{pattern: "src/**/*.go", value: "src/engine/review.go", match: true},
		{pattern: "**/*.md", value: "README.md", match: true},
		{pattern: "test/*", value: "test/a/b", match: false},
	}
	for _, item := range cases {
		result := Glob(item.pattern, item.value)
		if result != item.match {
			t.Fatalf("unexpected result for %s %s", item.pattern, item.value)
		}
	}
}
