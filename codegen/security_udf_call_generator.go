package codegen

import (
	"fmt"
	"strings"

	"github.com/quant5-lab/runner/ast"
)

// SecurityUDFCallGenerator generates code for user-defined function calls executed
// within a security() context.  Bar data is sourced from secCtx (the foreign symbol's
// context) rather than the main strategy ctx, so the arrow context is bound to secCtx.
//
// Used by both the tuple security handler (N-return UDF) and the single-return
// security expression handler, ensuring a single canonical code-generation path.
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

// EmitSingleCall evaluates a UDF on foreign-symbol bars by binding the arrow context
// to secCtx instead of the main strategy ctx.
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

// EmitTupleCall evaluates a multi-return UDF on foreign-symbol bars, binding
// the arrow context to secCtx for each return element.
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

func (c *SecurityUDFCallGenerator) buildCallExpression(funcName, arrowCtxVar string, call *ast.CallExpression) (string, error) {
	return c.gen.generateUserDefinedFunctionCallWithContext(call, arrowCtxVar)
}
