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
	valueGenerator     *ArrowValueFunctionGenerator
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
	valueGenerator := NewArrowValueFunctionGenerator(exprGen)

	exprGen.accessorFactory = accessorFactory
	exprGen.inlineTAGenerator = inlineTAGenerator
	exprGen.valueGenerator = valueGenerator

	return exprGen
}

func (e *ArrowExpressionGeneratorImpl) Generate(expr ast.Expression) (string, error) {
	return e.generateExpression(expr)
}

func (e *ArrowExpressionGeneratorImpl) generateExpression(expr ast.Expression) (string, error) {
	switch ex := expr.(type) {
	case *ast.ForStatement:
		cfGenerator := NewControlFlowExpressionGenerator(e.gen)
		return cfGenerator.GenerateForExpressionAsIIFE(ex)

	case *ast.IfStatement:
		cfGenerator := NewControlFlowExpressionGenerator(e.gen)
		return cfGenerator.GenerateIfExpressionAsIIFE(ex)

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
		return e.generateMemberExpression(ex)

	default:
		return "", fmt.Errorf("unsupported arrow expression type: %T", expr)
	}
}

func (e *ArrowExpressionGeneratorImpl) generateCallExpression(call *ast.CallExpression) (string, error) {
	funcName := extractCallFunctionName(call)

	if e.valueGenerator.CanHandle(funcName) {
		return e.valueGenerator.Generate(call)
	}

	code, handled, err := e.inlineTAGenerator.GenerateInlineTACall(call)
	if err != nil {
		return "", err
	}
	if handled {
		return code, nil
	}

	if isTAFunction(funcName) {
		taHandler := NewArrowFunctionTACallGenerator(e.gen, e)
		return taHandler.Generate(call)
	}

	if e.gen.callRouter != nil {
		routedCode, routeErr := e.gen.callRouter.RouteCall(e.gen, call)
		if routeErr != nil {
			return "", routeErr
		}
		if routedCode != "" {
			return routedCode, nil
		}
	}

	return "", fmt.Errorf("unhandled call expression: %s", funcName)
}

func isTAFunction(funcName string) bool {
	if len(funcName) > 3 && funcName[:3] == "ta." {
		return true
	}
	switch funcName {
	case "sma", "ema", "rma", "wma", "stdev",
		"highest", "lowest", "change",
		"crossover", "crossunder",
		"rsi":
		return true
	default:
		return false
	}
}

func (e *ArrowExpressionGeneratorImpl) generateIdentifier(id *ast.Identifier) (string, error) {
	/* Loop counters need float64() cast for arithmetic compatibility */
	if e.gen.loopContextStack != nil && e.gen.loopContextStack.IsLoopCounter(id.Name) {
		return fmt.Sprintf("float64(%s)", id.Name), nil
	}

	if access, resolved := e.accessResolver.ResolveAccess(id.Name); resolved {
		return access, nil
	}

	if code, resolved := e.gen.builtinHandler.TryResolveIdentifier(id, false); resolved {
		return code, nil
	}

	return id.Name, nil
}

func (e *ArrowExpressionGeneratorImpl) generateLiteral(lit *ast.Literal) (string, error) {
	return fmt.Sprintf("%v", lit.Value), nil
}

func (e *ArrowExpressionGeneratorImpl) generateNumericExpression(expr ast.Expression) (string, error) {
	if lit, ok := expr.(*ast.Literal); ok {
		if boolVal, ok := lit.Value.(bool); ok {
			if boolVal {
				return "1.0", nil
			}
			return "0.0", nil
		}
	}

	return e.generateExpression(expr)
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

	test = e.gen.addBoolConversionIfNeeded(condExpr.Test, test)

	consequent, err := e.generateNumericExpression(condExpr.Consequent)
	if err != nil {
		return "", err
	}

	alternate, err := e.generateNumericExpression(condExpr.Alternate)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("func() float64 { if %s { return %s } else { return %s } }()",
		test, consequent, alternate), nil
}

func (e *ArrowExpressionGeneratorImpl) generateMemberExpression(mem *ast.MemberExpression) (string, error) {
	if mem.Computed {
		if obj, ok := mem.Object.(*ast.Identifier); ok {
			return e.resolveArrowSubscript(obj.Name, mem.Property)
		}
	}

	if code, resolved := e.gen.builtinHandler.TryResolveMemberExpression(mem, false); resolved {
		return code, nil
	}

	if obj, ok := mem.Object.(*ast.Identifier); ok {
		if prop, ok := mem.Property.(*ast.Identifier); ok {
			if obj.Name == "strategy" {
				switch prop.Name {
				case "long":
					return "strategy.Long", nil
				case "short":
					return "strategy.Short", nil
				}
			}
			return obj.Name + "." + prop.Name, nil
		}
	}

	return "", fmt.Errorf("unsupported member expression pattern")
}

/*
resolveArrowSubscript generates Series.Get() or builtin bounds-checked access.

Handles series parameters, local series variables, and builtin series (close/open/etc).
*/
func (e *ArrowExpressionGeneratorImpl) resolveArrowSubscript(seriesName string, indexExpr ast.Expression) (string, error) {
	isSeriesParam := e.accessResolver.IsParameter(seriesName)
	isBuiltin := seriesName == "close" || seriesName == "open" || seriesName == "high" || seriesName == "low" || seriesName == "volume"

	indexCode, err := e.generateArrowIndexExpression(indexExpr)
	if err != nil {
		return "", fmt.Errorf("failed to generate index expression: %w", err)
	}

	if isBuiltin {
		return e.generateBuiltinSubscript(seriesName, indexCode), nil
	}

	if isSeriesParam {
		return fmt.Sprintf("%sSeries.Get(int(%s))", seriesName, indexCode), nil
	}

	return fmt.Sprintf("%sSeries.Get(int(%s))", seriesName, indexCode), nil
}

func (e *ArrowExpressionGeneratorImpl) generateArrowIndexExpression(expr ast.Expression) (string, error) {
	switch ex := expr.(type) {
	case *ast.Identifier:
		/* Loop counters cast to float64 for arithmetic with other floats */
		if e.gen.loopContextStack != nil && e.gen.loopContextStack.IsLoopCounter(ex.Name) {
			return fmt.Sprintf("float64(%s)", ex.Name), nil
		}
		if access, resolved := e.accessResolver.ResolveAccess(ex.Name); resolved {
			return access, nil
		}
		return ex.Name, nil

	case *ast.Literal:
		return e.generateLiteral(ex)

	case *ast.BinaryExpression:
		left, err := e.generateArrowIndexExpression(ex.Left)
		if err != nil {
			return "", err
		}
		right, err := e.generateArrowIndexExpression(ex.Right)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("(%s %s %s)", left, ex.Operator, right), nil

	default:
		return e.generateExpression(expr)
	}
}

func (e *ArrowExpressionGeneratorImpl) generateBuiltinSubscript(seriesName, indexCode string) string {
	capitalName := capitalizeFirstLetter(seriesName)
	return fmt.Sprintf("func() float64 { barIdx := ctx.BarIndex-%s; if barIdx >= 0 && barIdx < len(ctx.Data) { return ctx.Data[barIdx].%s }; return math.NaN() }()", indexCode, capitalName)
}
