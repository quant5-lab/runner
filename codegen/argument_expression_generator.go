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
	}
}

func (g *ArgumentExpressionGenerator) Generate(expr ast.Expression) (string, error) {
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
	default:
		return "", fmt.Errorf("unsupported argument expression type: %T", expr)
	}
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
	leftGen := NewArgumentExpressionGenerator(g.generator, g.functionName, g.parameterIndex)
	left, err := leftGen.Generate(bin.Left)
	if err != nil {
		return "", err
	}

	rightGen := NewArgumentExpressionGenerator(g.generator, g.functionName, g.parameterIndex)
	right, err := rightGen.Generate(bin.Right)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("(%s %s %s)", left, bin.Operator, right), nil
}
