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
	default:
		return "", fmt.Errorf("unsupported plot expression type: %T", expr)
	}
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
	funcName := h.generator.extractFunctionName(call.Callee)

	if h.taRegistry.IsSupported(funcName) {
		return h.HandleTAFunction(call, funcName)
	}

	if h.isMathFunction(funcName) {
		return h.mathHandler.GenerateMathCall(funcName, call.Arguments, h.generator)
	}

	return "", fmt.Errorf("unsupported inline function in plot: %s", funcName)
}

func (h *PlotExpressionHandler) HandleTAFunction(call *ast.CallExpression, funcName string) (string, error) {
	if len(call.Arguments) < 2 {
		return "", fmt.Errorf("%s requires at least 2 arguments (source, period)", funcName)
	}

	sourceExpr := h.generator.extractSeriesExpression(call.Arguments[0])
	classifier := NewSeriesSourceClassifier()
	sourceInfo := classifier.Classify(sourceExpr)
	accessor := CreateAccessGenerator(sourceInfo)

	periodArg, ok := call.Arguments[1].(*ast.Literal)
	if !ok {
		return "", fmt.Errorf("%s period must be literal", funcName)
	}

	period, err := h.extractPeriod(periodArg)
	if err != nil {
		return "", fmt.Errorf("%s: %w", funcName, err)
	}

	if !strings.HasPrefix(funcName, "ta.") {
		funcName = "ta." + funcName
	}

	code, ok := h.taRegistry.Generate(funcName, accessor, period)
	if !ok {
		return "", fmt.Errorf("inline plot() not implemented for %s", funcName)
	}

	return code, nil
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

func (h *PlotExpressionHandler) isMathFunction(funcName string) bool {
	return funcName == "math.abs" || funcName == "math.max" || funcName == "math.min"
}
