package analyze

import (
	"go/ast"
	"go/parser"
	"go/token"
	"strings"
)

type Flow struct {
	Sink   string
	Line   int
	Taints []string
}

func GoFlows(body string) []Flow {
	wrapped := "package p\nfunc _() {\n" + body + "\n}\n"
	file, err := parser.ParseFile(token.NewFileSet(), "flow.go", wrapped, parser.ParseComments)
	if err != nil {
		return nil
	}
	state := map[string]struct{}{}
	flows := make([]Flow, 0)
	ast.Inspect(file, func(node ast.Node) bool {
		stmt, ok := node.(ast.Stmt)
		if !ok {
			return true
		}
		assign, ok := stmt.(*ast.AssignStmt)
		if ok {
			taints := exprTaints(assign.Rhs, state)
			for _, left := range assign.Lhs {
				ident, ok := left.(*ast.Ident)
				if !ok {
					continue
				}
				if len(taints) == 0 {
					delete(state, strings.ToLower(ident.Name))
					continue
				}
				state[strings.ToLower(ident.Name)] = struct{}{}
			}
			return true
		}
		exprStmt, ok := stmt.(*ast.ExprStmt)
		if ok {
			call, ok := exprStmt.X.(*ast.CallExpr)
			if ok {
				flow := callFlow(call, state)
				if flow.Sink != "" {
					flows = append(flows, flow)
				}
			}
			return true
		}
		return true
	})
	return flows
}

func exprTaints(expressions []ast.Expr, state map[string]struct{}) []string {
	set := map[string]struct{}{}
	for _, expression := range expressions {
		collectTaints(expression, state, set)
	}
	values := make([]string, 0, len(set))
	for item := range set {
		values = append(values, item)
	}
	return values
}

func collectTaints(node ast.Expr, state map[string]struct{}, set map[string]struct{}) {
	ident, ok := node.(*ast.Ident)
	if ok {
		name := strings.ToLower(ident.Name)
		if _, ok := state[name]; ok {
			set[name] = struct{}{}
		}
		return
	}
	call, ok := node.(*ast.CallExpr)
	if ok {
		name := goCallName(call.Fun)
		lower := strings.ToLower(name)
		if strings.Contains(lower, "query") || strings.Contains(lower, "param") || strings.Contains(lower, "header") || strings.Contains(lower, "cookie") || strings.Contains(lower, "body") {
			set[lower] = struct{}{}
		}
		for _, argument := range call.Args {
			collectTaints(argument, state, set)
		}
		return
	}
	binary, ok := node.(*ast.BinaryExpr)
	if ok {
		collectTaints(binary.X, state, set)
		collectTaints(binary.Y, state, set)
		return
	}
	selector, ok := node.(*ast.SelectorExpr)
	if ok {
		collectTaints(selector.X, state, set)
	}
}

func callFlow(call *ast.CallExpr, state map[string]struct{}) Flow {
	flow := Flow{}
	flow.Sink = goCallName(call.Fun)
	if flow.Sink == "" {
		return flow
	}
	lower := strings.ToLower(flow.Sink)
	if !strings.Contains(lower, "http.get") && !strings.Contains(lower, "exec.command") && !strings.Contains(lower, "os.open") && !strings.Contains(lower, "template.new") {
		return Flow{}
	}
	flow.Taints = exprTaints(call.Args, state)
	if len(flow.Taints) == 0 {
		return Flow{}
	}
	return flow
}

func GoUnsafePointer(body string) bool {
	wrapped := "package p\nfunc _() {\n" + body + "\n}\n"
	file, err := parser.ParseFile(token.NewFileSet(), "unsafe.go", wrapped, 0)
	if err != nil {
		return false
	}
	found := false
	ast.Inspect(file, func(node ast.Node) bool {
		call, ok := node.(*ast.CallExpr)
		if ok {
			name := goCallName(call.Fun)
			if strings.EqualFold(name, "unsafe.Pointer") {
				found = true
				return false
			}
		}
		return true
	})
	return found
}
