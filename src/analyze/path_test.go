package analyze

import "testing"

func TestPathHelpers(t *testing.T) {
	if !Executable("service/auth.go") {
		t.Fatal("expected executable go path")
	}
	if Executable("docs/guide.md") {
		t.Fatal("expected markdown to stay non-executable")
	}
	if !Fixture("src/testdata/demo.diff") {
		t.Fatal("expected testdata fixture")
	}
	if Language("app.ts") != "typescript" {
		t.Fatal("expected typescript language")
	}
}
