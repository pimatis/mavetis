package analyze

import "testing"

func TestSuspiciousPackage(t *testing.T) {
	if SuspiciousPackage("lodas") != "lodash" {
		t.Fatalf("expected lodash typo detection")
	}
	if SuspiciousPackage("lodash") != "" {
		t.Fatalf("expected exact popular package to stay clean")
	}
}
