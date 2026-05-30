package codegen

import (
	"fmt"

	"github.com/quant5-lab/runner/ast"
)

/* ArgumentExpressionGenerator converts AST expressions to Go function call arguments */
type ArgumentExpressionGenerator struct {
	generator         *generator
	functionName      string
	parameterIndex    int
	signatureRegistry *FunctionSignatureRegistry
	builtinHandler    *BuiltinIdentifierHandler
	scope             AccessScope
	coercer           *NumericExpressionCoercer
}

func NewArgumentExpressionGenerator(
	gen *generator,
	funcName string,
	paramIdx int,
) *ArgumentExpressionGenerator {
	return &ArgumentExpressionGenerator{
		generator:         gen,
		functionName:      funcName,
		parameterIndex:    paramIdx,
		signatureRegistry: gen.funcSigRegistry,
		builtinHandler:    gen.builtinHandler,
		scope:             gen.accessScope(),
		coercer:           NewNumericExpressionCoercer(gen.boolConverter),
	}
}

/* Generate produces a correctly-typed Go expression for use as a function argument.
 * String-typed parameters bypass float64 coercion. */
func (g *ArgumentExpressionGenerator) Generate(expr ast.Expression) (string, error) {
	code, err := g.generate(expr)
	if err != nil {
		return "", err
	}
	if g.isStringParam() {
		return code, nil
	}
	return g.ensureFloat64(expr, code), nil
}

func (g *ArgumentExpressionGenerator) isStringParam() bool {
	if g.signatureRegistry == nil {
		return false
	}
	paramType, ok := g.signatureRegistry.GetParameterType(g.functionName, g.parameterIndex)
	return ok && paramType == ParamTypeString
}

/* generate produces a raw Go expression preserving its native Go type (bool stays bool) */
func (g *ArgumentExpressionGenerator) generate(expr ast.Expression) (string, error) {
	switch e := expr.(type) {
	case *ast.Identifier:
		return g.generateIdentifier(e)
	case *ast.Literal:
		return g.generateLiteral(e)
	case *ast.CallExpression:
		return g.generator.generateCallExpression(e)
	case *ast.BinaryExpression:
		return g.generateBinaryExpression(e)
	case *ast.MemberExpression:
		return g.generator.generateMemberExpression(e)
	case *ast.UnaryExpression:
		return g.generateUnaryExpression(e)
	case *ast.LogicalExpression:
		return g.generateLogicalExpression(e)
	case *ast.ConditionalExpression:
		return g.generateConditionalExpression(e)
	case *ast.IfStatement:
		cfGenerator := NewControlFlowExpressionGenerator(g.generator)
		return cfGenerator.GenerateIfExpressionAsIIFE(e)
	case *ast.ForStatement:
		cfGenerator := NewControlFlowExpressionGenerator(g.generator)
		return cfGenerator.GenerateForExpressionAsIIFE(e)
	case *ast.ForInStatement:
		cfGenerator := NewControlFlowExpressionGenerator(g.generator)
		return cfGenerator.GenerateForInExpressionAsIIFE(e)
	case *ast.WhileStatement:
		cfGenerator := NewControlFlowExpressionGenerator(g.generator)
		return cfGenerator.GenerateWhileExpressionAsIIFE(e)
	default:
		return "", fmt.Errorf("unsupported argument expression type: %T", expr)
	}
}

func (g *ArgumentExpressionGenerator) ensureFloat64(expr ast.Expression, code string) string {
	return g.coercer.CoerceToFloat64(expr, code)
}

func (g *ArgumentExpressionGenerator) generateIdentifier(id *ast.Identifier) (string, error) {
	expectsSeries := false
	if g.signatureRegistry != nil {
		paramType, hasSignature := g.signatureRegistry.GetParameterType(g.functionName, g.parameterIndex)
		expectsSeries = hasSignature && paramType == ParamTypeSeries
	}

	if constVal, isConstant := g.generator.constants[id.Name]; isConstant {
		if constVal == "input.source" {
			if expectsSeries {
				return fmt.Sprintf("%sSeries", id.Name), nil
			}
			return fmt.Sprintf("%sSeries.GetCurrent()", id.Name), nil
		}
		return id.Name, nil
	}

	// Arrow resolver checked before builtins; bypassed when passing to a series-typed parameter
	// so the identifier routes to *Series by name convention instead of scalar resolution.
	if !expectsSeries && g.generator.arrowAccessResolver != nil {
		if access, resolved := g.generator.arrowAccessResolver.ResolveAccess(id.Name); resolved {
			return access, nil
		}
	}

	if code, resolved := g.builtinHandler.TryResolveIdentifier(id, g.scope); resolved {
		if expectsSeries {
			return g.resolveBuiltinToSeries(id.Name, code)
		}
		return g.resolveBuiltinToValue(id.Name, code)
	}

	if expectsSeries {
		return fmt.Sprintf("%sSeries", id.Name), nil
	}

	return g.generator.resolveUserIdentifierAccess(id.Name), nil
}

func (g *ArgumentExpressionGenerator) resolveBuiltinToSeries(name, fallback string) (string, error) {
	if g.scope == ArrowScope {
		if _, ok := OHLCVFieldName(name); ok {
			return SeriesPointerLookupIIFE(name + "Series"), nil
		}
		return fallback, nil
	}
	switch name {
	case "close":
		return "closeSeries", nil
	case "open":
		return "openSeries", nil
	case "high":
		return "highSeries", nil
	case "low":
		return "lowSeries", nil
	case "volume":
		return "volumeSeries", nil
	default:
		return fallback, nil
	}
}

func (g *ArgumentExpressionGenerator) resolveBuiltinToValue(name, fallback string) (string, error) {
	if g.scope != BarLoopScope {
		return fallback, nil
	}
	switch name {
	case "close":
		return "closeSeries.Get(0)", nil
	case "open":
		return "openSeries.Get(0)", nil
	case "high":
		return "highSeries.Get(0)", nil
	case "low":
		return "lowSeries.Get(0)", nil
	case "volume":
		return "volumeSeries.Get(0)", nil
	default:
		return fallback, nil
	}
}

func (g *ArgumentExpressionGenerator) generateLiteral(lit *ast.Literal) (string, error) {
	switch v := lit.Value.(type) {
	case float64:
		return fmt.Sprintf("%.1f", v), nil
	case int:
		return fmt.Sprintf("%d.0", v), nil
	case string:
		return fmt.Sprintf("%q", v), nil
	default:
		return fmt.Sprintf("%v", v), nil
	}
}

func (g *ArgumentExpressionGenerator) generateBinaryExpression(bin *ast.BinaryExpression) (string, error) {
	left, err := g.generate(bin.Left)
	if err != nil {
		return "", err
	}

	right, err := g.generate(bin.Right)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("(%s %s %s)", left, bin.Operator, right), nil
}

func (g *ArgumentExpressionGenerator) generateUnaryExpression(unary *ast.UnaryExpression) (string, error) {
	operandCode, err := g.generate(unary.Argument)
	if err != nil {
		return "", err
	}

	op := unary.Operator
	if op == "not" {
		op = "!"
		operandCode = g.generator.ensureBooleanOperand(unary.Argument, operandCode)
	}

	return fmt.Sprintf("%s%s", op, operandCode), nil
}

func (g *ArgumentExpressionGenerator) generateLogicalExpression(logical *ast.LogicalExpression) (string, error) {
	leftCode, err := g.generate(logical.Left)
	if err != nil {
		return "", err
	}

	rightCode, err := g.generate(logical.Right)
	if err != nil {
		return "", err
	}

	leftCode = g.generator.ensureBooleanOperand(logical.Left, leftCode)
	rightCode = g.generator.ensureBooleanOperand(logical.Right, rightCode)

	op := NormalizeLogicalOperator(logical.Operator)
	return fmt.Sprintf("(%s %s %s)", leftCode, op, rightCode), nil
}

func (g *ArgumentExpressionGenerator) generateConditionalExpression(cond *ast.ConditionalExpression) (string, error) {
	testCode, err := g.generate(cond.Test)
	if err != nil {
		return "", err
	}
	testCode = g.generator.addBoolConversionIfNeeded(cond.Test, testCode)

	consequentCode, err := g.generate(cond.Consequent)
	if err != nil {
		return "", err
	}
	consequentCode = g.coercer.CoerceToFloat64(cond.Consequent, consequentCode)

	alternateCode, err := g.generate(cond.Alternate)
	if err != nil {
		return "", err
	}
	alternateCode = g.coercer.CoerceToFloat64(cond.Alternate, alternateCode)

	return fmt.Sprintf("func() float64 { if %s { return %s } else { return %s } }()",
		testCode, consequentCode, alternateCode), nil
}
