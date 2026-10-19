package codegen

import "github.com/quant5-lab/runner/ast"

// ChartOnlyUDFDetector identifies user-defined functions whose output is consumed
// exclusively by chart-drawing namespaces (label/line/box/table/linefill), either
// directly in their bodies or transitively through other chart-only UDFs.
// These UDFs are semantically no-ops for strategy execution: their results are
// used only for visual chart overlays which the runner does not render.
type ChartOnlyUDFDetector struct {
	voidHandler *VoidNamespaceCallHandler
}

func NewChartOnlyUDFDetector() *ChartOnlyUDFDetector {
	return &ChartOnlyUDFDetector{voidHandler: &VoidNamespaceCallHandler{}}
}

// Detect returns the set of chart-only UDF names found in program.
// Uses two-phase fixed-point: direct body analysis, then transitive usage analysis.
func (d *ChartOnlyUDFDetector) Detect(program *ast.Program) map[string]bool {
	udfs := d.collectUDFBodies(program)
	chartOnly := make(map[string]bool)

	// Phase 1: mark UDFs whose every body statement is a chart call.
	for name, body := range udfs {
		if d.bodyIsAllChartOnly(body, chartOnly) {
			chartOnly[name] = true
		}
	}

	// Phase 2: iteratively mark UDFs that are only ever called from chart-only
	// contexts — either as arguments to chart-namespace calls, or inside the body
	// of another chart-only UDF.
	for {
		changed := false
		for name := range udfs {
			if chartOnly[name] {
				continue
			}
			if d.allCallSitesAreChartOnly(name, udfs, chartOnly, program) {
				chartOnly[name] = true
				changed = true
			}
		}
		if !changed {
			break
		}
	}

	return chartOnly
}

// collectUDFBodies returns a map from function name to its body statements,
// extracted from top-level arrow-function variable declarations.
func (d *ChartOnlyUDFDetector) collectUDFBodies(program *ast.Program) map[string][]ast.Node {
	result := make(map[string][]ast.Node)
	for _, stmt := range program.Body {
		varDecl, ok := stmt.(*ast.VariableDeclaration)
		if !ok {
			continue
		}
		for _, decl := range varDecl.Declarations {
			id, ok := decl.ID.(*ast.Identifier)
			if !ok {
				continue
			}
			arrow, ok := decl.Init.(*ast.ArrowFunctionExpression)
			if !ok {
				continue
			}
			result[id.Name] = arrow.Body
		}
	}
	return result
}

// bodyIsAllChartOnly returns true if every statement in body produces only
// chart-side-effects or delegates to a known chart-only UDF.
func (d *ChartOnlyUDFDetector) bodyIsAllChartOnly(body []ast.Node, known map[string]bool) bool {
	if len(body) == 0 {
		return false
	}
	for _, stmt := range body {
		if !d.statementIsChartOnly(stmt, known) {
			return false
		}
	}
	return true
}

func (d *ChartOnlyUDFDetector) statementIsChartOnly(stmt ast.Node, known map[string]bool) bool {
	switch s := stmt.(type) {
	case *ast.ExpressionStatement:
		return d.expressionIsChartOnly(s.Expression, known)
	case *ast.VariableDeclaration:
		for _, decl := range s.Declarations {
			if decl.Init != nil && !d.expressionIsChartOnly(decl.Init, known) {
				return false
			}
		}
		return true
	}
	return false
}

func (d *ChartOnlyUDFDetector) expressionIsChartOnly(expr ast.Expression, known map[string]bool) bool {
	if expr == nil {
		return true
	}
	call, ok := expr.(*ast.CallExpression)
	if !ok {
		return false
	}
	funcName := extractCallFunctionName(call)
	return d.voidHandler.CanHandle(funcName) || known[funcName]
}

// allCallSitesAreChartOnly reports whether every occurrence of a call to funcName
// in the entire program appears inside a chart-only context:
//   - as a direct argument to a chart-namespace call, OR
//   - inside the body of a known chart-only UDF.
//
// Returns false if funcName has no call sites (avoid false-positives for dead UDFs).
func (d *ChartOnlyUDFDetector) allCallSitesAreChartOnly(
	funcName string,
	udfs map[string][]ast.Node,
	known map[string]bool,
	program *ast.Program,
) bool {
	finder := &callSiteFinder{funcName: funcName, voidHandler: d.voidHandler, knownChartOnly: known}

	var foundAny bool
	var foundInvalid bool

	for udfName, body := range udfs {
		for _, stmt := range body {
			sites := finder.findInStatement(stmt)
			if len(sites) == 0 {
				continue
			}
			foundAny = true
			if !known[udfName] {
				foundInvalid = true
			}
		}
	}

	for _, stmt := range program.Body {
		if _, isUDF := d.isUDFDeclaration(stmt, udfs); isUDF {
			continue // already scanned above
		}
		sites := finder.findInStatement(stmt)
		if len(sites) == 0 {
			continue
		}
		foundAny = true
		for _, site := range sites {
			if !site.inChartArgPosition {
				foundInvalid = true
			}
		}
	}

	return foundAny && !foundInvalid
}

// isUDFDeclaration reports whether stmt declares a UDF known to udfs.
func (d *ChartOnlyUDFDetector) isUDFDeclaration(stmt ast.Node, udfs map[string][]ast.Node) (string, bool) {
	varDecl, ok := stmt.(*ast.VariableDeclaration)
	if !ok {
		return "", false
	}
	for _, decl := range varDecl.Declarations {
		id, ok := decl.ID.(*ast.Identifier)
		if !ok {
			continue
		}
		if _, known := udfs[id.Name]; known {
			return id.Name, true
		}
	}
	return "", false
}

// callSite records a single occurrence of a target UDF call and whether
// that occurrence is directly inside a chart-namespace call's argument list.
type callSite struct {
	inChartArgPosition bool
}

// callSiteFinder walks AST expressions and collects call sites for a target function.
type callSiteFinder struct {
	funcName       string
	voidHandler    *VoidNamespaceCallHandler
	knownChartOnly map[string]bool
}

func (f *callSiteFinder) findInStatement(stmt ast.Node) []callSite {
	var sites []callSite
	switch s := stmt.(type) {
	case *ast.ExpressionStatement:
		sites = append(sites, f.findInExpr(s.Expression, false)...)
	case *ast.VariableDeclaration:
		for _, decl := range s.Declarations {
			if decl.Init != nil {
				sites = append(sites, f.findInExpr(decl.Init, false)...)
			}
		}
	}
	return sites
}

// findInExpr recursively finds call sites of f.funcName within expr.
// inChartArg is true when the containing call is a chart-namespace or chart-only call.
func (f *callSiteFinder) findInExpr(expr ast.Expression, inChartArg bool) []callSite {
	if expr == nil {
		return nil
	}
	var sites []callSite
	switch e := expr.(type) {
	case *ast.CallExpression:
		callee := extractCallFunctionName(e)
		if callee == f.funcName {
			sites = append(sites, callSite{inChartArgPosition: inChartArg})
		}
		callIsChart := f.voidHandler.CanHandle(callee) || f.knownChartOnly[callee]
		for _, arg := range e.Arguments {
			sites = append(sites, f.findInExpr(arg, callIsChart)...)
		}
	case *ast.BinaryExpression:
		sites = append(sites, f.findInExpr(e.Left, inChartArg)...)
		sites = append(sites, f.findInExpr(e.Right, inChartArg)...)
	case *ast.LogicalExpression:
		sites = append(sites, f.findInExpr(e.Left, inChartArg)...)
		sites = append(sites, f.findInExpr(e.Right, inChartArg)...)
	case *ast.ConditionalExpression:
		sites = append(sites, f.findInExpr(e.Test, inChartArg)...)
		sites = append(sites, f.findInExpr(e.Consequent, inChartArg)...)
		sites = append(sites, f.findInExpr(e.Alternate, inChartArg)...)
	case *ast.UnaryExpression:
		sites = append(sites, f.findInExpr(e.Argument, inChartArg)...)
	case *ast.MemberExpression:
		sites = append(sites, f.findInExpr(e.Object, inChartArg)...)
	}
	return sites
}

// filterChartOnlyCallSites removes call sites for chart-only UDFs from the slice.
// Used to suppress ArrowContext hoisting for UDFs whose call sites are elided to math.NaN().
func filterChartOnlyCallSites(sites []ArrowCallSite, chartOnly map[string]bool) []ArrowCallSite {
	if len(chartOnly) == 0 {
		return sites
	}
	filtered := sites[:0]
	for _, site := range sites {
		if !chartOnly[site.FunctionName] {
			filtered = append(filtered, site)
		}
	}
	return filtered
}
