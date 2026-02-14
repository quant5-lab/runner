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
	coercer            *NumericExpressionCoercer
}

func NewArrowExpressionGeneratorImpl(gen *generator, resolver *ArrowSeriesAccessResolver) *ArrowExpressionGeneratorImpl {
	identifierResolver := NewArrowIdentifierResolver(resolver)

	exprGen := &ArrowExpressionGeneratorImpl{
		gen:                gen,
		accessResolver:     resolver,
		identifierResolver: identifierResolver,
		coercer:            NewNumericExpressionCoercer(gen.boolConverter),
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

	case *ast.ForInStatement:
		cfGenerator := NewControlFlowExpressionGenerator(e.gen)
		return cfGenerator.GenerateForInExpressionAsIIFE(ex)

	case *ast.WhileStatement:
		cfGenerator := NewControlFlowExpressionGenerator(e.gen)
		return cfGenerator.GenerateWhileExpressionAsIIFE(ex)

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

	securityGen := NewArrowSecurityCallGenerator(e.gen)
	if securityGen.CanHandle(call) {
		return securityGen.Generate(call)
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
		e.gen.arrowAccessResolver = e.accessResolver
		defer func() { e.gen.arrowAccessResolver = nil }()

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
	return sharedTASignatures.Contains(funcName)
}

func (e *ArrowExpressionGeneratorImpl) generateIdentifier(id *ast.Identifier) (string, error) {
	/* Loop counters need float64() cast for arithmetic compatibility */
	if e.gen.loopContextStack != nil && e.gen.loopContextStack.IsLoopCounter(id.Name) {
		return fmt.Sprintf("float64(%s)", id.Name), nil
	}

	if access, resolved := e.accessResolver.ResolveAccess(id.Name); resolved {
		return access, nil
	}

	if code, resolved := e.gen.builtinHandler.TryResolveIdentifier(id, ArrowScope); resolved {
		return code, nil
	}

	return id.Name, nil
}

func (e *ArrowExpressionGeneratorImpl) generateLiteral(lit *ast.Literal) (string, error) {
	return fmt.Sprintf("%v", lit.Value), nil
}

func (e *ArrowExpressionGeneratorImpl) generateNumericExpression(expr ast.Expression) (string, error) {
	code, err := e.generateExpression(expr)
	if err != nil {
		return "", err
	}
	return e.coercer.CoerceToFloat64(expr, code), nil
}

/*
	ensureBooleanOperand converts float64 arrow values to bool via != 0 (PineScript truthiness).

All arrow locals and parameters are float64, so all identifiers need conversion.
*/
func (e *ArrowExpressionGeneratorImpl) ensureBooleanOperand(expr ast.Expression, code string) string {
	if e.gen.boolConverter.IsAlreadyBoolean(expr) {
		return code
	}
	return fmt.Sprintf("(%s != 0)", code)
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
		operand = e.ensureBooleanOperand(unaryExpr.Argument, operand)
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

	left = e.ensureBooleanOperand(logExpr.Left, left)
	right = e.ensureBooleanOperand(logExpr.Right, right)

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

	test = e.ensureBooleanOperand(condExpr.Test, test)

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

	if code, resolved := e.gen.builtinHandler.TryResolveMemberExpression(mem, ArrowScope); resolved {
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

/* Routes builtin vs parameter vs local series to their respective access patterns */
func (e *ArrowExpressionGeneratorImpl) resolveArrowSubscript(seriesName string, indexExpr ast.Expression) (string, error) {
	isSeriesParam := e.accessResolver.IsParameter(seriesName)
	isBuiltin := e.gen.builtinHandler.IsBuiltinSeriesIdentifier(seriesName)

	indexCode, err := e.generateArrowIndexExpression(indexExpr)
	if err != nil {
		return "", fmt.Errorf("failed to generate index expression: %w", err)
	}

	if isBuiltin {
		return e.generateBuiltinSubscript(seriesName, indexCode), nil
	}

	/* Constant builtins (last_bar_index) are time-invariant — subscript returns scalar */
	if e.gen.builtinHandler.IsConstantBuiltin(seriesName) {
		return seriesName, nil
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
	handler := e.gen.builtinHandler

	if info, ok := handler.CalendarInfo(seriesName); ok {
		return CalendarFieldArrowIIFE(info.ArrowExpression, indexCode)
	}

	if seriesName == "time" {
		return TimeArrowIIFE(indexCode)
	}

	if seriesName == "bar_index" {
		return BarIndexArrowIIFE(indexCode)
	}

	if handler.IsDerivedPrice(seriesName) {
		return e.generateDerivedPriceSubscript(seriesName, indexCode)
	}

	if seriesName == "tr" {
		return TrueRangeArrowIIFE(indexCode)
	}

	return OHLCVFieldArrowIIFE(capitalizeFirstLetter(seriesName), indexCode)
}

func (e *ArrowExpressionGeneratorImpl) generateDerivedPriceSubscript(priceName, indexCode string) string {
	formula := e.gen.builtinHandler.GenerateDerivedPriceFormula(priceName,
		"ctx.Data[barIdx].High", "ctx.Data[barIdx].Low", "ctx.Data[barIdx].Close", "ctx.Data[barIdx].Open")
	if formula == "" {
		return "math.NaN()"
	}
	return boundsCheckedArrowIIFE(indexCode, "return "+formula)
}
