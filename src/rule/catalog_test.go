package rule

import (
	"testing"

	"github.com/Pimatis/mavetis/src/model"
)

func TestBuiltinsExposeStandardsForEveryRule(t *testing.T) {
	rules := Builtins(model.Config{})
	if len(rules) == 0 {
		t.Fatal("expected builtin rules")
	}
	for _, item := range rules {
		if item.ID == "" {
			t.Fatal("expected rule id")
		}
		if item.Title == "" {
			t.Fatalf("expected title for %s", item.ID)
		}
		if len(item.Standards) == 0 {
			t.Fatalf("expected standards for %s", item.ID)
		}
	}
}

func TestBuiltinsExposeExpandedFamilies(t *testing.T) {
	rules := Builtins(model.Config{})
	expected := []string{
		"session.fixation.input",
		"authorization.scope.deleted",
		"oauth.state.disabled",
		"crypto.verify.deleted",
		"supply.remote.dependency",
		"config.csp.disabled",
		"observe.request.body",
		"auth.mfa.deleted",
		"boundary.ui.auth",
		"auth.password.plaintext",
		"auth.password.weakhash",
		"crypto.rsa.keysize",
		"inject.redos",
		"inject.xml.xxee",
		"inject.openredirect",
		"inject.lfi",
		"inject.rfi",
		"file.download.traversal",
		"config.hsts.missing",
		"config.xframe.missing",
		"config.xcontenttype.missing",
		"logic.mass.assignment",
		"logic.price.tampering",
		"secret.pii.exposed",
		"observe.healthdata",
		"webhook.signature.missing",
		"webhook.replay.window.missing",
		"webhook.rawbody.missing",
		"authorization.tenant.lookup.missing",
		"authorization.cross.tenant.query",
		"auth.reset.token.logged",
		"auth.password.change.without.current",
		"cloud.storage.public.read",
		"cloud.storage.public.write",
		"cloud.storage.presigned.longlived",
		"iac.iam.policy.wildcard",
		"iac.securitygroup.open.ssh",
		"ai.prompt.secret.exposure",
		"ai.prompt.user.system.mix",
		"ai.tool.untrusted.input",
	}
	for _, id := range expected {
		if !contains(rules, id) {
			t.Fatalf("expected builtin rule %s", id)
		}
	}
}

func contains(rules []model.Rule, id string) bool {
	for _, item := range rules {
		if item.ID == id {
			return true
		}
	}
	return false
}
