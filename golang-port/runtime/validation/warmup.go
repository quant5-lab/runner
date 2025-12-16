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
	requirements        []WarmupRequirement
	constantRegistry    *ConstantRegistry
	expressionEvaluator *ExpressionEvaluator
}

// NewWarmupAnalyzer creates a new warmup analyzer
func NewWarmupAnalyzer() *WarmupAnalyzer {
	registry := NewConstantRegistry()
	return &WarmupAnalyzer{
		requirements:        []WarmupRequirement{},
		constantRegistry:    registry,
		expressionEvaluator: NewExpressionEvaluator(registry),
	}
}

// AddConstant adds a constant value for use in expression evaluation
func (w *WarmupAnalyzer) AddConstant(name string, value float64) {
	w.constantRegistry.Set(name, value)
}

func (w *WarmupAnalyzer) AnalyzeScript(program *ast.Program) []WarmupRequirement {
	w.requirements = []WarmupRequirement{}
	w.constantRegistry.Clear()

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
				if id, ok := decl.ID.(*ast.Identifier); ok {
					if val := w.EvaluateConstant(decl.Init); !math.IsNaN(val) {
						w.constantRegistry.Set(id.Name, val)
					}
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
	return w.expressionEvaluator.Evaluate(expr)
}

func (w *WarmupAnalyzer) scanNode(node ast.Node) {
	switch n := node.(type) {
	case *ast.VariableDeclaration:
		for _, decl := range n.Declarations {
			if decl.Init != nil {
				varName := "unknown"
				if id, ok := decl.ID.(*ast.Identifier); ok {
					varName = id.Name
				}
				w.scanExpression(decl.Init, varName)
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

	lookback := w.EvaluateConstant(indexExpr)

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
