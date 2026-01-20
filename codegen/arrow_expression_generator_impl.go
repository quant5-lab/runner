package codegen

import (
	"fmt"

	"github.com/quant5-lab/runner/ast"
)

/*
ArrowExpressionGeneratorImpl generates expressions with Series-aware identifier resolution.

This implementation ensures all local variables in arrow functions resolve to Series access.
*/
type ArrowExpressionGeneratorImpl struct {
	gen                *generator
	accessResolver     *ArrowSeriesAccessResolver
	identifierResolver *ArrowIdentifierResolver
	accessorFactory    *ArrowAwareAccessorFactory
	inlineTAGenerator  *ArrowInlineTACallGenerator
}

func NewArrowExpressionGeneratorImpl(gen *generator, resolver *ArrowSeriesAccessResolver) *ArrowExpressionGeneratorImpl {
	identifierResolver := NewArrowIdentifierResolver(resolver)

	exprGen := &ArrowExpressionGeneratorImpl{
		gen:                gen,
		accessResolver:     resolver,
		identifierResolver: identifierResolver,
	}

	accessorFactory := NewArrowAwareAccessorFactory(identifierResolver, exprGen, gen, gen.symbolTable)
	iifeRegistry := NewInlineTAIIFERegistry()
	inlineTAGenerator := NewArrowInlineTACallGenerator(accessorFactory, iifeRegistry)

	exprGen.accessorFactory = accessorFactory
	exprGen.inlineTAGenerator = inlineTAGenerator

	return exprGen
}

func (e *ArrowExpressionGeneratorImpl) Generate(expr ast.Expression) (string, error) {
	return e.generateExpression(expr)
}

func (e *ArrowExpressionGeneratorImpl) generateExpression(expr ast.Expression) (string, error) {
	switch ex := expr.(type) {
	case *ast.Identifier:
		return e.generateIdentifier(ex)

	case *ast.Literal:
		return e.generateLiteral(ex)

	case *ast.CallExpression:
		return e.generateCallExpression(ex)

	case *ast.BinaryExpression:
		return e.generateBinaryExpression(ex)

	case *ast.UnaryExpression:
		return e.generateUnaryExpression(ex)

	case *ast.LogicalExpression:
		return e.generateLogicalExpression(ex)

	case *ast.ConditionalExpression:
		return e.generateConditionalExpression(ex)

	case *ast.MemberExpression:
		return e.gen.generateMemberExpression(ex)

	default:
		return "", fmt.Errorf("unsupported arrow expression type: %T", expr)
	}
}

func (e *ArrowExpressionGeneratorImpl) generateCallExpression(call *ast.CallExpression) (string, error) {
	// Try inline TA generation first (compile-time constant periods)
	code, handled, err := e.inlineTAGenerator.GenerateInlineTACall(call)
	if err != nil {
		return "", err
	}
	if handled {
		return code, nil
	}

	// Not handled by inline generator
	funcName := extractCallFunctionName(call)

	// Check if it's a TA function - if so, use arrow-aware TA handler directly
	if isTAFunction(funcName) {
		taHandler := NewArrowFunctionTACallGenerator(e.gen, e)
		return taHandler.Generate(call)
	}

	// For non-TA functions (math, user-defined, etc.), try standard call routing
	if e.gen.callRouter != nil {
		routedCode, routeErr := e.gen.callRouter.RouteCall(e.gen, call)
		if routeErr == nil && routedCode != "" {
			return routedCode, nil
		}
	}

	// Final fallback
	return "", fmt.Errorf("unhandled call expression: %s", funcName)
}

func isTAFunction(funcName string) bool {
	switch funcName {
	case "ta.sma", "ta.ema", "ta.rma", "ta.wma", "ta.stdev",
		"ta.highest", "ta.lowest", "ta.change",
		"ta.crossover", "ta.crossunder",
		"ta.pivothigh", "ta.pivotlow",
		"sma", "ema", "rma", "wma", "stdev",
		"highest", "lowest", "change",
		"crossover", "crossunder",
		"fixnan", "ta.fixnan":
		return true
	default:
		return false
	}
}

func (e *ArrowExpressionGeneratorImpl) generateFixnanExpression(call *ast.CallExpression) (string, error) {
	if len(call.Arguments) < 1 {
		return "", fmt.Errorf("fixnan() requires 1 argument")
	}

	sourceExpr := call.Arguments[0]

	sourceCode, err := e.generateExpression(sourceExpr)
	if err != nil {
		return "", fmt.Errorf("fixnan: failed to generate source expression: %w", err)
	}

	return fmt.Sprintf("func() float64 { val := (%s); if math.IsNaN(val) { return 0.0 }; return val }()", sourceCode), nil
}

func (e *ArrowExpressionGeneratorImpl) generateIdentifier(id *ast.Identifier) (string, error) {
	// Try access resolver first (parameters and local variables)
	if access, resolved := e.accessResolver.ResolveAccess(id.Name); resolved {
		return access, nil
	}

	// Try builtin handler (close, high, low, etc.)
	if code, resolved := e.gen.builtinHandler.TryResolveIdentifier(id, false); resolved {
		return code, nil
	}

	// Fallback: direct identifier access (constants, etc.)
	return id.Name, nil
}

func (e *ArrowExpressionGeneratorImpl) generateLiteral(lit *ast.Literal) (string, error) {
	return fmt.Sprintf("%v", lit.Value), nil
}

func (e *ArrowExpressionGeneratorImpl) generateBinaryExpression(binExpr *ast.BinaryExpression) (string, error) {
	left, err := e.generateExpression(binExpr.Left)
	if err != nil {
		return "", err
	}

	right, err := e.generateExpression(binExpr.Right)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("(%s %s %s)", left, binExpr.Operator, right), nil
}

func (e *ArrowExpressionGeneratorImpl) generateUnaryExpression(unaryExpr *ast.UnaryExpression) (string, error) {
	operand, err := e.generateExpression(unaryExpr.Argument)
	if err != nil {
		return "", err
	}

	op := unaryExpr.Operator
	if op == "not" {
		op = "!"
	}

	return fmt.Sprintf("%s%s", op, operand), nil
}

func (e *ArrowExpressionGeneratorImpl) generateLogicalExpression(logExpr *ast.LogicalExpression) (string, error) {
	left, err := e.generateExpression(logExpr.Left)
	if err != nil {
		return "", err
	}

	right, err := e.generateExpression(logExpr.Right)
	if err != nil {
		return "", err
	}

	goOp := logExpr.Operator
	if goOp == "and" {
		goOp = "&&"
	} else if goOp == "or" {
		goOp = "||"
	}

	return fmt.Sprintf("(%s %s %s)", left, goOp, right), nil
}

func (e *ArrowExpressionGeneratorImpl) generateConditionalExpression(condExpr *ast.ConditionalExpression) (string, error) {
	test, err := e.generateExpression(condExpr.Test)
	if err != nil {
		return "", err
	}

	// Add bool conversion if needed
	test = e.gen.addBoolConversionIfNeeded(condExpr.Test, test)

	consequent, err := e.generateExpression(condExpr.Consequent)
	if err != nil {
		return "", err
	}

	alternate, err := e.generateExpression(condExpr.Alternate)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("func() float64 { if %s { return %s } else { return %s } }()",
		test, consequent, alternate), nil
}
