package codegen

import (
	"fmt"

	"github.com/quant5-lab/runner/ast"
)

/* SupertrendHandler: [supertrend, direction] = hl2 ± factor * ATR(atrPeriod)
 * direction = 1 (uptrend, price above supertrend), -1 (downtrend, price below supertrend).
 * Implements CustomTupleFunctionHandler — stateful ForwardSeriesBuffer computation. */
type SupertrendHandler struct{}

func (h *SupertrendHandler) CanHandle(funcName string) bool {
	return funcName == "ta.supertrend" || funcName == "supertrend"
}

func (h *SupertrendHandler) GenerateTupleCode(g *generator, varNames []string, call *ast.CallExpression) (string, error) {
	if len(varNames) != 2 {
		return "", fmt.Errorf("ta.supertrend produces 2 outputs [supertrend, direction], got %d variable names", len(varNames))
	}

	factor, atrPeriod, err := extractSupertrendArguments(call)
	if err != nil {
		return "", err
	}

	context := supertrendStatefulContext(g)
	builder := NewSupertrendIndicatorBuilder(
		[2]string{varNames[0], varNames[1]},
		atrPeriod,
		factor,
		context,
	)

	return g.indentCode(builder.Build()), nil
}

func (h *SupertrendHandler) InternalSeriesNames(firstOutputVar string, _ *ast.CallExpression) []string {
	return []string{
		fmt.Sprintf("_%s_atr", firstOutputVar),
		fmt.Sprintf("_%s_upper", firstOutputVar),
		fmt.Sprintf("_%s_lower", firstOutputVar),
		fmt.Sprintf("_%s_dir", firstOutputVar),
	}
}

func extractSupertrendArguments(call *ast.CallExpression) (factor string, atrPeriod int, err error) {
	if len(call.Arguments) != 2 {
		err = fmt.Errorf("ta.supertrend requires exactly 2 arguments (factor, atrPeriod), got %d", len(call.Arguments))
		return
	}

	factor = stExtractFactorExpr(call, 0)
	atrPeriod, err = stExtractAtrPeriod(call, 1)
	return
}

func stExtractFactorExpr(call *ast.CallExpression, argIdx int) string {
	if argIdx >= len(call.Arguments) {
		return "3.0"
	}
	switch m := call.Arguments[argIdx].(type) {
	case *ast.Literal:
		if v, ok := m.Value.(float64); ok {
			return fmt.Sprintf("%g", v)
		}
	case *ast.Identifier:
		return m.Name
	}
	return "3.0"
}

func stExtractAtrPeriod(call *ast.CallExpression, argIdx int) (int, error) {
	if argIdx >= len(call.Arguments) {
		return 0, fmt.Errorf("ta.supertrend: missing atrPeriod argument")
	}
	lit, ok := call.Arguments[argIdx].(*ast.Literal)
	if !ok {
		return 0, fmt.Errorf("ta.supertrend: atrPeriod must be a compile-time constant")
	}
	return extractPeriod(lit)
}

func supertrendStatefulContext(g *generator) StatefulIndicatorContext {
	if g.inArrowFunctionBody {
		return NewArrowFunctionIndicatorContext()
	}
	return NewTopLevelIndicatorContext()
}
