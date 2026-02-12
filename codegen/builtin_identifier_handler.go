package codegen

import (
	"fmt"

	"github.com/quant5-lab/runner/ast"
)

type BuiltinIdentifierHandler struct {
	registry          *BuiltinIdentifierRegistry
	formulaGen        *DerivedPriceFormulaGenerator
	colorResolver     *ColorConstantResolver
	namespaceResolver *BuiltinNamespaceResolver
}

func NewBuiltinIdentifierHandler() *BuiltinIdentifierHandler {
	return &BuiltinIdentifierHandler{
		registry:          NewBuiltinIdentifierRegistry(),
		formulaGen:        NewDerivedPriceFormulaGenerator(),
		colorResolver:     NewColorConstantResolver(),
		namespaceResolver: NewBuiltinNamespaceResolver(),
	}
}

func (h *BuiltinIdentifierHandler) IsBuiltinSeriesIdentifier(name string) bool {
	if h == nil || h.registry == nil {
		return false
	}
	return h.registry.IsBuiltinSeriesIdentifier(name)
}

func (h *BuiltinIdentifierHandler) IsStrategyRuntimeValue(obj, prop string) bool {
	if obj != "strategy" {
		return false
	}
	switch prop {
	case "position_avg_price", "position_size", "position_entry_name",
		"equity", "netprofit", "closedtrades":
		return true
	default:
		return false
	}
}

func (h *BuiltinIdentifierHandler) GenerateCurrentBarAccess(name string) string {
	if h.registry.IsDerivedPrice(name) {
		return h.formulaGen.Generate(name, "bar.High", "bar.Low", "bar.Close", "bar.Open")
	}

	switch name {
	case "close":
		return "bar.Close"
	case "open":
		return "bar.Open"
	case "high":
		return "bar.High"
	case "low":
		return "bar.Low"
	case "volume":
		return "bar.Volume"
	case "tr":
		return h.generateTrueRangeCalculation("bar")
	case "bar_index":
		return "float64(i)"
	case "time":
		return "float64(bar.Time * 1000)"
	default:
		return ""
	}
}

func (h *BuiltinIdentifierHandler) GenerateSecurityContextAccess(name string) string {
	if h.registry.IsDerivedPrice(name) {
		return h.formulaGen.Generate(name,
			"highSeries.GetCurrent()",
			"lowSeries.GetCurrent()",
			"closeSeries.GetCurrent()",
			"openSeries.GetCurrent()")
	}

	switch name {
	case "close":
		return "closeSeries.GetCurrent()"
	case "open":
		return "openSeries.GetCurrent()"
	case "high":
		return "highSeries.GetCurrent()"
	case "low":
		return "lowSeries.GetCurrent()"
	case "volume":
		return "volumeSeries.GetCurrent()"
	case "tr":
		return h.generateTrueRangeCalculationSeries()
	case "bar_index":
		return "float64(ctx.BarIndex)"
	case "time":
		return "timeSeries.GetCurrent()"
	default:
		return ""
	}
}

func (h *BuiltinIdentifierHandler) GenerateHistoricalAccess(name string, offset int) string {
	if name == "tr" {
		return h.generateHistoricalTrueRange(offset)
	}

	if h.registry.IsDerivedPrice(name) {
		accessor := NewDerivedPriceAccessor(name, offset)
		formula := accessor.GenerateFormulaAtOffset(fmt.Sprintf("i-%d", offset))
		return fmt.Sprintf("func() float64 { if i-%d >= 0 { return %s }; return math.NaN() }()", offset, formula)
	}

	if name == "bar_index" {
		return fmt.Sprintf("bar_indexSeries.Get(%d)", offset)
	}

	if name == "time" {
		return fmt.Sprintf("timeSeries.Get(%d)", offset)
	}

	field := ""
	switch name {
	case "close":
		field = "Close"
	case "open":
		field = "Open"
	case "high":
		field = "High"
	case "low":
		field = "Low"
	case "volume":
		field = "Volume"
	default:
		return ""
	}

	return fmt.Sprintf("func() float64 { if i-%d >= 0 { return ctx.Data[i-%d].%s }; return math.NaN() }()",
		offset, offset, field)
}

func (h *BuiltinIdentifierHandler) GenerateStrategyRuntimeAccess(property string) string {
	switch property {
	case "position_avg_price":
		return "strategy_position_avg_priceSeries.Get(0)"
	case "position_size":
		return "strategy_position_sizeSeries.Get(0)"
	case "position_entry_name":
		return "strat.GetPositionEntryName()"
	case "equity":
		return "strategy_equitySeries.Get(0)"
	case "netprofit":
		return "strategy_netprofitSeries.Get(0)"
	case "closedtrades":
		return "strategy_closedtradesSeries.Get(0)"
	default:
		return ""
	}
}

func (h *BuiltinIdentifierHandler) IsColorIdentifier(name string) bool {
	if h == nil || h.colorResolver == nil {
		return false
	}
	return h.colorResolver.IsColorIdentifier(name)
}

func (h *BuiltinIdentifierHandler) ResolveColorHex(name string) (string, bool) {
	if h == nil || h.colorResolver == nil {
		return "", false
	}
	return h.colorResolver.ResolveIdentifierToHex(name)
}

func (h *BuiltinIdentifierHandler) ResolveMemberExpressionColorHex(expr *ast.MemberExpression) (string, bool) {
	if h == nil || h.colorResolver == nil {
		return "", false
	}
	return h.colorResolver.ResolveMemberExpressionToHex(expr)
}

func (h *BuiltinIdentifierHandler) TryResolveIdentifier(expr *ast.Identifier, inSecurityContext bool) (string, bool) {
	if h == nil || h.colorResolver == nil {
		return "", false
	}

	if expr.Name == "na" {
		return "math.NaN()", true
	}

	if hex, found := h.colorResolver.ResolveIdentifierToHex(expr.Name); found {
		return fmt.Sprintf("%q", hex), true
	}

	if !h.IsBuiltinSeriesIdentifier(expr.Name) {
		return "", false
	}

	if inSecurityContext {
		return h.GenerateSecurityContextAccess(expr.Name), true
	}

	return h.GenerateCurrentBarAccess(expr.Name), true
}

func (h *BuiltinIdentifierHandler) TryResolveMemberExpression(expr *ast.MemberExpression, inSecurityContext bool) (string, bool) {
	obj, okObj := expr.Object.(*ast.Identifier)
	if !okObj {
		if objMember, ok := expr.Object.(*ast.MemberExpression); ok && expr.Computed {
			baseObj, baseOk := objMember.Object.(*ast.Identifier)
			baseProp, basePropOk := objMember.Property.(*ast.Identifier)

			if baseOk && basePropOk && baseObj.Name == "ta" && baseProp.Name == "tr" {
				offset := h.extractOffset(expr.Property)
				return h.generateHistoricalTrueRange(offset), true
			}
		}
		return "", false
	}

	prop, okProp := expr.Property.(*ast.Identifier)
	if !okProp && !expr.Computed {
		return "", false
	}

	if okProp && obj.Name == "ta" && prop.Name == "tr" {
		return h.GenerateCurrentBarAccess("tr"), true
	}

	if okProp && h.IsStrategyRuntimeValue(obj.Name, prop.Name) {
		return h.GenerateStrategyRuntimeAccess(prop.Name), true
	}

	if okProp && obj.Name == "strategy" && (prop.Name == "long" || prop.Name == "short") {
		return "", false
	}

	if okProp && h.namespaceResolver != nil {
		if resolution, found := h.namespaceResolver.Resolve(obj.Name, prop.Name); found {
			return resolution.Code, true
		}
	}

	if h.IsBuiltinSeriesIdentifier(obj.Name) && expr.Computed {
		// Delegate variable subscripts to subscriptResolver for loop counter handling
		if _, isLiteral := expr.Property.(*ast.Literal); !isLiteral {
			return "", false
		}

		offset := h.extractOffset(expr.Property)
		if offset == 0 {
			if inSecurityContext {
				return h.GenerateSecurityContextAccess(obj.Name), true
			}
			return h.GenerateCurrentBarAccess(obj.Name), true
		}
		return h.GenerateHistoricalAccess(obj.Name, offset), true
	}

	return "", false
}

func (h *BuiltinIdentifierHandler) extractOffset(expr ast.Expression) int {
	lit, ok := expr.(*ast.Literal)
	if !ok {
		return 0
	}

	switch v := lit.Value.(type) {
	case float64:
		return int(v)
	case int:
		return v
	default:
		return 0
	}
}

func (h *BuiltinIdentifierHandler) generateTrueRangeCalculation(barAccessor string) string {
	return fmt.Sprintf(
		"func() float64 { if ctx.BarIndex < 1 { return %s.High - %s.Low }; "+
			"prevClose := ctx.Data[ctx.BarIndex-1].Close; "+
			"return math.Max(%s.High - %s.Low, math.Max(math.Abs(%s.High - prevClose), math.Abs(%s.Low - prevClose))) }()",
		barAccessor, barAccessor,
		barAccessor, barAccessor, barAccessor, barAccessor,
	)
}

func (h *BuiltinIdentifierHandler) generateTrueRangeCalculationSeries() string {
	return "func() float64 { if ctx.BarIndex < 1 { return highSeries.GetCurrent() - lowSeries.GetCurrent() }; " +
		"prevClose := closeSeries.Get(1); " +
		"return math.Max(highSeries.GetCurrent() - lowSeries.GetCurrent(), math.Max(math.Abs(highSeries.GetCurrent() - prevClose), math.Abs(lowSeries.GetCurrent() - prevClose))) }()"
}

func (h *BuiltinIdentifierHandler) generateHistoricalTrueRange(offset int) string {
	return fmt.Sprintf(
		"func() float64 { "+
			"if i-%d < 0 { return math.NaN() }; "+
			"barIdx := i-%d; "+
			"if barIdx < 1 { return ctx.Data[barIdx].High - ctx.Data[barIdx].Low }; "+
			"prevClose := ctx.Data[barIdx-1].Close; "+
			"currentBar := ctx.Data[barIdx]; "+
			"return math.Max(currentBar.High - currentBar.Low, math.Max(math.Abs(currentBar.High - prevClose), math.Abs(currentBar.Low - prevClose))) "+
			"}()",
		offset, offset,
	)
}
