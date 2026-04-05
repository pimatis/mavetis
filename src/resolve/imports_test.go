package resolve

import (
	"strings"
	"testing"
)

func TestGoImports(t *testing.T) {
	content := "package main\nimport (\n\t\"fmt\"\n\t\"github.com/Pimatis/mavetis/src/scan\"\n)\n"
	refs := Imports("app.go", content)
	if len(refs) != 2 {
		t.Fatalf("unexpected refs: %#v", refs)
	}
	if refs[1].Module != "github.com/Pimatis/mavetis/src/scan" {
		t.Fatalf("unexpected ref: %#v", refs[1])
	}
}

func TestJSImports(t *testing.T) {
	content := strings.Join([]string{
		"import auth from './auth'",
		"const dep = require('../dep')",
		"export { run } from './runner'",
		"await import('./lazy')",
	}, "\n")
	refs := Imports("app.ts", content)
	if len(refs) != 4 {
		t.Fatalf("unexpected refs: %#v", refs)
	}
	if refs[0].Module != "./auth" || refs[1].Kind != "require" || refs[2].Kind != "export-from" || refs[3].Kind != "dynamic-import" {
		t.Fatalf("unexpected refs: %#v", refs)
	}
}

func TestPythonImports(t *testing.T) {
	content := strings.Join([]string{
		"from . import utils",
		"from pkg.auth import verify",
		"import os, app.service as service",
	}, "\n")
	refs := Imports("app.py", content)
	if len(refs) != 4 {
		t.Fatalf("unexpected refs: %#v", refs)
	}
	if refs[0].Module != ".utils" || refs[1].Module != "pkg.auth" || refs[3].Module != "app.service" {
		t.Fatalf("unexpected refs: %#v", refs)
	}
}

func TestJavaImports(t *testing.T) {
	content := "import java.util.List;\nimport com.example.auth.Service;\n"
	refs := Imports("App.java", content)
	if len(refs) != 2 {
		t.Fatalf("unexpected refs: %#v", refs)
	}
	if refs[1].Module != "com.example.auth.Service" {
		t.Fatalf("unexpected ref: %#v", refs[1])
	}
}

func TestImportsEdgeCap(t *testing.T) {
	builder := strings.Builder{}
	for index := 0; index < 80; index++ {
		builder.WriteString("import item")
		builder.WriteString(string(rune('A' + (index % 26))))
		builder.WriteString(" from './mod")
		builder.WriteString(strings.Repeat("x", index%3))
		builder.WriteString("'\n")
	}
	refs := Imports("many.ts", builder.String())
	if len(refs) != 64 {
		t.Fatalf("unexpected ref count: %d", len(refs))
	}
}

func TestImportsUnknownLanguage(t *testing.T) {
	refs := Imports("README.md", "import './auth'")
	if len(refs) != 0 {
		t.Fatalf("expected no refs, got %#v", refs)
	}
}
