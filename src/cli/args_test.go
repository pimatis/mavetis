package cli

import "testing"

func TestParseReviewSupportsPathAndExplain(t *testing.T) {
	spec, err := parseReview([]string{"--staged", "--path", "src/**", "--explain"}, false)
	if err != nil {
		t.Fatalf("parse review: %v", err)
	}
	if !spec.Staged || spec.Path != "src/**" || !spec.Explain {
		t.Fatalf("unexpected spec: %#v", spec)
	}
}

func TestParseReviewRejectsConflictingModes(t *testing.T) {
	_, err := parseReview([]string{"--staged", "--base", "main"}, false)
	if err == nil {
		t.Fatal("expected conflicting mode error")
	}
}

func TestParseReviewRejectsHeadWithoutBase(t *testing.T) {
	_, err := parseReview([]string{"--head", "feature"}, false)
	if err == nil {
		t.Fatal("expected missing base error")
	}
}

func TestParseReviewRejectsInvalidFormat(t *testing.T) {
	_, err := parseReview([]string{"--staged", "--format", "xml"}, false)
	if err == nil {
		t.Fatal("expected invalid format error")
	}
}
