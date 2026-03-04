package codegen

import (
	"fmt"
	"strings"

	"github.com/quant5-lab/runner/ast"
)

/* PivotPointLevelsHandler: ta.pivot_point_levels(type, anchor [, developing]) → array<float>[11]
 * State machine (ForwardSeriesBuffer paradigm): four internal Series accumulate period H/L/C/O.
 * On anchor: levels computed from completed period, accumulator reset to current bar.
 * Between anchors: developing=false carries levels forward; developing=true recomputes each bar.
 * Implements TAFunctionHandler and CompositeIndicatorMetadata. */
type PivotPointLevelsHandler struct{}

func (h *PivotPointLevelsHandler) CanHandle(funcName string) bool {
	return funcName == "ta.pivot_point_levels" || funcName == "pivot_point_levels"
}

func (h *PivotPointLevelsHandler) GenerateCode(g *generator, varName string, call *ast.CallExpression) (string, error) {
	typeExpr, anchorExpr, developingExpr, err := extractPivotArguments(call)
	if err != nil {
		return "", err
	}

	pivotTypeCode, err := resolvePivotTypeCode(g, typeExpr)
	if err != nil {
		return "", err
	}

	anchorCode, err := g.generateArrowFunctionExpression(anchorExpr)
	if err != nil {
		return "", fmt.Errorf("ta.pivot_point_levels: anchor expression: %w", err)
	}
	anchorBoolCode := g.addBoolConversionIfNeeded(anchorExpr, anchorCode)

	developingCode, err := resolveDevelopingCode(g, developingExpr)
	if err != nil {
		return "", err
	}

	raw := buildPivotPerBarBlock(varName, pivotTypeCode, anchorBoolCode, developingCode)
	return g.indentCode(raw), nil
}

func (h *PivotPointLevelsHandler) GetInternalSeriesNames(varName string, _ *ast.CallExpression) ([]string, error) {
	return pivotInternalSeriesNames(varName), nil
}

func pivotInternalSeriesNames(varName string) []string {
	return []string{
		fmt.Sprintf("_%s_periodH", varName),
		fmt.Sprintf("_%s_periodL", varName),
		fmt.Sprintf("_%s_periodC", varName),
		fmt.Sprintf("_%s_periodO", varName),
	}
}

func extractPivotArguments(call *ast.CallExpression) (typeExpr, anchorExpr, developingExpr ast.Expression, err error) {
	switch len(call.Arguments) {
	case 2:
		typeExpr = call.Arguments[0]
		anchorExpr = call.Arguments[1]
		developingExpr = nil
	case 3:
		typeExpr = call.Arguments[0]
		anchorExpr = call.Arguments[1]
		developingExpr = call.Arguments[2]
	default:
		err = fmt.Errorf("ta.pivot_point_levels requires 2 or 3 arguments, got %d", len(call.Arguments))
	}
	return
}

func resolvePivotTypeCode(g *generator, typeExpr ast.Expression) (string, error) {
	if strVal := g.evaluateStringConstant(typeExpr); strVal != "" {
		return fmt.Sprintf("ta.PivotType(%q)", strVal), nil
	}
	exprCode, err := g.generateArrowFunctionExpression(typeExpr)
	if err != nil {
		return "", fmt.Errorf("ta.pivot_point_levels: type expression: %w", err)
	}
	return fmt.Sprintf("ta.PivotType(%s)", exprCode), nil
}

func resolveDevelopingCode(g *generator, developingExpr ast.Expression) (string, error) {
	if developingExpr == nil {
		return "false", nil
	}
	if lit, ok := developingExpr.(*ast.Literal); ok {
		if b, ok := lit.Value.(bool); ok {
			if b {
				return "true", nil
			}
			return "false", nil
		}
	}
	exprCode, err := g.generateArrowFunctionExpression(developingExpr)
	if err != nil {
		return "", fmt.Errorf("ta.pivot_point_levels: developing expression: %w", err)
	}
	return g.addBoolConversionIfNeeded(developingExpr, exprCode), nil
}

func buildPivotPerBarBlock(varName, pivotTypeCode, anchorBoolCode, developingCode string) string {
	ind := newBlockIndenter()

	periodH := fmt.Sprintf("_%s_periodH", varName)
	periodL := fmt.Sprintf("_%s_periodL", varName)
	periodC := fmt.Sprintf("_%s_periodC", varName)
	periodO := fmt.Sprintf("_%s_periodO", varName)

	prevH := fmt.Sprintf("_%s_prevH", varName)
	prevL := fmt.Sprintf("_%s_prevL", varName)
	prevC := fmt.Sprintf("_%s_prevC", varName)
	prevO := fmt.Sprintf("_%s_prevO", varName)

	newH := fmt.Sprintf("_%s_newH", varName)
	newL := fmt.Sprintf("_%s_newL", varName)
	newC := fmt.Sprintf("_%s_newC", varName)
	newO := fmt.Sprintf("_%s_newO", varName)

	arraySeriesVar := varName + "ArraySeries"

	var b strings.Builder

	b.WriteString(ind.line("{"))
	ind.push()

	b.WriteString(ind.linef("%s := %sSeries.Get(1)", prevH, periodH))
	b.WriteString(ind.linef("%s := %sSeries.Get(1)", prevL, periodL))
	b.WriteString(ind.linef("%s := %sSeries.Get(1)", prevC, periodC))
	b.WriteString(ind.linef("%s := %sSeries.Get(1)", prevO, periodO))

	b.WriteString(ind.linef("_pivot_type := %s", pivotTypeCode))
	b.WriteString(ind.linef("_anchor := %s", anchorBoolCode))
	b.WriteString(ind.linef("_developing := %s", developingCode))

	b.WriteString(ind.line("if _anchor {"))
	ind.push()
	b.WriteString(ind.linef("%s.Set(ta.ComputePivotLevels(_pivot_type, %s, %s, %s, %s, openSeries.Get(0)))",
		arraySeriesVar, prevH, prevL, prevC, prevO))
	b.WriteString(ind.linef("%sSeries.Set(highSeries.Get(0))", periodH))
	b.WriteString(ind.linef("%sSeries.Set(lowSeries.Get(0))", periodL))
	b.WriteString(ind.linef("%sSeries.Set(closeSeries.Get(0))", periodC))
	b.WriteString(ind.linef("%sSeries.Set(openSeries.Get(0))", periodO))
	ind.pop()

	b.WriteString(ind.line("} else {"))
	ind.push()

	b.WriteString(ind.linef("var %s, %s, %s float64", newH, newL, newO))
	b.WriteString(ind.linef("if math.IsNaN(%s) {", prevH))
	ind.push()
	b.WriteString(ind.linef("%s = highSeries.Get(0)", newH))
	b.WriteString(ind.linef("%s = lowSeries.Get(0)", newL))
	b.WriteString(ind.linef("%s = openSeries.Get(0)", newO))
	ind.pop()
	b.WriteString(ind.line("} else {"))
	ind.push()
	b.WriteString(ind.linef("%s = math.Max(%s, highSeries.Get(0))", newH, prevH))
	b.WriteString(ind.linef("%s = math.Min(%s, lowSeries.Get(0))", newL, prevL))
	b.WriteString(ind.linef("%s = %s", newO, prevO))
	ind.pop()
	b.WriteString(ind.line("}"))

	b.WriteString(ind.linef("%s := closeSeries.Get(0)", newC))
	b.WriteString(ind.linef("%sSeries.Set(%s)", periodH, newH))
	b.WriteString(ind.linef("%sSeries.Set(%s)", periodL, newL))
	b.WriteString(ind.linef("%sSeries.Set(%s)", periodC, newC))
	b.WriteString(ind.linef("%sSeries.Set(%s)", periodO, newO))

	b.WriteString(ind.line("if _developing {"))
	ind.push()
	b.WriteString(ind.linef("%s.Set(ta.ComputePivotLevels(_pivot_type, %s, %s, %s, %s, %s))",
		arraySeriesVar, newH, newL, newC, newO, newO))
	ind.pop()
	b.WriteString(ind.line("} else {"))
	ind.push()
	b.WriteString(ind.linef("if _carry := %s.Get(1); _carry != nil {", arraySeriesVar))
	ind.push()
	b.WriteString(ind.linef("%s.Set(_carry)", arraySeriesVar))
	ind.pop()
	b.WriteString(ind.line("} else {"))
	ind.push()
	b.WriteString(ind.linef("%s.Set(ta.NaNLevels())", arraySeriesVar))
	ind.pop()
	b.WriteString(ind.line("}"))
	ind.pop()
	b.WriteString(ind.line("}"))

	ind.pop()
	b.WriteString(ind.line("}"))
	ind.pop()
	b.WriteString(ind.line("}"))

	return b.String()
}

type blockIndenter struct {
	depth int
}

func newBlockIndenter() *blockIndenter {
	return &blockIndenter{}
}

func (bi *blockIndenter) push() { bi.depth++ }
func (bi *blockIndenter) pop()  { bi.depth-- }

func (bi *blockIndenter) line(s string) string {
	return strings.Repeat("\t", bi.depth) + s + "\n"
}

func (bi *blockIndenter) linef(format string, args ...interface{}) string {
	return bi.line(fmt.Sprintf(format, args...))
}
