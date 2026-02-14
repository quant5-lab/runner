package codegen

import (
	"fmt"

	"github.com/quant5-lab/runner/ast"
)

/*
CrossInlineHandler generates crossover/crossunder condition detection.

Responsibility (SRP):
  - Generate IIFE-based crossover condition check
  - Orchestrate argument resolution and TA call generation
  - Enforce stateful indicator extraction requirement

Architecture:
  - Window functions (SMA, WMA): Inline IIFE (stateless, efficient)
  - Stateful indicators (EMA, RMA): Must be extracted to variables first
  - Complex expressions: Direct Series access with curr/prev transformation

Examples:

	✅ ta.crossover(ta.sma(close, 10), ta.sma(close, 20))  // Window functions
	✅ ta.crossover(ema10, 50)                             // Variable reference
	❌ ta.crossover(ta.ema(close, 10), 50)                 // EMA must be variable
*/
type CrossInlineHandler struct {
	isUnder              bool
	statefulDetector     *StatefulIndicatorDetector
	inlineTAIIFERegistry *InlineTAIIFERegistry
}

func NewCrossoverInlineHandler() *CrossInlineHandler {
	return &CrossInlineHandler{
		isUnder:              false,
		statefulDetector:     NewStatefulIndicatorDetector(),
		inlineTAIIFERegistry: NewInlineTAIIFERegistry(),
	}
}

func NewCrossunderInlineHandler() *CrossInlineHandler {
	return &CrossInlineHandler{
		isUnder:              true,
		statefulDetector:     NewStatefulIndicatorDetector(),
		inlineTAIIFERegistry: NewInlineTAIIFERegistry(),
	}
}

func (h *CrossInlineHandler) CanHandle(funcName string) bool {
	if h.isUnder {
		return funcName == "ta.crossunder" || funcName == "crossunder"
	}
	return funcName == "ta.crossover" || funcName == "crossover"
}

func (h *CrossInlineHandler) GenerateInline(expr *ast.CallExpression, g *generator) (string, error) {
	funcName := h.functionName()

	if len(expr.Arguments) < 2 {
		return "", fmt.Errorf("%s requires 2 arguments", funcName)
	}

	series1Curr, series1Prev, err := h.resolveArgument(expr.Arguments[0], g)
	if err != nil {
		return "", fmt.Errorf("%s arg1: %w", funcName, err)
	}

	series2Curr, series2Prev, err := h.resolveArgument(expr.Arguments[1], g)
	if err != nil {
		return "", fmt.Errorf("%s arg2: %w", funcName, err)
	}

	condition := h.buildCondition()

	return fmt.Sprintf("(func() bool { if ctx.BarIndex == 0 { return false }; curr1 := %s; curr2 := %s; prev1 := %s; prev2 := %s; return %s }())",
		series1Curr, series2Curr, series1Prev, series2Prev, condition), nil
}

func (h *CrossInlineHandler) resolveArgument(arg ast.Expression, g *generator) (currAccess, prevAccess string, err error) {
	if err := h.statefulDetector.DetectStateful(arg, g); err != nil {
		return "", "", err
	}

	switch e := arg.(type) {
	case *ast.Identifier:
		return h.resolveIdentifier(e, g)
	case *ast.MemberExpression:
		return h.resolveMemberExpression(e, g)
	case *ast.Literal:
		return h.resolveLiteral(e, g)
	case *ast.BinaryExpression, *ast.UnaryExpression, *ast.ConditionalExpression:
		return h.resolveExpression(e, g)
	case *ast.CallExpression:
		return h.resolveTACall(e, g)
	default:
		return "", "", fmt.Errorf("unsupported argument type: %T", arg)
	}
}

func (h *CrossInlineHandler) resolveIdentifier(ident *ast.Identifier, g *generator) (string, string, error) {
	if code, resolved := g.builtinHandler.TryResolveIdentifier(ident, BarLoopScope); resolved {
		prevCode := g.convertSeriesAccessToPrev(code)
		return code, prevCode, nil
	}
	curr := fmt.Sprintf("%sSeries.GetCurrent()", ident.Name)
	prev := fmt.Sprintf("%sSeries.Get(1)", ident.Name)
	return curr, prev, nil
}

func (h *CrossInlineHandler) resolveMemberExpression(expr *ast.MemberExpression, g *generator) (string, string, error) {
	currCode := g.extractSeriesExpression(expr)
	prevCode := g.convertSeriesAccessToPrev(currCode)
	return currCode, prevCode, nil
}

func (h *CrossInlineHandler) resolveLiteral(lit *ast.Literal, g *generator) (string, string, error) {
	literalStr, err := h.formatLiteral(lit.Value, g)
	if err != nil {
		return "", "", err
	}
	return literalStr, literalStr, nil
}

func (h *CrossInlineHandler) resolveExpression(expr ast.Expression, g *generator) (string, string, error) {
	currCode := g.extractSeriesExpression(expr)
	prevCode := g.convertSeriesAccessToPrev(currCode)
	return currCode, prevCode, nil
}

func (h *CrossInlineHandler) resolveTACall(call *ast.CallExpression, g *generator) (currAccess, prevAccess string, err error) {
	funcName := g.extractFunctionName(call.Callee)

	if !h.inlineTAIIFERegistry.IsSupported(funcName) {
		return "", "", fmt.Errorf("unsupported inline TA function: %s", funcName)
	}

	if len(call.Arguments) < 2 {
		return "", "", fmt.Errorf("%s requires at least 2 arguments", funcName)
	}

	sourceInfo, period, err := h.extractTAArguments(call, funcName, g)
	if err != nil {
		return "", "", err
	}

	if !containsPrefix(funcName, "ta.") {
		funcName = "ta." + funcName
	}

	hasher := &ExpressionHasher{}
	sourceHash := hasher.Hash(call.Arguments[0])

	accessor := CreateAccessGenerator(sourceInfo)
	currIIFE, ok := h.inlineTAIIFERegistry.Generate(funcName, accessor, NewConstantPeriod(period), sourceHash)
	if !ok {
		return "", "", fmt.Errorf("failed to generate inline %s", funcName)
	}

	prevAccessor := CreatePreviousBarAccessGenerator(sourceInfo)
	prevIIFE, ok := h.inlineTAIIFERegistry.Generate(funcName, prevAccessor, NewConstantPeriod(period), sourceHash+"_prev")
	if !ok {
		return "", "", fmt.Errorf("failed to generate inline %s for previous bar", funcName)
	}

	return currIIFE, prevIIFE, nil
}

func (h *CrossInlineHandler) extractTAArguments(call *ast.CallExpression, funcName string, g *generator) (SourceInfo, int, error) {
	sourceExpr := g.extractSeriesExpression(call.Arguments[0])
	classifier := NewSeriesSourceClassifier()
	sourceInfo := classifier.Classify(sourceExpr)

	periodArg, ok := call.Arguments[1].(*ast.Literal)
	if !ok {
		return SourceInfo{}, 0, fmt.Errorf("%s period must be literal", funcName)
	}

	period, err := extractIntFromLiteral(periodArg)
	if err != nil {
		return SourceInfo{}, 0, fmt.Errorf("%s period: %w", funcName, err)
	}

	return sourceInfo, period, nil
}

func (h *CrossInlineHandler) formatLiteral(value interface{}, g *generator) (string, error) {
	switch v := value.(type) {
	case float64:
		return g.literalFormatter.FormatFloat(v), nil
	case int:
		return fmt.Sprintf("%d", v), nil
	case bool:
		return g.literalFormatter.FormatBool(v), nil
	case string:
		return g.literalFormatter.FormatString(v), nil
	default:
		return g.literalFormatter.FormatGeneric(v)
	}
}

func (h *CrossInlineHandler) buildCondition() string {
	if h.isUnder {
		return "curr1 < curr2 && prev1 >= prev2"
	}
	return "curr1 > curr2 && prev1 <= prev2"
}

func (h *CrossInlineHandler) functionName() string {
	if h.isUnder {
		return "ta.crossunder"
	}
	return "ta.crossover"
}

func extractIntFromLiteral(lit *ast.Literal) (int, error) {
	switch v := lit.Value.(type) {
	case float64:
		return int(v), nil
	case int:
		return v, nil
	default:
		return 0, fmt.Errorf("must be numeric, got %T", v)
	}
}

func containsPrefix(s, prefix string) bool {
	return len(s) >= len(prefix) && s[:len(prefix)] == prefix
}
