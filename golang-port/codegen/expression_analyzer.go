package codegen

import (
	"crypto/sha256"
	"fmt"

	"github.com/quant5-lab/runner/ast"
)

// CallInfo contains metadata about a detected TA function call in an expression
type CallInfo struct {
	Call     *ast.CallExpression // Original AST call node
	FuncName string              // Extracted function name (e.g., "ta.sma")
	ArgHash  string              // Hash of arguments for unique identification
}

// ExpressionAnalyzer traverses AST expressions to find nested TA function calls.
//
// Purpose: Single Responsibility - detect CallExpression nodes in ANY expression context
// Usage: Reusable across BinaryExpression, ConditionalExpression, security(), fixnan()
//
// Example:
//
//	analyzer := NewExpressionAnalyzer(g)
//	calls := analyzer.FindNestedCalls(binaryExpr)
//	// Returns: [CallInfo{sma(close,50)}, CallInfo{sma(close,200)}]
type ExpressionAnalyzer struct {
	gen *generator // Reference to generator for extractFunctionName()
}

// NewExpressionAnalyzer creates analyzer with generator context
func NewExpressionAnalyzer(g *generator) *ExpressionAnalyzer {
	return &ExpressionAnalyzer{gen: g}
}

// FindNestedCalls recursively traverses expression tree to find all CallExpression nodes
func (ea *ExpressionAnalyzer) FindNestedCalls(expr ast.Expression) []CallInfo {
	calls := []CallInfo{}
	ea.traverse(expr, &calls)
	return calls
}

// traverse implements recursive descent through expression AST
func (ea *ExpressionAnalyzer) traverse(expr ast.Expression, calls *[]CallInfo) {
	if expr == nil {
		return
	}

	switch e := expr.(type) {
	case *ast.CallExpression:
		// Found TA function call - extract metadata
		funcName := ea.gen.extractFunctionName(e.Callee)
		argHash := ea.computeArgHash(e)
		*calls = append(*calls, CallInfo{
			Call:     e,
			FuncName: funcName,
			ArgHash:  argHash,
		})
		// Continue traversing arguments (nested calls possible)
		for _, arg := range e.Arguments {
			ea.traverse(arg, calls)
		}

	case *ast.BinaryExpression:
		ea.traverse(e.Left, calls)
		ea.traverse(e.Right, calls)

	case *ast.LogicalExpression:
		ea.traverse(e.Left, calls)
		ea.traverse(e.Right, calls)

	case *ast.ConditionalExpression:
		ea.traverse(e.Test, calls)
		ea.traverse(e.Consequent, calls)
		ea.traverse(e.Alternate, calls)

	case *ast.UnaryExpression:
		ea.traverse(e.Argument, calls)

	case *ast.MemberExpression:
		ea.traverse(e.Object, calls)
		ea.traverse(e.Property, calls)

	case *ast.Identifier, *ast.Literal:
		// Leaf nodes - no traversal needed
		return

	default:
		// Unknown expression type - safe to skip
		return
	}
}

// computeArgHash creates unique identifier for call based on arguments
//
// Purpose: Differentiate sma(close,50) from sma(close,200)
// Method: Hash function name + argument string representations
func (ea *ExpressionAnalyzer) computeArgHash(call *ast.CallExpression) string {
	h := sha256.New()

	// Include function name in hash
	funcName := ea.gen.extractFunctionName(call.Callee)
	h.Write([]byte(funcName))

	// Include each argument
	for _, arg := range call.Arguments {
		argStr := ea.argToString(arg)
		h.Write([]byte(argStr))
	}

	// Return first 8 hex chars (sufficient for uniqueness)
	return fmt.Sprintf("%x", h.Sum(nil))[:8]
}

// argToString converts argument expression to string representation for hashing
func (ea *ExpressionAnalyzer) argToString(expr ast.Expression) string {
	switch e := expr.(type) {
	case *ast.Literal:
		return fmt.Sprintf("%v", e.Value)
	case *ast.Identifier:
		return e.Name
	case *ast.MemberExpression:
		obj := ea.argToString(e.Object)
		prop := ea.argToString(e.Property)
		return obj + "." + prop
	case *ast.CallExpression:
		funcName := ea.gen.extractFunctionName(e.Callee)
		return funcName + "()"
	default:
		return "expr"
	}
}
