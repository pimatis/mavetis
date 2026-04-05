package cli

import (
	"os"
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
	if !strings.Contains(message, "file review") {
		t.Fatalf("expected file review help text: %q", message)
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

func TestParseReviewSupportsFileTargets(t *testing.T) {
	spec, err := parseReview([]string{"@src/app.go", "src/auth.ts", "--format", "json"}, false)
	if err != nil {
		t.Fatalf("parse review: %v", err)
	}
	if spec.Mode != "file" {
		t.Fatalf("unexpected mode: %#v", spec)
	}
	if len(spec.Files) != 2 || spec.Files[0] != "src/app.go" || spec.Files[1] != "src/auth.ts" {
		t.Fatalf("unexpected files: %#v", spec.Files)
	}
}

func TestParseReviewRejectsFileTargetsWithStaged(t *testing.T) {
	_, err := parseReview([]string{"@src/app.go", "--staged"}, false)
	if err == nil {
		t.Fatal("expected mutual exclusion error")
	}
}

func TestParseReviewRejectsFileTargetsInCI(t *testing.T) {
	_, err := parseReview([]string{"@src/app.go"}, true)
	if err == nil {
		t.Fatal("expected ci rejection")
	}
}

func TestParseReviewSupportsPlainRelativeTargets(t *testing.T) {
	spec, err := parseReview([]string{"src/app.go", "src/auth.ts"}, false)
	if err != nil {
		t.Fatalf("parse review: %v", err)
	}
	if spec.Mode != "file" {
		t.Fatalf("unexpected mode: %#v", spec)
	}
	if len(spec.Files) != 2 || spec.Files[0] != "src/app.go" || spec.Files[1] != "src/auth.ts" {
		t.Fatalf("unexpected files: %#v", spec.Files)
	}
}

func TestParseReviewSupportsFlagsAfterPlainTarget(t *testing.T) {
	spec, err := parseReview([]string{"src/app.go", "--format", "json", "--explain"}, false)
	if err != nil {
		t.Fatalf("parse review: %v", err)
	}
	if spec.Format != "json" || !spec.Explain {
		t.Fatalf("unexpected spec: %#v", spec)
	}
}

func TestParseReviewSupportsWithSuggested(t *testing.T) {
	spec, err := parseReview([]string{"src/app.go", "--with-suggested"}, false)
	if err != nil {
		t.Fatalf("parse review: %v", err)
	}
	if !spec.WithSuggested {
		t.Fatalf("expected with suggested: %#v", spec)
	}
}

func TestParseReviewRejectsWithSuggestedWithoutFiles(t *testing.T) {
	_, err := parseReview([]string{"--with-suggested"}, false)
	if err == nil {
		t.Fatal("expected with-suggested validation error")
	}
}

func TestParseReviewSupportsStdinTargets(t *testing.T) {
	stdin, err := os.CreateTemp(t.TempDir(), "targets.txt")
	if err != nil {
		t.Fatalf("create temp: %v", err)
	}
	if _, err := stdin.WriteString("src/app.go\nsrc/auth.ts\n"); err != nil {
		t.Fatalf("write temp: %v", err)
	}
	if _, err := stdin.Seek(0, 0); err != nil {
		t.Fatalf("seek temp: %v", err)
	}
	previous := os.Stdin
	os.Stdin = stdin
	defer func() {
		os.Stdin = previous
		_ = stdin.Close()
	}()
	spec, err := parseReview([]string{"--stdin-targets"}, false)
	if err != nil {
		t.Fatalf("parse review: %v", err)
	}
	if spec.Mode != "file" {
		t.Fatalf("unexpected mode: %#v", spec)
	}
	if len(spec.Files) != 2 || spec.Files[0] != "src/app.go" || spec.Files[1] != "src/auth.ts" {
		t.Fatalf("unexpected files: %#v", spec.Files)
	}
}

func TestParseReviewRejectsEmptyAtTarget(t *testing.T) {
	_, err := parseReview([]string{"@"}, false)
	if err == nil {
		t.Fatal("expected empty target rejection")
	}
}
