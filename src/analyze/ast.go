package analyze

import (
	"go/ast"
	"go/parser"
	"go/token"
	"sort"
	"strings"
)

func GoCalls(body string) []string {
	wrapped := "package p\nfunc _() {\n" + body + "\n}\n"
	file, err := parser.ParseFile(token.NewFileSet(), "hunk.go", wrapped, 0)
	if err != nil {
		return nil
	}
	set := map[string]struct{}{}
	ast.Inspect(file, func(node ast.Node) bool {
		call, ok := node.(*ast.CallExpr)
		if !ok {
			return true
		}
		name := goCallName(call.Fun)
		if name == "" {
			return true
		}
		set[name] = struct{}{}
		return true
	})
	values := make([]string, 0, len(set))
	for item := range set {
		values = append(values, item)
	}
	sort.Strings(values)
	return values
}

func goCallName(node ast.Expr) string {
	ident, ok := node.(*ast.Ident)
	if ok {
		return ident.Name
	}
	selector, ok := node.(*ast.SelectorExpr)
	if ok {
		prefix := goCallName(selector.X)
		if prefix == "" {
			return selector.Sel.Name
		}
		return prefix + "." + selector.Sel.Name
	}
	index, ok := node.(*ast.IndexExpr)
	if ok {
		return goCallName(index.X)
	}
	generic, ok := node.(*ast.IndexListExpr)
	if ok {
		return goCallName(generic.X)
	}
	return strings.TrimSpace("")
}
