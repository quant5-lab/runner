package codegen

import (
	"crypto/sha256"
	"fmt"

	"github.com/quant5-lab/runner/ast"
)

type CallInfo struct {
	Call      *ast.CallExpression
	FuncName  string
	ArgHash   string
	StmtIndex int /* top-level statement index for per-statement calc emission */
}

// ExpressionAnalyzer traverses AST expressions to find nested TA function calls.
// Reusable across BinaryExpression, ConditionalExpression, security(), fixnan()
type ExpressionAnalyzer struct {
	gen *generator
}

func NewExpressionAnalyzer(g *generator) *ExpressionAnalyzer {
	return &ExpressionAnalyzer{gen: g}
}

// FindNestedCalls recursively finds all CallExpression nodes in expression tree
func (ea *ExpressionAnalyzer) FindNestedCalls(expr ast.Expression) []CallInfo {
	calls := []CallInfo{}
	ea.traverse(expr, &calls)
	return calls
}

// IsInsideSecurityCall detects if targetCall is nested inside security() call
func (ea *ExpressionAnalyzer) IsInsideSecurityCall(targetCall *ast.CallExpression, rootExpr ast.Expression) bool {
	return ea.findSecurityCallContaining(targetCall, rootExpr)
}

func (ea *ExpressionAnalyzer) findSecurityCallContaining(targetCall *ast.CallExpression, expr ast.Expression) bool {
	if expr == nil {
		return false
	}

	switch e := expr.(type) {
	case *ast.CallExpression:
		if ea.isSecurityCall(e) && len(e.Arguments) >= 3 {
			return ea.expressionContainsCall(targetCall, e.Arguments[2])
		}
		return ea.anyChildMatches(e.Arguments, func(arg ast.Expression) bool {
			return ea.findSecurityCallContaining(targetCall, arg)
		})

	case *ast.BinaryExpression:
		return ea.findSecurityCallContaining(targetCall, e.Left) ||
			ea.findSecurityCallContaining(targetCall, e.Right)

	case *ast.LogicalExpression:
		return ea.findSecurityCallContaining(targetCall, e.Left) ||
			ea.findSecurityCallContaining(targetCall, e.Right)

	case *ast.ConditionalExpression:
		return ea.findSecurityCallContaining(targetCall, e.Test) ||
			ea.findSecurityCallContaining(targetCall, e.Consequent) ||
			ea.findSecurityCallContaining(targetCall, e.Alternate)

	case *ast.UnaryExpression:
		return ea.findSecurityCallContaining(targetCall, e.Argument)

	case *ast.MemberExpression:
		return ea.findSecurityCallContaining(targetCall, e.Object) ||
			ea.findSecurityCallContaining(targetCall, e.Property)
	}

	return false
}

func (ea *ExpressionAnalyzer) isSecurityCall(call *ast.CallExpression) bool {
	funcName := ea.gen.extractFunctionName(call.Callee)
	return funcName == "security" || funcName == "request.security"
}

func (ea *ExpressionAnalyzer) expressionContainsCall(targetCall *ast.CallExpression, expr ast.Expression) bool {
	if expr == nil {
		return false
	}

	if callExpr, ok := expr.(*ast.CallExpression); ok {
		if callExpr == targetCall {
			return true
		}
		if ea.anyChildMatches(callExpr.Arguments, func(arg ast.Expression) bool {
			return ea.expressionContainsCall(targetCall, arg)
		}) {
			return true
		}
	}

	return ea.traverseExpression(expr, func(child ast.Expression) bool {
		return ea.expressionContainsCall(targetCall, child)
	})
}

func (ea *ExpressionAnalyzer) traverseExpression(expr ast.Expression, visitor func(ast.Expression) bool) bool {
	switch e := expr.(type) {
	case *ast.BinaryExpression:
		return visitor(e.Left) || visitor(e.Right)
	case *ast.LogicalExpression:
		return visitor(e.Left) || visitor(e.Right)
	case *ast.ConditionalExpression:
		return visitor(e.Test) || visitor(e.Consequent) || visitor(e.Alternate)
	case *ast.UnaryExpression:
		return visitor(e.Argument)
	case *ast.MemberExpression:
		return visitor(e.Object) || visitor(e.Property)
	}
	return false
}

func (ea *ExpressionAnalyzer) anyChildMatches(exprs []ast.Expression, visitor func(ast.Expression) bool) bool {
	for _, expr := range exprs {
		if visitor(expr) {
			return true
		}
	}
	return false
}

func (ea *ExpressionAnalyzer) traverse(expr ast.Expression, calls *[]CallInfo) {
	if expr == nil {
		return
	}

	switch e := expr.(type) {
	case *ast.CallExpression:
		funcName := ea.gen.extractFunctionName(e.Callee)
		argHash := ea.ComputeArgHash(e)
		*calls = append(*calls, CallInfo{
			Call:     e,
			FuncName: funcName,
			ArgHash:  argHash,
		})
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
		return

	default:
		return
	}
}

// ComputeArgHash creates unique identifier for call based on function name and arguments.
// Differentiates sma(close,50) from sma(close,200) for temp variable registration.
func (ea *ExpressionAnalyzer) ComputeArgHash(call *ast.CallExpression) string {
	h := sha256.New()

	funcName := ea.gen.extractFunctionName(call.Callee)
	h.Write([]byte(funcName))

	for _, arg := range call.Arguments {
		argStr := ea.argToString(arg)
		h.Write([]byte(argStr))
	}

	return fmt.Sprintf("%x", h.Sum(nil))[:8]
}

func (ea *ExpressionAnalyzer) argToString(expr ast.Expression) string {
	switch e := expr.(type) {
	case *ast.Literal:
		return fmt.Sprintf("%v", e.Value)
	case *ast.Identifier:
		return e.Name
	case *ast.MemberExpression:
		obj := ea.argToString(e.Object)
		prop := ea.argToString(e.Property)
		if e.Computed {
			return obj + "[" + prop + "]"
		}
		return obj + "." + prop
	case *ast.CallExpression:
		funcName := ea.gen.extractFunctionName(e.Callee)
		args := ""
		for i, arg := range e.Arguments {
			if i > 0 {
				args += ","
			}
			args += ea.argToString(arg)
		}
		return funcName + "(" + args + ")"
	case *ast.BinaryExpression:
		left := ea.argToString(e.Left)
		right := ea.argToString(e.Right)
		return "(" + left + e.Operator + right + ")"
	case *ast.UnaryExpression:
		operand := ea.argToString(e.Argument)
		return e.Operator + operand
	default:
		return "expr"
	}
}
