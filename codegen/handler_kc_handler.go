package codegen

import (
	"fmt"

	"github.com/quant5-lab/runner/ast"
)

/* KcHandler: [upper, basis, lower] = EMA(source,length) ± mult * range(length)
 * range = ATR(length) when useTrueRange=true, or SMA(high-low, length) otherwise.
 * Implements CustomTupleFunctionHandler — stateful ForwardSeriesBuffer computation. */
type KcHandler struct{}

func (h *KcHandler) CanHandle(funcName string) bool {
	return funcName == "ta.kc" || funcName == "kc"
}

func (h *KcHandler) GenerateTupleCode(g *generator, varNames []string, call *ast.CallExpression) (string, error) {
	if len(varNames) != 3 {
		return "", fmt.Errorf("ta.kc produces 3 outputs [upper, basis, lower], got %d variable names", len(varNames))
	}

	sourceExpr, period, multExpr, useTrueRange, err := extractKCArguments(call)
	if err != nil {
		return "", err
	}

	classifier := &SeriesSourceClassifier{}
	sourceInfo := classifier.ClassifyAST(sourceExpr)
	accessor := CreateAccessGenerator(sourceInfo)

	context := kcStatefulContext(g)
	builder := NewKCIndicatorBuilder(
		[3]string{varNames[0], varNames[1], varNames[2]},
		NewConstantPeriod(period),
		accessor,
		multExpr,
		useTrueRange,
		context,
	)

	return g.indentCode(builder.Build()), nil
}

func (h *KcHandler) InternalSeriesNames(firstOutputVar string, _ *ast.CallExpression) []string {
	return []string{
		fmt.Sprintf("_%s_ema", firstOutputVar),
		fmt.Sprintf("_%s_atr", firstOutputVar),
	}
}

func extractKCArguments(call *ast.CallExpression) (sourceExpr ast.Expression, period int, multExpr string, useTrueRange bool, err error) {
	switch len(call.Arguments) {
	case 2:
		sourceExpr = &ast.Identifier{Name: "close"}
		period, err = kcExtractPeriod(call, 0)
		multExpr = kcExtractMultExpr(call, 1)
		useTrueRange = true
	case 3:
		sourceExpr = call.Arguments[0]
		period, err = kcExtractPeriod(call, 1)
		multExpr = kcExtractMultExpr(call, 2)
		useTrueRange = true
	case 4:
		sourceExpr = call.Arguments[0]
		period, err = kcExtractPeriod(call, 1)
		multExpr = kcExtractMultExpr(call, 2)
		useTrueRange = extractUseTrueRangeArg(call, 3)
	default:
		err = fmt.Errorf("ta.kc requires 2-4 arguments, got %d", len(call.Arguments))
	}
	return
}

func kcExtractPeriod(call *ast.CallExpression, argIdx int) (int, error) {
	if argIdx >= len(call.Arguments) {
		return 0, fmt.Errorf("ta.kc: missing period argument")
	}
	lit, ok := call.Arguments[argIdx].(*ast.Literal)
	if !ok {
		return 0, fmt.Errorf("ta.kc: period must be a compile-time constant")
	}
	return extractPeriod(lit)
}

func kcExtractMultExpr(call *ast.CallExpression, argIdx int) string {
	if argIdx >= len(call.Arguments) {
		return "2.0"
	}
	switch m := call.Arguments[argIdx].(type) {
	case *ast.Literal:
		if v, ok := m.Value.(float64); ok {
			return fmt.Sprintf("%g", v)
		}
	case *ast.Identifier:
		return m.Name
	}
	return "2.0"
}

func kcStatefulContext(g *generator) StatefulIndicatorContext {
	if g.inArrowFunctionBody {
		return NewArrowFunctionIndicatorContext()
	}
	return NewTopLevelIndicatorContext()
}
