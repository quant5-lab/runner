package codegen

import (
	"github.com/quant5-lab/runner/ast"
)

// SymbolExpr is keyed by pointer identity — valid because all AST nodes come from
// a single parsed tree so each call site has a unique first-argument pointer.
type securityLHSIndex map[ast.Expression][]string

func buildSecurityLHSIndex(program *ast.Program) securityLHSIndex {
	index := make(securityLHSIndex)
	for _, stmt := range program.Body {
		varDecl, ok := stmt.(*ast.VariableDeclaration)
		if !ok {
			continue
		}
		for _, decl := range varDecl.Declarations {
			call, ok := decl.Init.(*ast.CallExpression)
			if !ok || !IsSecurityFunction(extractCallFunctionName(call)) {
				continue
			}
			if len(call.Arguments) == 0 {
				continue
			}
			index[call.Arguments[0]] = extractPatternNames(decl.ID)
		}
	}
	return index
}

func extractPatternNames(pattern ast.Pattern) []string {
	switch p := pattern.(type) {
	case *ast.Identifier:
		return []string{p.Name}
	case *ast.ArrayPattern:
		names := make([]string, 0, len(p.Elements))
		for _, el := range p.Elements {
			names = append(names, el.Name)
		}
		return names
	}
	return nil
}

// Arrow function bodies are excluded: their locals occupy a separate scope and
// must not propagate outer taint.
func expandVariableTaint(seeds []string, program *ast.Program) map[string]bool {
	tainted := make(map[string]bool, len(seeds))
	for _, s := range seeds {
		tainted[s] = true
	}
	for changed := true; changed; {
		changed = false
		walkControlFlow(program.Body, func(node ast.Node) {
			decl, ok := node.(*ast.VariableDeclaration)
			if !ok {
				return
			}
			for _, d := range decl.Declarations {
				if d.Init == nil || !exprMentionsTainted(d.Init, tainted) {
					continue
				}
				for _, name := range extractPatternNames(d.ID) {
					if !tainted[name] {
						tainted[name] = true
						changed = true
					}
				}
			}
		})
	}
	return tainted
}

// Absence of any consuming call is treated as unknown — the fetch is kept
// conservatively rather than eliminated.
func outputIsChartOnly(tainted map[string]bool, chartOnlyUDFs map[string]bool, program *ast.Program) bool {
	hasAnyUse := false
	hasGoldenUse := false
	chartNS := &VoidNamespaceCallHandler{}
	walkCallExpressions(program.Body, func(call *ast.CallExpression) {
		name := extractCallFunctionName(call)
		for _, arg := range call.Arguments {
			if !exprMentionsTainted(arg, tainted) {
				continue
			}
			hasAnyUse = true
			if !chartOnlyUDFs[name] && !chartNS.CanHandle(name) {
				hasGoldenUse = true
			}
		}
	})
	return hasAnyUse && !hasGoldenUse
}

func exprMentionsTainted(expr ast.Expression, tainted map[string]bool) bool {
	if expr == nil {
		return false
	}
	switch e := expr.(type) {
	case *ast.Identifier:
		return tainted[e.Name]
	case *ast.CallExpression:
		for _, arg := range e.Arguments {
			if exprMentionsTainted(arg, tainted) {
				return true
			}
		}
	case *ast.BinaryExpression:
		return exprMentionsTainted(e.Left, tainted) || exprMentionsTainted(e.Right, tainted)
	case *ast.LogicalExpression:
		return exprMentionsTainted(e.Left, tainted) || exprMentionsTainted(e.Right, tainted)
	case *ast.ConditionalExpression:
		return exprMentionsTainted(e.Test, tainted) ||
			exprMentionsTainted(e.Consequent, tainted) ||
			exprMentionsTainted(e.Alternate, tainted)
	case *ast.UnaryExpression:
		return exprMentionsTainted(e.Argument, tainted)
	case *ast.MemberExpression:
		if exprMentionsTainted(e.Object, tainted) {
			return true
		}
		// Computed access (arr[taintedIdx]) — the index is a consuming reference.
		if e.Computed {
			return exprMentionsTainted(e.Property, tainted)
		}
		return false
	}
	return false
}

// Stops at arrow function boundaries so UDF locals are never conflated with
// outer-scope tainted names.
func walkControlFlow(nodes []ast.Node, visit func(ast.Node)) {
	for _, node := range nodes {
		visit(node)
		switch s := node.(type) {
		case *ast.IfStatement:
			walkControlFlow(s.Consequent, visit)
			walkControlFlow(s.Alternate, visit)
		case *ast.ForStatement:
			walkControlFlow(s.Body, visit)
		case *ast.ForInStatement:
			walkControlFlow(s.Body, visit)
		case *ast.WhileStatement:
			walkControlFlow(s.Body, visit)
		}
	}
}

// Arrow bodies are entered (unlike walkControlFlow) so top-level UDF bodies are
// scanned; the chartOnlyUDFs guard in callers prevents false golden-sink signals
// from their internal chart calls.
func walkCallExpressions(nodes []ast.Node, visit func(*ast.CallExpression)) {
	walkControlFlow(nodes, func(node ast.Node) {
		switch s := node.(type) {
		case *ast.ExpressionStatement:
			walkExprCalls(s.Expression, visit)
		case *ast.VariableDeclaration:
			for _, d := range s.Declarations {
				walkExprCalls(d.Init, visit)
				if arrow, ok := d.Init.(*ast.ArrowFunctionExpression); ok {
					walkCallExpressions(arrow.Body, visit)
				}
			}
		case *ast.IfStatement:
			walkExprCalls(s.Test, visit)
		case *ast.ForStatement:
			walkExprCalls(s.From, visit)
			walkExprCalls(s.To, visit)
			walkExprCalls(s.Step, visit)
		case *ast.ForInStatement:
			walkExprCalls(s.Collection, visit)
		case *ast.WhileStatement:
			walkExprCalls(s.Condition, visit)
		}
	})
}

func walkExprCalls(expr ast.Expression, visit func(*ast.CallExpression)) {
	if expr == nil {
		return
	}
	switch e := expr.(type) {
	case *ast.CallExpression:
		visit(e)
		for _, arg := range e.Arguments {
			walkExprCalls(arg, visit)
		}
	case *ast.BinaryExpression:
		walkExprCalls(e.Left, visit)
		walkExprCalls(e.Right, visit)
	case *ast.LogicalExpression:
		walkExprCalls(e.Left, visit)
		walkExprCalls(e.Right, visit)
	case *ast.ConditionalExpression:
		walkExprCalls(e.Test, visit)
		walkExprCalls(e.Consequent, visit)
		walkExprCalls(e.Alternate, visit)
	case *ast.UnaryExpression:
		walkExprCalls(e.Argument, visit)
	case *ast.MemberExpression:
		walkExprCalls(e.Object, visit)
	}
}

// SecurityChartOnlyClassifier classifies resolved security() calls as chart-only
// when every consuming use of their output is within chart-only namespaces.
type SecurityChartOnlyClassifier struct{}

// ChartOnlySecurityKeys returns keys for which every resolved call is chart-only.
// A key is omitted unless ALL calls sharing it are chart-only, so a symbol used
// in both a label and a strategy.entry always keeps its fetch.
func (SecurityChartOnlyClassifier) ChartOnlySecurityKeys(
	resolved []resolvedSecurityCall,
	lhsIndex securityLHSIndex,
	chartOnlyUDFs map[string]bool,
	program *ast.Program,
) map[string]bool {
	callChartOnly := make(map[ast.Expression]bool, len(resolved))
	for _, r := range resolved {
		symExpr := r.call.SymbolExpr
		names, found := lhsIndex[symExpr]
		if !found || len(names) == 0 {
			callChartOnly[symExpr] = false
			continue
		}
		tainted := expandVariableTaint(names, program)
		callChartOnly[symExpr] = outputIsChartOnly(tainted, chartOnlyUDFs, program)
	}

	allChartOnly := make(map[string]bool)
	seen := make(map[string]bool)
	for _, r := range resolved {
		key := resolvedDedupKey(r)
		if !seen[key] {
			allChartOnly[key] = true
			seen[key] = true
		}
		if !callChartOnly[r.call.SymbolExpr] {
			allChartOnly[key] = false
		}
	}

	result := make(map[string]bool)
	for key, ok := range allChartOnly {
		if ok {
			result[key] = true
		}
	}
	return result
}

func resolvedDedupKey(r resolvedSecurityCall) string {
	sym := r.resolvedSym
	if r.isSymRuntime {
		sym = runtimePlaceholder()
	}
	tf := r.resolvedTf
	if r.isTfRuntime {
		tf = runtimePlaceholder()
	}
	return sym + ":" + tf
}
