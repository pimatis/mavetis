package cli

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRunRulesTest(t *testing.T) {
	dir := t.TempDir()
	diffPath := filepath.Join(dir, "sample.diff")
	rulesPath := filepath.Join(dir, "rules.yaml")
	diffContent := `diff --git a/app.js b/app.js
--- a/app.js
+++ b/app.js
@@ -1 +1,2 @@
 const a = 1;
+const secret = "corp_ABCDEFGH";
`
	rulesContent := `rules:
  - id: custom.secret
    title: Custom secret
    message: Custom secret found
    remediation: Remove it
    category: secret
    severity: high
    confidence: high
    target: added
    any:
      - corp_[A-Za-z0-9]{8,}
`
	if err := os.WriteFile(diffPath, []byte(diffContent), 0o600); err != nil {
		t.Fatalf("write diff: %v", err)
	}
	if err := os.WriteFile(rulesPath, []byte(rulesContent), 0o600); err != nil {
		t.Fatalf("write rules: %v", err)
	}
	code := runRules([]string{"test", "--diff", diffPath, "--rules", rulesPath})
	if code != 0 {
		t.Fatalf("expected success, got %d", code)
	}
}

func TestRunRulesRejectsDuplicateIDsAgainstBuiltins(t *testing.T) {
	dir := t.TempDir()
	rulesPath := filepath.Join(dir, "rules.yaml")
	content := `rules:
  - id: secret.jwt
    title: Duplicate builtin
    message: Duplicate builtin id
    category: secret
    require:
      - 'jwt'
`
	if err := os.WriteFile(rulesPath, []byte(content), 0o600); err != nil {
		t.Fatalf("write rules: %v", err)
	}
	code := runRules([]string{"list", "--rules", rulesPath})
	if code != 2 {
		t.Fatalf("expected duplicate id failure, got %d", code)
	}
}

func TestRunRulesMatrix(t *testing.T) {
	code := runRules([]string{"matrix"})
	if code != 0 {
		t.Fatalf("expected matrix success, got %d", code)
	}
}
