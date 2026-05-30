package codegen

import (
	"fmt"

	"github.com/quant5-lab/runner/ast"
)

// OuterScopeCaptureKind classifies how an outer-scope identifier is stored in Go,
// which drives its parameter type, call-site expression, and access resolver role.
type OuterScopeCaptureKind int

const (
	OuterScopeCaptureScalar            OuterScopeCaptureKind = iota // float64 — input constant or numeric literal
	OuterScopeCaptureString                                         // string  — string constant or string variable
	OuterScopeCaptureSeriesFloat                                    // *series.Series — float/bool series variable or input.source
	OuterScopeCaptureArraySeries                                    // *series.ArraySeries — array_series_float variable
	OuterScopeCaptureStringArraySeries                              // *series.StringArraySeries — array_series_string variable
)

// OuterScopeCapture describes a single outer-scope identifier that an arrow function
// reads but cannot access directly — it must be injected as an extra parameter.
type OuterScopeCapture struct {
	Name string                // Pine identifier name (e.g. "price")
	Kind OuterScopeCaptureKind // how the identifier is stored in the outer Go scope
}

// GoParamName returns the Go variable name used both in the function signature
// and as the argument expression at every call site.
func (c OuterScopeCapture) GoParamName() string {
	switch c.Kind {
	case OuterScopeCaptureSeriesFloat:
		return c.Name + "Series"
	case OuterScopeCaptureArraySeries:
		return c.Name + "ArraySeries"
	case OuterScopeCaptureStringArraySeries:
		return c.Name + "StringArraySeries"
	default:
		return c.Name
	}
}

func (c OuterScopeCapture) GoParamType() string {
	switch c.Kind {
	case OuterScopeCaptureSeriesFloat:
		return "*series.Series"
	case OuterScopeCaptureArraySeries:
		return "*series.ArraySeries"
	case OuterScopeCaptureStringArraySeries:
		return "*series.StringArraySeries"
	case OuterScopeCaptureString:
		return "string"
	default:
		return "float64"
	}
}

// GoCallSiteExpression returns the Go expression used at a call site for this capture.
// Bool scalar constants must be bridged to float64 via an IIFE; all other captures
// use GoParamName() directly.
func (c OuterScopeCapture) GoCallSiteExpression(constants map[string]interface{}) string {
	if c.Kind == OuterScopeCaptureScalar {
		if val, ok := constants[c.Name]; ok {
			if _, isBool := val.(bool); isBool {
				return fmt.Sprintf("func() float64 { if %s { return 1.0 } else { return 0.0 } }()", c.Name)
			}
		}
	}
	return c.GoParamName()
}

// NeedsSeriesAccessRegistration reports whether the access resolver must track
// this capture as a series parameter so bare references emit GetCurrent() and
// subscript references emit Get(n).
func (c OuterScopeCapture) NeedsSeriesAccessRegistration() bool {
	return c.Kind == OuterScopeCaptureSeriesFloat
}

// ArrowCaptureRegistry stores the outer-scope captures required by each arrow function.
// Call-site generators query this to inject captured values as extra arguments.
type ArrowCaptureRegistry struct {
	captures map[string][]OuterScopeCapture
}

func NewArrowCaptureRegistry() *ArrowCaptureRegistry {
	return &ArrowCaptureRegistry{captures: make(map[string][]OuterScopeCapture)}
}

func (r *ArrowCaptureRegistry) Register(funcName string, captures []OuterScopeCapture) {
	r.captures[funcName] = captures
}

func (r *ArrowCaptureRegistry) Get(funcName string) []OuterScopeCapture {
	return r.captures[funcName]
}

// OuterScopeCaptureAnalyzer walks an arrow function body to find every identifier
// that is neither a parameter, a local variable, nor a builtin — but IS declared
// in the enclosing strategy scope. Those identifiers must be injected as extra
// parameters so the generated Go function can access them.
type OuterScopeCaptureAnalyzer struct {
	params    map[string]bool
	locals    map[string]bool
	constants map[string]interface{}
	variables map[string]string
}

func NewOuterScopeCaptureAnalyzer(
	params map[string]bool,
	locals map[string]bool,
	constants map[string]interface{},
	variables map[string]string,
) *OuterScopeCaptureAnalyzer {
	return &OuterScopeCaptureAnalyzer{
		params:    params,
		locals:    locals,
		constants: constants,
		variables: variables,
	}
}

func (a *OuterScopeCaptureAnalyzer) Analyze(body []ast.Node) []OuterScopeCapture {
	seen := make(map[string]bool)
	var captures []OuterScopeCapture

	var scanExpr func(ast.Expression)
	var scanStmt func(ast.Node)

	scanExpr = func(expr ast.Expression) {
		if expr == nil {
			return
		}
		switch e := expr.(type) {
		case *ast.Identifier:
			a.tryCapture(e.Name, seen, &captures)
		case *ast.BinaryExpression:
			scanExpr(e.Left)
			scanExpr(e.Right)
		case *ast.UnaryExpression:
			scanExpr(e.Argument)
		case *ast.LogicalExpression:
			scanExpr(e.Left)
			scanExpr(e.Right)
		case *ast.ConditionalExpression:
			scanExpr(e.Test)
			scanExpr(e.Consequent)
			scanExpr(e.Alternate)
		case *ast.CallExpression:
			for _, arg := range e.Arguments {
				scanExpr(arg)
			}
		case *ast.MemberExpression:
			scanExpr(e.Object)
			if e.Computed {
				scanExpr(e.Property)
			}
		case *ast.ForStatement:
			scanExpr(e.From)
			scanExpr(e.To)
			if e.Step != nil {
				scanExpr(e.Step)
			}
			for _, s := range e.Body {
				scanStmt(s)
			}
		case *ast.ForInStatement:
			scanExpr(e.Collection)
			for _, s := range e.Body {
				scanStmt(s)
			}
		case *ast.WhileStatement:
			scanExpr(e.Condition)
			for _, s := range e.Body {
				scanStmt(s)
			}
		case *ast.IfStatement:
			scanExpr(e.Test)
			for _, s := range e.Consequent {
				scanStmt(s)
			}
			for _, s := range e.Alternate {
				scanStmt(s)
			}
		}
	}

	scanStmt = func(node ast.Node) {
		switch s := node.(type) {
		case *ast.ExpressionStatement:
			scanExpr(s.Expression)
		case *ast.VariableDeclaration:
			for _, decl := range s.Declarations {
				if decl.Init != nil {
					scanExpr(decl.Init)
				}
			}
		case *ast.ForStatement:
			scanExpr(s.From)
			scanExpr(s.To)
			if s.Step != nil {
				scanExpr(s.Step)
			}
			for _, bodyStmt := range s.Body {
				scanStmt(bodyStmt)
			}
		case *ast.ForInStatement:
			scanExpr(s.Collection)
			for _, bodyStmt := range s.Body {
				scanStmt(bodyStmt)
			}
		case *ast.WhileStatement:
			scanExpr(s.Condition)
			for _, bodyStmt := range s.Body {
				scanStmt(bodyStmt)
			}
		case *ast.IfStatement:
			scanExpr(s.Test)
			for _, conseq := range s.Consequent {
				scanStmt(conseq)
			}
			for _, alt := range s.Alternate {
				scanStmt(alt)
			}
		}
	}

	for _, stmt := range body {
		scanStmt(stmt)
	}

	return captures
}

// tryCapture resolves name against the outer-scope lookup table and appends a capture
// if found. Constants are checked before variables so that input parameters — which
// appear in both maps — are captured as compile-time scalars, not as series.
func (a *OuterScopeCaptureAnalyzer) tryCapture(name string, seen map[string]bool, captures *[]OuterScopeCapture) {
	if seen[name] || a.params[name] || a.locals[name] {
		return
	}

	if constVal, inConstants := a.constants[name]; inConstants {
		seen[name] = true
		*captures = append(*captures, OuterScopeCapture{Name: name, Kind: kindForConstant(constVal)})
		return
	}

	if varType, inVariables := a.variables[name]; inVariables {
		if kind, capturable := kindForVariable(varType); capturable {
			seen[name] = true
			*captures = append(*captures, OuterScopeCapture{Name: name, Kind: kind})
		}
	}
}

// input.source is stored as a string constant but resolves to *series.Series at runtime.
func kindForConstant(val interface{}) OuterScopeCaptureKind {
	strVal, isString := val.(string)
	if !isString {
		return OuterScopeCaptureScalar
	}
	if strVal == "input.source" {
		return OuterScopeCaptureSeriesFloat
	}
	return OuterScopeCaptureString
}

// kindForVariable maps a generator variable type tag to its capture kind.
// Returns (_, false) for types that are resolved by other mechanisms (e.g. functions
// via callRouter) and therefore must not be injected as extra parameters.
func kindForVariable(varType string) (OuterScopeCaptureKind, bool) {
	switch varType {
	case "float", "float64", "bool":
		return OuterScopeCaptureSeriesFloat, true
	case "string":
		return OuterScopeCaptureString, true
	case "array_series_float":
		return OuterScopeCaptureArraySeries, true
	case "array_series_string":
		return OuterScopeCaptureStringArraySeries, true
	default:
		return 0, false
	}
}
