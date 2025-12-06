// Package validation provides compile-time analysis of Pine Script strategies
// to detect data requirements before execution.
//
// Problem: Strategies using historical data (e.g., close[1260]) fail silently
// when insufficient bars are provided, producing all-null outputs.
//
// Solution: Static analysis of subscript expressions to determine minimum
// data requirements, enabling early validation and clear error messages.
package validation

import (
	"fmt"
	"math"

	"github.com/quant5-lab/runner/ast"
)

// WarmupRequirement represents data requirements for a strategy
type WarmupRequirement struct {
	// MaxLookback is the maximum historical bars required (e.g., src[nA] where nA=1260)
	MaxLookback int
	// Source describes where the requirement comes from (e.g., "src[nA] at line 15")
	Source string
	// Expression is the original AST expression that caused this requirement
	Expression string
}

// WarmupAnalyzer detects data requirements in Pine Script strategies through
// compile-time constant evaluation. Handles Pine's declaration-before-use
// semantics in a single pass over the AST.
//
// Parser quirk: Variables are wrapped as MemberExpression[0], e.g.,
// "years" becomes MemberExpression(years, Literal(0)). The analyzer unwraps
// these to enable constant propagation across multi-step calculations like
// total = years * days.
type WarmupAnalyzer struct {
	requirements []WarmupRequirement
	constants    map[string]float64
}

// NewWarmupAnalyzer creates a new warmup analyzer
func NewWarmupAnalyzer() *WarmupAnalyzer {
	return &WarmupAnalyzer{
		requirements: []WarmupRequirement{},
		constants:    make(map[string]float64),
	}
}

// AddConstant adds a constant value for use in expression evaluation
func (w *WarmupAnalyzer) AddConstant(name string, value float64) {
	w.constants[name] = value
}

func (w *WarmupAnalyzer) AnalyzeScript(program *ast.Program) []WarmupRequirement {
	w.requirements = []WarmupRequirement{}
	w.constants = make(map[string]float64)

	for _, node := range program.Body {
		w.collectConstants(node)
	}

	for _, node := range program.Body {
		w.scanNode(node)
	}

	return w.requirements
}

// CollectConstants extracts constant values from variable declarations
// Public method for use by codegen package
func (w *WarmupAnalyzer) CollectConstants(node ast.Node) {
	switch n := node.(type) {
	case *ast.VariableDeclaration:
		for _, decl := range n.Declarations {
			if decl.Init != nil {
				if val := w.EvaluateConstant(decl.Init); !math.IsNaN(val) {
					w.constants[decl.ID.Name] = val
				}
			}
		}
	}
}

// collectConstants is internal helper for AnalyzeScript
func (w *WarmupAnalyzer) collectConstants(node ast.Node) {
	w.CollectConstants(node)
}

// EvaluateConstant attempts to evaluate an expression to a constant value
// Public method for use by codegen package
func (w *WarmupAnalyzer) EvaluateConstant(expr ast.Expression) float64 {
	return w.evaluateConstant(expr)
}

// evaluateConstant is internal implementation
func (w *WarmupAnalyzer) evaluateConstant(expr ast.Expression) float64 {
	switch e := expr.(type) {
	case *ast.Literal:
		if v, ok := e.Value.(float64); ok {
			return v
		}
		if v, ok := e.Value.(int); ok {
			return float64(v)
		}
	case *ast.Identifier:
		if val, exists := w.constants[e.Name]; exists {
			return val
		}
	case *ast.MemberExpression:
		if isParserWrappedVariable(e) {
			return w.lookupConstant(e)
		}
		return math.NaN()
	case *ast.BinaryExpression:
		left := w.evaluateConstant(e.Left)
		right := w.evaluateConstant(e.Right)
		if math.IsNaN(left) || math.IsNaN(right) {
			return math.NaN()
		}
		switch e.Operator {
		case "+":
			return left + right
		case "-":
			return left - right
		case "*":
			return left * right
		case "/":
			if right != 0 {
				return left / right
			}
		}
	case *ast.CallExpression:
		return w.evaluateMathCall(e)
	case *ast.ConditionalExpression:
		return math.NaN()
	}
	return math.NaN()
}

func isParserWrappedVariable(e *ast.MemberExpression) bool {
	if !e.Computed {
		return false
	}
	lit, ok := e.Property.(*ast.Literal)
	if !ok {
		return false
	}
	idx, ok := lit.Value.(int)
	return ok && idx == 0
}

func (w *WarmupAnalyzer) lookupConstant(e *ast.MemberExpression) float64 {
	ident, ok := e.Object.(*ast.Identifier)
	if !ok {
		return math.NaN()
	}
	if val, exists := w.constants[ident.Name]; exists {
		return val
	}
	return math.NaN()
}

// evaluateMathCall handles math.pow(), round(), sqrt(), etc.
func (w *WarmupAnalyzer) evaluateMathCall(e *ast.CallExpression) float64 {
	// Extract function name
	funcName := ""
	if member, ok := e.Callee.(*ast.MemberExpression); ok {
		if obj, ok := member.Object.(*ast.Identifier); ok && obj.Name == "math" {
			if prop, ok := member.Property.(*ast.Identifier); ok {
				funcName = prop.Name
			}
		}
	} else if ident, ok := e.Callee.(*ast.Identifier); ok {
		// Pine functions without math. prefix
		funcName = ident.Name
	}

	// Evaluate based on function
	switch funcName {
	case "pow", "math.pow":
		if len(e.Arguments) == 2 {
			base := w.evaluateConstant(e.Arguments[0])
			exp := w.evaluateConstant(e.Arguments[1])
			if !math.IsNaN(base) && !math.IsNaN(exp) {
				return math.Pow(base, exp)
			}
		}
	case "round", "math.round":
		if len(e.Arguments) >= 1 {
			val := w.evaluateConstant(e.Arguments[0])
			if !math.IsNaN(val) {
				return math.Round(val)
			}
		}
	case "sqrt", "math.sqrt":
		if len(e.Arguments) == 1 {
			val := w.evaluateConstant(e.Arguments[0])
			if !math.IsNaN(val) {
				return math.Sqrt(val)
			}
		}
	case "floor", "math.floor":
		if len(e.Arguments) == 1 {
			val := w.evaluateConstant(e.Arguments[0])
			if !math.IsNaN(val) {
				return math.Floor(val)
			}
		}
	case "ceil", "math.ceil":
		if len(e.Arguments) == 1 {
			val := w.evaluateConstant(e.Arguments[0])
			if !math.IsNaN(val) {
				return math.Ceil(val)
			}
		}
	}
	return math.NaN()
}

func (w *WarmupAnalyzer) evaluateMathPow(e *ast.CallExpression) float64 {
	// Legacy method - delegate to evaluateMathCall
	return w.evaluateMathCall(e)
}

func (w *WarmupAnalyzer) scanNode(node ast.Node) {
	switch n := node.(type) {
	case *ast.VariableDeclaration:
		for _, decl := range n.Declarations {
			if decl.Init != nil {
				w.scanExpression(decl.Init, decl.ID.Name)
			}
		}
	case *ast.ExpressionStatement:
		w.scanExpression(n.Expression, "expression")
	case *ast.IfStatement:
		w.scanExpression(n.Test, "if-condition")
		for _, stmt := range n.Consequent {
			w.scanNode(stmt)
		}
		for _, stmt := range n.Alternate {
			w.scanNode(stmt)
		}
	}
}

func (w *WarmupAnalyzer) scanExpression(expr ast.Expression, context string) {
	if expr == nil {
		return
	}

	switch e := expr.(type) {
	case *ast.MemberExpression:
		if e.Computed {
			w.analyzeSubscript(e, context)
		}
		w.scanExpression(e.Object, context)
		w.scanExpression(e.Property, context)
	case *ast.BinaryExpression:
		w.scanExpression(e.Left, context)
		w.scanExpression(e.Right, context)
	case *ast.CallExpression:
		for _, arg := range e.Arguments {
			w.scanExpression(arg, context)
		}
	case *ast.ConditionalExpression:
		w.scanExpression(e.Test, context)
		w.scanExpression(e.Consequent, context)
		w.scanExpression(e.Alternate, context)
	case *ast.UnaryExpression:
		w.scanExpression(e.Argument, context)
	}
}

func (w *WarmupAnalyzer) analyzeSubscript(member *ast.MemberExpression, context string) {
	indexExpr := member.Property

	if nestedMember, ok := indexExpr.(*ast.MemberExpression); ok {
		indexExpr = nestedMember.Object
	}

	lookback := w.evaluateConstant(indexExpr)

	if !math.IsNaN(lookback) && lookback > 0 {
		w.requirements = append(w.requirements, WarmupRequirement{
			MaxLookback: int(lookback),
			Source:      fmt.Sprintf("%s[%.0f] in %s", w.extractVariableName(member.Object), lookback, context),
			Expression:  fmt.Sprintf("%s[%.0f]", w.extractVariableName(member.Object), lookback),
		})
	}
}

func (w *WarmupAnalyzer) extractVariableName(expr ast.Expression) string {
	if ident, ok := expr.(*ast.Identifier); ok {
		return ident.Name
	}
	return "variable"
}

func ValidateDataAvailability(barCount int, requirements []WarmupRequirement) error {
	if len(requirements) == 0 {
		return nil
	}

	maxLookback := 0
	var maxSource string
	for _, req := range requirements {
		if req.MaxLookback > maxLookback {
			maxLookback = req.MaxLookback
			maxSource = req.Source
		}
	}

	if barCount <= maxLookback {
		return fmt.Errorf(
			"insufficient data: need %d+ bars for warmup, have %d bars\n"+
				"  Largest requirement: %s\n"+
				"  Solution: fetch more historical data or reduce rolling period",
			maxLookback+1, barCount, maxSource,
		)
	}

	return nil
}

func GetWarmupInfo(barCount int, requirements []WarmupRequirement) string {
	if len(requirements) == 0 {
		return "No warmup period required"
	}

	maxLookback := 0
	for _, req := range requirements {
		if req.MaxLookback > maxLookback {
			maxLookback = req.MaxLookback
		}
	}

	validBars := barCount - maxLookback
	if validBars < 0 {
		validBars = 0
	}

	return fmt.Sprintf(
		"Warmup: %d bars, Valid output: %d bars (%.1f%%)",
		maxLookback, validBars, float64(validBars)/float64(barCount)*100,
	)
}
