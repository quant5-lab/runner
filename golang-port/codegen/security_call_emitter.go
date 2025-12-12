package codegen

import (
	"fmt"

	"github.com/quant5-lab/runner/ast"
)

type SecurityCallEmitter struct {
	gen      *generator
	resolver *ConstantResolver
}

func NewSecurityCallEmitter(gen *generator) *SecurityCallEmitter {
	return &SecurityCallEmitter{
		gen:      gen,
		resolver: NewConstantResolver(),
	}
}

func (e *SecurityCallEmitter) EmitSecurityCall(varName string, call *ast.CallExpression) (string, error) {
	if len(call.Arguments) < 3 {
		return "", fmt.Errorf("request.security requires 3 arguments")
	}

	symbolExpr := call.Arguments[0]
	timeframeExpr := call.Arguments[1]
	exprArg := call.Arguments[2]

	lookahead := false
	if len(call.Arguments) >= 4 {
		fourthArg := call.Arguments[3]

		if objExpr, ok := fourthArg.(*ast.ObjectExpression); ok {
			for _, prop := range objExpr.Properties {
				if keyIdent, ok := prop.Key.(*ast.Identifier); ok && keyIdent.Name == "lookahead" {
					if resolved, ok := e.resolver.ResolveToBool(prop.Value); ok {
						lookahead = resolved
					}
					break
				}
			}
		} else {
			if resolved, ok := e.resolver.ResolveToBool(fourthArg); ok {
				lookahead = resolved
			}
		}
	}

	symbolCode, err := e.extractSymbolCode(symbolExpr)
	if err != nil {
		return "", err
	}

	timeframeCode, err := e.extractTimeframeCode(timeframeExpr)
	if err != nil {
		return "", err
	}

	return e.emitStreamingEvaluation(varName, symbolCode, timeframeCode, exprArg, lookahead)
}

func (e *SecurityCallEmitter) extractSymbolCode(expr ast.Expression) (string, error) {
	switch exp := expr.(type) {
	case *ast.Identifier:
		if exp.Name == "tickerid" {
			return "ctx.Symbol", nil
		}
		return fmt.Sprintf("%q", exp.Name), nil
	case *ast.MemberExpression:
		return "ctx.Symbol", nil
	case *ast.Literal:
		if s, ok := exp.Value.(string); ok {
			return fmt.Sprintf("%q", s), nil
		}
		return "", fmt.Errorf("invalid symbol literal type")
	default:
		return "", fmt.Errorf("unsupported symbol expression type: %T", expr)
	}
}

func (e *SecurityCallEmitter) extractTimeframeCode(expr ast.Expression) (string, error) {
	if lit, ok := expr.(*ast.Literal); ok {
		if s, ok := lit.Value.(string); ok {
			return fmt.Sprintf("%q", s), nil
		}
	}
	return "", fmt.Errorf("invalid timeframe expression")
}

func (e *SecurityCallEmitter) emitStreamingEvaluation(varName, symbolCode, timeframeCode string, expr ast.Expression, lookahead bool) (string, error) {
	var code string

	code += e.gen.ind() + fmt.Sprintf("secKey := fmt.Sprintf(\"%%s:%%s\", %s, %s)\n", symbolCode, timeframeCode)
	code += e.gen.ind() + "secCtx, secFound := securityContexts[secKey]\n"
	code += e.gen.ind() + "if !secFound {\n"
	e.gen.indent++
	code += e.gen.ind() + fmt.Sprintf("%sSeries.Set(math.NaN())\n", varName)
	e.gen.indent--
	code += e.gen.ind() + "} else {\n"
	e.gen.indent++

	if lookahead {
		code += e.gen.ind() + "secBarIdx := context.FindBarIndexByTimestampWithLookahead(secCtx, ctx.Data[ctx.BarIndex].Time)\n"
	} else {
		code += e.gen.ind() + "secBarIdx := context.FindBarIndexByTimestamp(secCtx, ctx.Data[ctx.BarIndex].Time)\n"
	}
	code += e.gen.ind() + "if secBarIdx < 0 {\n"
	e.gen.indent++
	code += e.gen.ind() + fmt.Sprintf("%sSeries.Set(math.NaN())\n", varName)
	e.gen.indent--
	code += e.gen.ind() + "} else {\n"
	e.gen.indent++

	exprCode, err := e.emitExpressionEvaluation(varName, expr)
	if err != nil {
		return "", err
	}
	code += exprCode

	e.gen.indent--
	code += e.gen.ind() + "}\n"
	e.gen.indent--
	code += e.gen.ind() + "}\n"

	return code, nil
}

func (e *SecurityCallEmitter) emitExpressionEvaluation(varName string, expr ast.Expression) (string, error) {
	switch exp := expr.(type) {
	case *ast.Identifier:
		return e.emitIdentifierEvaluation(varName, exp)
	case *ast.CallExpression:
		return e.emitTAFunctionEvaluation(varName, exp)
	case *ast.BinaryExpression:
		return e.emitBinaryExpressionEvaluation(varName, exp)
	default:
		return "", fmt.Errorf("unsupported security expression type: %T", expr)
	}
}

func (e *SecurityCallEmitter) emitIdentifierEvaluation(varName string, id *ast.Identifier) (string, error) {
	var code string

	switch id.Name {
	case "close":
		code += e.gen.ind() + fmt.Sprintf("%sSeries.Set(secCtx.Data[secBarIdx].Close)\n", varName)
	case "open":
		code += e.gen.ind() + fmt.Sprintf("%sSeries.Set(secCtx.Data[secBarIdx].Open)\n", varName)
	case "high":
		code += e.gen.ind() + fmt.Sprintf("%sSeries.Set(secCtx.Data[secBarIdx].High)\n", varName)
	case "low":
		code += e.gen.ind() + fmt.Sprintf("%sSeries.Set(secCtx.Data[secBarIdx].Low)\n", varName)
	case "volume":
		code += e.gen.ind() + fmt.Sprintf("%sSeries.Set(secCtx.Data[secBarIdx].Volume)\n", varName)
	default:
		code += e.gen.ind() + fmt.Sprintf("%sSeries.Set(math.NaN())\n", varName)
	}

	return code, nil
}

func (e *SecurityCallEmitter) emitTAFunctionEvaluation(varName string, call *ast.CallExpression) (string, error) {
	var code string

	evaluatorVar := "secBarEvaluator"
	code += e.gen.ind() + fmt.Sprintf("if %s == nil {\n", evaluatorVar)
	e.gen.indent++
	code += e.gen.ind() + fmt.Sprintf("%s = security.NewStreamingBarEvaluator()\n", evaluatorVar)
	e.gen.indent--
	code += e.gen.ind() + "}\n"

	exprJSON, err := e.serializeExpressionToCode(call)
	if err != nil {
		return "", err
	}

	code += e.gen.ind() + fmt.Sprintf("secValue, err := %s.EvaluateAtBar(%s, secCtx, secBarIdx)\n", evaluatorVar, exprJSON)
	code += e.gen.ind() + "if err != nil {\n"
	e.gen.indent++
	code += e.gen.ind() + fmt.Sprintf("%sSeries.Set(math.NaN())\n", varName)
	e.gen.indent--
	code += e.gen.ind() + "} else {\n"
	e.gen.indent++
	code += e.gen.ind() + fmt.Sprintf("%sSeries.Set(secValue)\n", varName)
	e.gen.indent--
	code += e.gen.ind() + "}\n"

	return code, nil
}

func (e *SecurityCallEmitter) emitBinaryExpressionEvaluation(varName string, binExpr *ast.BinaryExpression) (string, error) {
	var code string

	evaluatorVar := "secBarEvaluator"
	code += e.gen.ind() + fmt.Sprintf("if %s == nil {\n", evaluatorVar)
	e.gen.indent++
	code += e.gen.ind() + fmt.Sprintf("%s = security.NewStreamingBarEvaluator()\n", evaluatorVar)
	e.gen.indent--
	code += e.gen.ind() + "}\n"

	exprJSON, err := e.serializeExpressionToCode(binExpr)
	if err != nil {
		return "", err
	}

	code += e.gen.ind() + fmt.Sprintf("secValue, err := %s.EvaluateAtBar(%s, secCtx, secBarIdx)\n", evaluatorVar, exprJSON)
	code += e.gen.ind() + "if err != nil {\n"
	e.gen.indent++
	code += e.gen.ind() + fmt.Sprintf("%sSeries.Set(math.NaN())\n", varName)
	e.gen.indent--
	code += e.gen.ind() + "} else {\n"
	e.gen.indent++
	code += e.gen.ind() + fmt.Sprintf("%sSeries.Set(secValue)\n", varName)
	e.gen.indent--
	code += e.gen.ind() + "}\n"

	return code, nil
}

func (e *SecurityCallEmitter) serializeExpressionToCode(expr ast.Expression) (string, error) {
	switch exp := expr.(type) {
	case *ast.CallExpression:
		return e.serializeCallExpression(exp)
	case *ast.BinaryExpression:
		return e.serializeBinaryExpression(exp)
	case *ast.Identifier:
		return fmt.Sprintf("&ast.Identifier{Name: %q}", exp.Name), nil
	case *ast.Literal:
		if val, ok := exp.Value.(float64); ok {
			return fmt.Sprintf("&ast.Literal{Value: %.1f}", val), nil
		}
		if val, ok := exp.Value.(string); ok {
			return fmt.Sprintf("&ast.Literal{Value: %q}", val), nil
		}
		return "", fmt.Errorf("unsupported literal type: %T", exp.Value)
	default:
		return "", fmt.Errorf("unsupported expression type for serialization: %T", expr)
	}
}

func (e *SecurityCallEmitter) serializeCallExpression(call *ast.CallExpression) (string, error) {
	funcName, err := e.extractFunctionName(call.Callee)
	if err != nil {
		return "", err
	}

	args := ""
	for i, arg := range call.Arguments {
		argCode, err := e.serializeExpressionToCode(arg)
		if err != nil {
			return "", err
		}
		if i > 0 {
			args += ", "
		}
		args += argCode
	}

	return fmt.Sprintf("&ast.CallExpression{Callee: &ast.MemberExpression{Object: &ast.Identifier{Name: %q}, Property: &ast.Identifier{Name: %q}}, Arguments: []ast.Expression{%s}}",
		funcName[:2], funcName[3:], args), nil
}

func (e *SecurityCallEmitter) serializeBinaryExpression(binExpr *ast.BinaryExpression) (string, error) {
	leftCode, err := e.serializeExpressionToCode(binExpr.Left)
	if err != nil {
		return "", err
	}

	rightCode, err := e.serializeExpressionToCode(binExpr.Right)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("&ast.BinaryExpression{Operator: %q, Left: %s, Right: %s}",
		binExpr.Operator, leftCode, rightCode), nil
}

func (e *SecurityCallEmitter) extractFunctionName(callee ast.Expression) (string, error) {
	if mem, ok := callee.(*ast.MemberExpression); ok {
		obj := ""
		if id, ok := mem.Object.(*ast.Identifier); ok {
			obj = id.Name
		}
		prop := ""
		if id, ok := mem.Property.(*ast.Identifier); ok {
			prop = id.Name
		}
		return obj + "." + prop, nil
	}

	if id, ok := callee.(*ast.Identifier); ok {
		return id.Name, nil
	}

	return "", fmt.Errorf("unsupported callee type: %T", callee)
}
