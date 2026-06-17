package codegen

import (
	"fmt"
	"strings"

	"github.com/quant5-lab/runner/ast"
)

// SecurityUDFCallGenerator bar data is sourced from secCtx (the foreign symbol's context)
// so the arrow context is bound to secCtx, not the main strategy ctx.
type SecurityUDFCallGenerator struct {
	gen *generator
}

func NewSecurityUDFCallGenerator(gen *generator) *SecurityUDFCallGenerator {
	return &SecurityUDFCallGenerator{gen: gen}
}

// IsUDFCall reports whether expr is a call to a known user-defined function.
func (c *SecurityUDFCallGenerator) IsUDFCall(expr ast.Expression) bool {
	callExpr, ok := expr.(*ast.CallExpression)
	if !ok {
		return false
	}
	return NewUserDefinedFunctionDetector(c.gen.variables).
		IsUserDefinedFunction(extractCallFunctionName(callExpr))
}

func (c *SecurityUDFCallGenerator) EmitSingleCall(varName string, call *ast.CallExpression) (string, error) {
	funcName := extractCallFunctionName(call)
	arrowCtxVar := c.gen.arrowContextLifecycle.AllocateContextVariable(funcName + "_sec")

	code := c.gen.ind() + fmt.Sprintf("%s := context.NewArrowContext(secCtx)\n", arrowCtxVar)

	callCode, err := c.buildCallExpression(funcName, arrowCtxVar, call)
	if err != nil {
		return "", err
	}

	code += c.gen.ind() + fmt.Sprintf("%sSeries.Set(%s)\n", varName, callCode)
	return code, nil
}

func (c *SecurityUDFCallGenerator) EmitTupleCall(varNames []string, call *ast.CallExpression) (string, error) {
	funcName := extractCallFunctionName(call)
	arrowCtxVar := c.gen.arrowContextLifecycle.AllocateContextVariable(funcName + "_sec")

	code := c.gen.ind() + fmt.Sprintf("%s := context.NewArrowContext(secCtx)\n", arrowCtxVar)

	callCode, err := c.buildCallExpression(funcName, arrowCtxVar, call)
	if err != nil {
		return "", err
	}

	code += c.gen.ind() + fmt.Sprintf("%s := %s\n", strings.Join(varNames, ", "), callCode)

	for _, varName := range varNames {
		code += c.gen.ind() + fmt.Sprintf("%sSeries.Set(%s)\n", varName, varName)
	}
	return code, nil
}

// EmitBarLoopUDFEval: keyed by secKey+":"+funcName so distinct (symbol, func) pairs each
// carry independent ArrowContext history.  varName is the base series name (without "Series").
func (c *SecurityUDFCallGenerator) EmitBarLoopUDFEval(varName string, call *ast.CallExpression) (string, error) {
	funcName := extractCallFunctionName(call)
	callExpr, err := c.gen.generateUserDefinedFunctionCallWithContext(call, "secArrowCtx")
	if err != nil {
		return "", fmt.Errorf("EmitBarLoopUDFEval %s: %w", funcName, err)
	}
	c.gen.hasSecurityUDFEvals = true

	ind := c.gen.ind
	code := ind() + "{\n"
	c.gen.indent++
	code += ind() + fmt.Sprintf("udfKey := secKey + %q\n", ":"+funcName)
	code += ind() + "if secUDFBarEvaluators == nil {\n"
	c.gen.indent++
	code += ind() + "secUDFBarEvaluators = make(map[string]security.BarEvaluator)\n"
	c.gen.indent--
	code += ind() + "}\n"
	code += ind() + "if secUDFBarEvaluators[udfKey] == nil {\n"
	c.gen.indent++
	code += ind() + "secUDFBarEvaluators[udfKey] = security.NewUDFBarEvaluator(len(secCtx.Data), secCtx, func(secArrowCtx *context.ArrowContext) float64 {\n"
	c.gen.indent++
	code += ind() + "return " + callExpr + "\n"
	c.gen.indent--
	code += ind() + "})\n"
	c.gen.indent--
	code += ind() + "}\n"
	code += ind() + "secUDFVal, _ := secUDFBarEvaluators[udfKey].EvaluateAtBar(nil, nil, secBarIdx)\n"
	code += ind() + fmt.Sprintf("%sSeries.Set(secUDFVal)\n", varName)
	c.gen.indent--
	code += ind() + "}\n"
	return code, nil
}

// EmitArrowContextUDFEval stores the evaluator in arrowCtx.GetOrCreateSecurityEvaluators()
// under "udf:"+secKey+":"+funcName so it persists across main bars through the ArrowContext
// lifetime.  Uses \t\t indentation shared by all IIFE evaluation paths.
func (c *SecurityUDFCallGenerator) EmitArrowContextUDFEval(call *ast.CallExpression) (string, error) {
	funcName := extractCallFunctionName(call)
	callExpr, err := c.gen.generateUserDefinedFunctionCallWithContext(call, "secArrowCtx")
	if err != nil {
		return "", fmt.Errorf("EmitArrowContextUDFEval %s: %w", funcName, err)
	}

	var b strings.Builder
	b.WriteString(fmt.Sprintf("\t\tudfKey := \"udf:\" + secKey + %q\n", ":"+funcName))
	b.WriteString("\t\tarrowUDFMap := arrowCtx.GetOrCreateSecurityEvaluators()\n")
	b.WriteString("\t\tif arrowUDFMap[udfKey] == nil {\n")
	b.WriteString("\t\t\tarrowUDFMap[udfKey] = security.NewUDFBarEvaluator(len(secCtx.Data), secCtx, func(secArrowCtx *context.ArrowContext) float64 {\n")
	b.WriteString("\t\t\t\treturn " + callExpr + "\n")
	b.WriteString("\t\t\t})\n")
	b.WriteString("\t\t}\n")
	b.WriteString("\t\tsecUDFVal, _ := arrowUDFMap[udfKey].(security.BarEvaluator).EvaluateAtBar(nil, nil, secBarIdx)\n")
	b.WriteString("\t\treturn secUDFVal\n")
	return b.String(), nil
}

func (c *SecurityUDFCallGenerator) buildCallExpression(funcName, arrowCtxVar string, call *ast.CallExpression) (string, error) {
	return c.gen.generateUserDefinedFunctionCallWithContext(call, arrowCtxVar)
}
