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
	if !Fixture("src/engine/review_test.go") {
		t.Fatal("expected go test file to stay fixture-like")
	}
	if !ReviewArtifact("src/rule/auth.go") {
		t.Fatal("expected detector rule source to stay excluded from self-review")
	}
	if Language("app.ts") != "typescript" {
		t.Fatal("expected typescript language")
	}
}
