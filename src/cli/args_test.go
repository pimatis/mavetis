package cli

import (
	"strings"
	"testing"
)

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

func TestHelpMessageIncludesDeliveredPhases(t *testing.T) {
	message := helpMessage()
	if !strings.Contains(message, "regression core") {
		t.Fatalf("expected regression help text: %q", message)
	}
	if !strings.Contains(message, "policy layer") {
		t.Fatalf("expected policy help text: %q", message)
	}
	if !strings.Contains(message, "zones.critical") {
		t.Fatalf("expected trust zone help text: %q", message)
	}
}

func TestParseReviewSupportsProfile(t *testing.T) {
	spec, err := parseReview([]string{"--staged", "--profile", "auth"}, false)
	if err != nil {
		t.Fatalf("parse review: %v", err)
	}
	if spec.Profile != "auth" {
		t.Fatalf("unexpected profile: %#v", spec)
	}
}

func TestParseReviewRejectsInvalidFormat(t *testing.T) {
	_, err := parseReview([]string{"--staged", "--format", "xml"}, false)
	if err == nil {
		t.Fatal("expected invalid format error")
	}
}

func TestParseReviewRejectsInvalidProfile(t *testing.T) {
	_, err := parseReview([]string{"--staged", "--profile", "unknown"}, false)
	if err == nil {
		t.Fatal("expected invalid profile error")
	}
}
