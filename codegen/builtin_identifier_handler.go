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
	arrowGen          *ArrowBuiltinAccessGenerator
}

func NewBuiltinIdentifierHandler() *BuiltinIdentifierHandler {
	registry := NewBuiltinIdentifierRegistry()
	formulaGen := NewDerivedPriceFormulaGenerator()
	return &BuiltinIdentifierHandler{
		registry:          registry,
		formulaGen:        formulaGen,
		colorResolver:     NewColorConstantResolver(),
		namespaceResolver: NewBuiltinNamespaceResolver(),
		arrowGen:          NewArrowBuiltinAccessGenerator(registry, formulaGen),
	}
}

func (h *BuiltinIdentifierHandler) IsBuiltinSeriesIdentifier(name string) bool {
	if h == nil || h.registry == nil {
		return false
	}
	return h.registry.IsBuiltinSeriesIdentifier(name)
}

func (h *BuiltinIdentifierHandler) IsConstantBuiltin(name string) bool {
	if h == nil || h.registry == nil {
		return false
	}
	return h.registry.IsConstantBuiltin(name)
}

func (h *BuiltinIdentifierHandler) IsStrategyRuntimeValue(obj, prop string) bool {
	if obj != "strategy" {
		return false
	}
	switch prop {
	case "position_avg_price", "position_size", "position_entry_name",
		"equity", "netprofit", "closedtrades",
		"initial_capital", "grossprofit", "grossloss",
		"wintrades", "losstrades", "eventrades":
		return true
	default:
		return false
	}
}

func (h *BuiltinIdentifierHandler) GenerateCurrentBarAccess(name string) string {
	if h.registry.IsDerivedPrice(name) {
		return h.formulaGen.Generate(name, "bar.High", "bar.Low", "bar.Close", "bar.Open")
	}

	if info, ok := h.registry.CalendarInfo(name); ok {
		return info.SeriesName + ".GetCurrent()"
	}

	if field, ok := OHLCVFieldName(name); ok {
		return fmt.Sprintf("bar.%s", field)
	}

	switch name {
	case "tr":
		return h.generateTrueRangeCalculation("bar")
	case "bar_index":
		return "float64(i)"
	case "time":
		return "float64(bar.Time * 1000)"
	case "time_close":
		return "time_closeSeries.GetCurrent()"
	case "time_tradingday":
		return "time_tradingdaySeries.GetCurrent()"
	case "last_bar_index":
		return "last_bar_index"
	case "last_bar_time":
		return "last_bar_time"
	case "timenow":
		return "timenow"
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

	if info, ok := h.registry.CalendarInfo(name); ok {
		return info.SeriesName + ".GetCurrent()"
	}

	if _, ok := OHLCVFieldName(name); ok {
		return fmt.Sprintf("%sSeries.GetCurrent()", name)
	}

	switch name {
	case "tr":
		return h.generateTrueRangeCalculationSeries()
	case "bar_index":
		return "float64(ctx.BarIndex)"
	case "time":
		return "timeSeries.GetCurrent()"
	case "time_close":
		return "time_closeSeries.GetCurrent()"
	case "time_tradingday":
		return "time_tradingdaySeries.GetCurrent()"
	case "last_bar_index":
		return "last_bar_index"
	case "last_bar_time":
		return "last_bar_time"
	case "timenow":
		return "timenow"
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

	if info, ok := h.registry.CalendarInfo(name); ok {
		return fmt.Sprintf("%s.Get(%d)", info.SeriesName, offset)
	}

	if name == "bar_index" {
		return fmt.Sprintf("bar_indexSeries.Get(%d)", offset)
	}

	if name == "time" {
		return fmt.Sprintf("timeSeries.Get(%d)", offset)
	}

	if name == "time_close" {
		return fmt.Sprintf("time_closeSeries.Get(%d)", offset)
	}

	if name == "time_tradingday" {
		return fmt.Sprintf("time_tradingdaySeries.Get(%d)", offset)
	}

	if name == "last_bar_index" {
		return "last_bar_index"
	}

	if name == "last_bar_time" {
		return "last_bar_time"
	}

	if name == "timenow" {
		return "timenow"
	}

	field, ok := OHLCVFieldName(name)
	if !ok {
		return ""
	}

	return fmt.Sprintf("func() float64 { if i-%d >= 0 { return ctx.Data[i-%d].%s }; return math.NaN() }()",
		offset, offset, field)
}

func (h *BuiltinIdentifierHandler) GenerateStrategyRuntimeAccess(property string) string {
	switch property {
	case "position_avg_price":
		return StrategyPositionAvgPriceSeriesName + ".Get(0)"
	case "position_size":
		return StrategyPositionSizeSeriesName + ".Get(0)"
	case "position_entry_name":
		return "strat.GetPositionEntryName()"
	case "equity":
		return StrategyEquitySeriesName + ".Get(0)"
	case "netprofit":
		return StrategyNetProfitSeriesName + ".Get(0)"
	case "closedtrades":
		return StrategyClosedTradesSeriesName + ".Get(0)"
	case "initial_capital":
		return "strat.GetInitialCapital()"
	case "grossprofit":
		return "strat.GetGrossProfit()"
	case "grossloss":
		return "strat.GetGrossLoss()"
	case "wintrades":
		return "float64(strat.GetWinningTradesCount())"
	case "losstrades":
		return "float64(strat.GetLosingTradesCount())"
	case "eventrades":
		return "float64(strat.GetEvenTradesCount())"
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

func (h *BuiltinIdentifierHandler) CalendarBuiltinNames() []string {
	return h.registry.CalendarBuiltinNames()
}

func (h *BuiltinIdentifierHandler) CalendarInfo(name string) (CalendarBuiltinInfo, bool) {
	return h.registry.CalendarInfo(name)
}

func (h *BuiltinIdentifierHandler) IsDerivedPrice(name string) bool {
	return h.registry.IsDerivedPrice(name)
}

func (h *BuiltinIdentifierHandler) generateBuiltinAccess(name string, scope AccessScope) string {
	switch scope {
	case ArrowScope:
		return h.arrowGen.GenerateCurrentAccess(name)
	case SecurityScope:
		return h.GenerateSecurityContextAccess(name)
	default:
		return h.GenerateCurrentBarAccess(name)
	}
}

func (h *BuiltinIdentifierHandler) generateHistoricalBuiltinAccess(name string, offset int, scope AccessScope) string {
	if scope == ArrowScope {
		return h.arrowGen.GenerateHistoricalAccess(name, offset)
	}
	return h.GenerateHistoricalAccess(name, offset)
}

func (h *BuiltinIdentifierHandler) generateStrategyAccess(property string, scope AccessScope) string {
	if scope == ArrowScope {
		return h.arrowGen.GenerateStrategyAccess(property)
	}
	return h.GenerateStrategyRuntimeAccess(property)
}

func (h *BuiltinIdentifierHandler) resolveNamespace(obj, prop string, scope AccessScope) (NamespaceResolution, bool) {
	if scope == ArrowScope {
		return h.namespaceResolver.ResolveForArrow(obj, prop)
	}
	return h.namespaceResolver.Resolve(obj, prop)
}

func (h *BuiltinIdentifierHandler) resolveNestedMemberExpression(expr *ast.MemberExpression, scope AccessScope) (string, bool) {
	objMember, ok := expr.Object.(*ast.MemberExpression)
	if !ok || !expr.Computed {
		return "", false
	}

	baseObj, baseOk := objMember.Object.(*ast.Identifier)
	baseProp, basePropOk := objMember.Property.(*ast.Identifier)
	if !baseOk || !basePropOk {
		return "", false
	}

	if baseObj.Name == "ta" && baseProp.Name == "tr" {
		offset := h.extractOffset(expr.Property)
		if scope == ArrowScope {
			return TrueRangeArrowIIFE(fmt.Sprintf("%d", offset)), true
		}
		return h.generateHistoricalTrueRange(offset), true
	}

	key := baseObj.Name + "." + baseProp.Name
	if h.registry.IsSessionSeriesBuiltin(key) {
		offset := h.extractOffset(expr.Property)
		seriesName := SessionSeriesName(key)
		if scope == ArrowScope {
			return SeriesLookupWithOffsetIIFE(seriesName, offset), true
		}
		if offset == 0 {
			return fmt.Sprintf("%s.GetCurrent() == 1.0", seriesName), true
		}
		return fmt.Sprintf("%s.Get(%d) == 1.0", seriesName, offset), true
	}

	return "", false
}

func (h *BuiltinIdentifierHandler) GenerateDerivedPriceFormula(name, highAccess, lowAccess, closeAccess, openAccess string) string {
	return h.formulaGen.Generate(name, highAccess, lowAccess, closeAccess, openAccess)
}

func (h *BuiltinIdentifierHandler) ResolveCalendarBuiltins(detectedNames map[string]bool) []CalendarBuiltinInfo {
	var resolved []CalendarBuiltinInfo
	for name := range detectedNames {
		if info, ok := h.registry.CalendarInfo(name); ok {
			resolved = append(resolved, info)
		}
	}
	return resolved
}

func (h *BuiltinIdentifierHandler) TryResolveIdentifier(expr *ast.Identifier, scope AccessScope) (string, bool) {
	if h == nil || h.colorResolver == nil {
		return "", false
	}

	if expr.Name == "na" {
		return "math.NaN()", true
	}

	if hex, found := h.colorResolver.ResolveIdentifierToHex(expr.Name); found {
		return fmt.Sprintf("%q", hex), true
	}

	if h.IsBuiltinSeriesIdentifier(expr.Name) || h.registry.IsConstantBuiltin(expr.Name) {
		code := h.generateBuiltinAccess(expr.Name, scope)
		if code != "" {
			return code, true
		}
	}

	return "", false
}

func (h *BuiltinIdentifierHandler) TryResolveMemberExpression(expr *ast.MemberExpression, scope AccessScope) (string, bool) {
	obj, okObj := expr.Object.(*ast.Identifier)
	if !okObj {
		return h.resolveNestedMemberExpression(expr, scope)
	}

	prop, okProp := expr.Property.(*ast.Identifier)
	if !okProp && !expr.Computed {
		return "", false
	}

	if okProp && obj.Name == "ta" && prop.Name == "tr" {
		return h.generateBuiltinAccess("tr", scope), true
	}

	if okProp && h.IsStrategyRuntimeValue(obj.Name, prop.Name) {
		return h.generateStrategyAccess(prop.Name, scope), true
	}

	if okProp && obj.Name == "strategy" && (prop.Name == "long" || prop.Name == "short") {
		return "", false
	}

	if okProp && h.namespaceResolver != nil {
		if resolution, found := h.resolveNamespace(obj.Name, prop.Name, scope); found {
			return resolution.Code, true
		}
	}

	if h.IsBuiltinSeriesIdentifier(obj.Name) && expr.Computed {
		if _, isLiteral := expr.Property.(*ast.Literal); !isLiteral {
			return "", false
		}

		offset := h.extractOffset(expr.Property)
		if offset == 0 {
			code := h.generateBuiltinAccess(obj.Name, scope)
			if code != "" {
				return code, true
			}
		}
		return h.generateHistoricalBuiltinAccess(obj.Name, offset, scope), true
	}

	return "", false
}

/* Type-only resolution for callers needing coercion without full code generation */
func (h *BuiltinIdentifierHandler) ResolveMemberExpressionGoType(expr *ast.MemberExpression) (GoValueType, bool) {
	obj, okObj := expr.Object.(*ast.Identifier)
	if !okObj {
		return GoFloat64, false
	}
	prop, okProp := expr.Property.(*ast.Identifier)
	if !okProp {
		return GoFloat64, false
	}
	if h.namespaceResolver != nil {
		if resolution, found := h.namespaceResolver.Resolve(obj.Name, prop.Name); found {
			return resolution.GoType, true
		}
	}
	return GoFloat64, false
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
