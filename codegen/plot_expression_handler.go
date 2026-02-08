package codegen

import (
	"fmt"
	"github.com/quant5-lab/runner/ast"
	"strings"
)

type PlotExpressionHandler struct {
	taRegistry  *InlineTAIIFERegistry
	mathHandler *MathHandler
	generator   *generator
}

func NewPlotExpressionHandler(g *generator) *PlotExpressionHandler {
	return &PlotExpressionHandler{
		taRegistry:  NewInlineTAIIFERegistry(),
		mathHandler: NewMathHandler(),
		generator:   g,
	}
}

func (h *PlotExpressionHandler) Generate(expr ast.Expression) (string, error) {
	switch e := expr.(type) {
	case *ast.ConditionalExpression:
		return h.handleConditional(e)
	case *ast.Identifier:
		return e.Name + "Series.Get(0)", nil
	case *ast.MemberExpression:
		return h.generator.extractSeriesExpression(e), nil
	case *ast.Literal:
		return h.generator.generateNumericExpression(e)
	case *ast.BinaryExpression, *ast.LogicalExpression:
		return h.generator.generateConditionExpression(expr)
	case *ast.CallExpression:
		return h.handleCallExpression(e)
	case *ast.ObjectExpression:
		return h.handleObjectExpression(e)
	default:
		return "", fmt.Errorf("unsupported plot expression type: %T", expr)
	}
}

func (h *PlotExpressionHandler) handleObjectExpression(obj *ast.ObjectExpression) (string, error) {
	for _, prop := range obj.Properties {
		if keyId, ok := prop.Key.(*ast.Identifier); ok {
			if keyId.Name == "type" {
				if memExpr, ok := prop.Value.(*ast.MemberExpression); ok {
					if objId, ok := memExpr.Object.(*ast.Identifier); ok {
						if propId, ok := memExpr.Property.(*ast.Identifier); ok {
							if objId.Name == "input" && propId.Name == "session" {
								return "", nil
							}
						}
					}
				}
			}
		}
	}
	return "", fmt.Errorf("unsupported ObjectExpression in plot context")
}

func (h *PlotExpressionHandler) handleConditional(expr *ast.ConditionalExpression) (string, error) {
	condCode, err := h.generator.generateConditionExpression(expr.Test)
	if err != nil {
		return "", err
	}

	if _, ok := expr.Test.(*ast.Identifier); ok {
		condCode = condCode + " != 0"
	} else if _, ok := expr.Test.(*ast.MemberExpression); ok {
		condCode = condCode + " != 0"
	}

	consequentCode, err := h.generator.generateNumericExpression(expr.Consequent)
	if err != nil {
		return "", err
	}
	alternateCode, err := h.generator.generateNumericExpression(expr.Alternate)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("func() float64 { if %s { return %s } else { return %s } }()",
		condCode, consequentCode, alternateCode), nil
}

func (h *PlotExpressionHandler) handleCallExpression(call *ast.CallExpression) (string, error) {
	/* Check if call was hoisted by InlineExpressionScanner */
	if hoistedVarName := h.generator.tempVarMgr.GetVarNameForCall(call); hoistedVarName != "" {
		return fmt.Sprintf("%sSeries.Get(0)", hoistedVarName), nil
	}

	funcName := h.generator.extractFunctionName(call.Callee)

	if funcName == "ta.atr" || funcName == "atr" {
		return h.HandleATRFunction(call, funcName)
	}

	if h.taRegistry.IsSupported(funcName) {
		return h.HandleTAFunction(call, funcName)
	}

	if h.mathHandler.CanHandle(funcName) {
		return h.mathHandler.GenerateMathCall(funcName, call.Arguments, h.generator)
	}

	/* Check ValueHandler for nz, fixnan, etc. */
	if h.generator.valueHandler.CanHandle(funcName) {
		return h.generator.valueHandler.GenerateInlineCall(funcName, call.Arguments, h.generator)
	}

	if varType, exists := h.generator.variables[funcName]; exists && varType == "function" {
		return h.generator.callRouter.RouteCall(h.generator, call)
	}

	return "", fmt.Errorf("unsupported inline function in plot: %s", funcName)
}

func (h *PlotExpressionHandler) HandleTAFunction(call *ast.CallExpression, funcName string) (string, error) {
	if len(call.Arguments) < 2 {
		return "", fmt.Errorf("%s requires at least 2 arguments (source, period)", funcName)
	}

	if !strings.HasPrefix(funcName, "ta.") {
		funcName = "ta." + funcName
	}

	sourceExpr, periodResult := extractTAArgumentsWithDynamic(h.generator, call, funcName)
	if periodResult.IsFailed() {
		return "", fmt.Errorf("%s: %s", funcName, periodResult.FailureReason)
	}

	if periodResult.IsRuntimeDynamic() {
		return "", fmt.Errorf("inline plot() with runtime dynamic period not supported for %s", funcName)
	}

	period := periodResult.StaticValue
	sourceExprStr := h.generator.extractSeriesExpression(sourceExpr)
	classifier := NewSeriesSourceClassifier()
	sourceInfo := classifier.Classify(sourceExprStr)
	accessor := CreateAccessGenerator(sourceInfo)

	hasher := &ExpressionHasher{}
	sourceHash := hasher.Hash(sourceExpr)

	code, ok := h.taRegistry.Generate(funcName, accessor, NewConstantPeriod(period), sourceHash)
	if !ok {
		return "", fmt.Errorf("inline plot() not implemented for %s", funcName)
	}

	return code, nil
}

func (h *PlotExpressionHandler) HandleATRFunction(call *ast.CallExpression, funcName string) (string, error) {
	periodResult := extractSinglePeriodWithDynamic(h.generator, call, "ta.atr")
	if periodResult.IsFailed() {
		return "", fmt.Errorf("ta.atr: %s", periodResult.FailureReason)
	}

	if periodResult.IsRuntimeDynamic() {
		return "", fmt.Errorf("inline plot() with runtime dynamic period not supported for ta.atr")
	}

	argHash := h.generator.exprAnalyzer.ComputeArgHash(call)
	callInfo := CallInfo{
		Call:     call,
		FuncName: "ta.atr",
		ArgHash:  argHash,
	}

	tempVarName := h.generator.tempVarMgr.GetOrCreate(callInfo)
	return fmt.Sprintf("%sSeries.Get(0)", tempVarName), nil
}

func (h *PlotExpressionHandler) extractPeriod(arg *ast.Literal) (int, error) {
	switch v := arg.Value.(type) {
	case float64:
		return int(v), nil
	case int:
		return v, nil
	default:
		return 0, fmt.Errorf("period must be numeric")
	}
}
