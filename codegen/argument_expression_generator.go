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
	inSecurityContext bool
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
		inSecurityContext: gen.inSecurityContext,
		coercer:           NewNumericExpressionCoercer(gen.boolConverter),
	}
}

/* Generate produces a float64-typed Go expression for use as a function argument */
func (g *ArgumentExpressionGenerator) Generate(expr ast.Expression) (string, error) {
	code, err := g.generate(expr)
	if err != nil {
		return "", err
	}
	return g.ensureFloat64(expr, code), nil
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
	if constVal, isConstant := g.generator.constants[id.Name]; isConstant {
		if constVal == "input.source" {
			return fmt.Sprintf("%sSeries.GetCurrent()", id.Name), nil
		}
		return id.Name, nil
	}

	if code, resolved := g.builtinHandler.TryResolveIdentifier(id, g.inSecurityContext); resolved {
		paramType, hasSignature := g.signatureRegistry.GetParameterType(g.functionName, g.parameterIndex)

		if hasSignature && paramType == ParamTypeSeries {
			return g.resolveBuiltinToSeries(id.Name, code)
		}
		return g.resolveBuiltinToValue(id.Name, code)
	}

	if g.signatureRegistry != nil {
		paramType, hasSignature := g.signatureRegistry.GetParameterType(g.functionName, g.parameterIndex)
		if hasSignature && paramType == ParamTypeSeries {
			return fmt.Sprintf("%sSeries", id.Name), nil
		}
	}

	return g.generator.resolveUserIdentifierAccess(id.Name), nil
}

func (g *ArgumentExpressionGenerator) resolveBuiltinToSeries(name, fallback string) (string, error) {
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

	op := logical.Operator
	switch op {
	case "and":
		op = "&&"
	case "or":
		op = "||"
	}

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
