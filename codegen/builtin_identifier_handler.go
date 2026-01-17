package codegen

import (
	"fmt"

	"github.com/quant5-lab/runner/ast"
)

// BuiltinIdentifierHandler resolves Pine Script built-in identifiers to Go runtime expressions.
type BuiltinIdentifierHandler struct{}

func NewBuiltinIdentifierHandler() *BuiltinIdentifierHandler {
	return &BuiltinIdentifierHandler{}
}

// IsBuiltinSeriesIdentifier checks if identifier is a Pine built-in series variable.
func (h *BuiltinIdentifierHandler) IsBuiltinSeriesIdentifier(name string) bool {
	switch name {
	case "close", "open", "high", "low", "volume", "tr", "bar_index",
		"hl2", "hlc3", "ohlc4", "hlcc4":
		return true
	default:
		return false
	}
}

// IsStrategyRuntimeValue checks if member expression is a strategy runtime value.
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

// GenerateCurrentBarAccess generates code for built-in series at current bar.
func (h *BuiltinIdentifierHandler) GenerateCurrentBarAccess(name string) string {
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
	case "hl2":
		return "((bar.High + bar.Low) / 2)"
	case "hlc3":
		return "((bar.High + bar.Low + bar.Close) / 3)"
	case "ohlc4":
		return "((bar.Open + bar.High + bar.Low + bar.Close) / 4)"
	case "hlcc4":
		return "((bar.High + bar.Low + bar.Close + bar.Close) / 4)"
	default:
		return ""
	}
}

// GenerateSecurityContextAccess generates code for built-in series in security() context.
func (h *BuiltinIdentifierHandler) GenerateSecurityContextAccess(name string) string {
	switch name {
	case "close":
		return "ctx.Data[ctx.BarIndex].Close"
	case "open":
		return "ctx.Data[ctx.BarIndex].Open"
	case "high":
		return "ctx.Data[ctx.BarIndex].High"
	case "low":
		return "ctx.Data[ctx.BarIndex].Low"
	case "volume":
		return "ctx.Data[ctx.BarIndex].Volume"
	case "tr":
		return h.generateTrueRangeCalculation("ctx.Data[ctx.BarIndex]")
	case "bar_index":
		return "float64(ctx.BarIndex)"
	case "hl2":
		return "((ctx.Data[ctx.BarIndex].High + ctx.Data[ctx.BarIndex].Low) / 2)"
	case "hlc3":
		return "((ctx.Data[ctx.BarIndex].High + ctx.Data[ctx.BarIndex].Low + ctx.Data[ctx.BarIndex].Close) / 3)"
	case "ohlc4":
		return "((ctx.Data[ctx.BarIndex].Open + ctx.Data[ctx.BarIndex].High + ctx.Data[ctx.BarIndex].Low + ctx.Data[ctx.BarIndex].Close) / 4)"
	case "hlcc4":
		return "((ctx.Data[ctx.BarIndex].High + ctx.Data[ctx.BarIndex].Low + ctx.Data[ctx.BarIndex].Close + ctx.Data[ctx.BarIndex].Close) / 4)"
	default:
		return ""
	}
}

// GenerateHistoricalAccess generates code for historical built-in series access with bounds checking.
func (h *BuiltinIdentifierHandler) GenerateHistoricalAccess(name string, offset int) string {
	if name == "tr" {
		return h.generateHistoricalTrueRange(offset)
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

// GenerateStrategyRuntimeAccess generates Series access for strategy runtime values.
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

// TryResolveIdentifier attempts to resolve identifier as builtin.
func (h *BuiltinIdentifierHandler) TryResolveIdentifier(expr *ast.Identifier, inSecurityContext bool) (string, bool) {
	if expr.Name == "na" {
		return "math.NaN()", true
	}

	if !h.IsBuiltinSeriesIdentifier(expr.Name) {
		return "", false
	}

	if inSecurityContext {
		return h.GenerateSecurityContextAccess(expr.Name), true
	}

	return h.GenerateCurrentBarAccess(expr.Name), true
}

// TryResolveMemberExpression attempts to resolve member expression as builtin.
func (h *BuiltinIdentifierHandler) TryResolveMemberExpression(expr *ast.MemberExpression, inSecurityContext bool) (string, bool) {
	obj, okObj := expr.Object.(*ast.Identifier)
	if !okObj {
		return "", false
	}

	prop, okProp := expr.Property.(*ast.Identifier)
	if !okProp && !expr.Computed {
		return "", false
	}

	// Strategy runtime values (non-computed member access)
	if okProp && h.IsStrategyRuntimeValue(obj.Name, prop.Name) {
		return h.GenerateStrategyRuntimeAccess(prop.Name), true
	}

	// Strategy constants (handled elsewhere)
	if okProp && obj.Name == "strategy" && (prop.Name == "long" || prop.Name == "short") {
		return "", false
	}

	// Built-in series with subscript access
	if h.IsBuiltinSeriesIdentifier(obj.Name) && expr.Computed {
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

// generateTrueRangeCalculation generates inline tr calculation.
func (h *BuiltinIdentifierHandler) generateTrueRangeCalculation(barAccessor string) string {
	return fmt.Sprintf(
		"func() float64 { if ctx.BarIndex < 1 { return %s.High - %s.Low }; "+
			"prevClose := ctx.Data[ctx.BarIndex-1].Close; "+
			"return math.Max(%s.High - %s.Low, math.Max(math.Abs(%s.High - prevClose), math.Abs(%s.Low - prevClose))) }()",
		barAccessor, barAccessor,
		barAccessor, barAccessor, barAccessor, barAccessor,
	)
}

// generateHistoricalTrueRange generates tr calculation for historical bar access with offset.
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
