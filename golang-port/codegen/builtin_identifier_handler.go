package codegen

import (
	"fmt"

	"github.com/quant5-lab/runner/ast"
)

// BuiltinIdentifierHandler resolves Pine Script built-in identifiers to Go runtime expressions.
//
// Responsibilities:
//   - Detect built-in series (close, open, high, low, volume)
//   - Detect strategy runtime values (strategy.position_avg_price, etc.)
//   - Generate correct Go code for each context (current bar vs security() context)
//
// Design: Centralized builtin detection prevents duplicate switch statements across generator.
type BuiltinIdentifierHandler struct{}

func NewBuiltinIdentifierHandler() *BuiltinIdentifierHandler {
	return &BuiltinIdentifierHandler{}
}

// IsBuiltinSeriesIdentifier checks if identifier is a Pine built-in series variable.
func (h *BuiltinIdentifierHandler) IsBuiltinSeriesIdentifier(name string) bool {
	switch name {
	case "close", "open", "high", "low", "volume":
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
	case "position_avg_price", "position_size", "position_entry_name":
		return true
	default:
		return false
	}
}

// GenerateCurrentBarAccess generates code for built-in series at current bar.
//
// Returns: bar.Close, bar.Open, etc.
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
	default:
		return ""
	}
}

// GenerateSecurityContextAccess generates code for built-in series in security() context.
//
// Why different: security() processes historical data, needs ctx.Data[ctx.BarIndex] access.
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
	default:
		return ""
	}
}

// GenerateHistoricalAccess generates code for historical built-in series access.
//
// Returns: Bounds-checked historical access with NaN fallback.
func (h *BuiltinIdentifierHandler) GenerateHistoricalAccess(name string, offset int) string {
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

// GenerateStrategyRuntimeAccess generates Series.Get(0) access for strategy runtime values
func (h *BuiltinIdentifierHandler) GenerateStrategyRuntimeAccess(property string) string {
	switch property {
	case "position_avg_price":
		return "strategy_position_avg_priceSeries.Get(0)"
	case "position_size":
		return "strategy_position_sizeSeries.Get(0)"
	case "position_entry_name":
		return "strat.GetPositionEntryName()"
	default:
		return ""
	}
}

// TryResolveIdentifier attempts to resolve identifier as builtin.
//
// Returns: (code, resolved)
//   - If builtin: (generated code, true)
//   - If not builtin: ("", false)
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
//
// Returns: (code, resolved)
//   - If builtin: (generated code, true)
//   - If not builtin: ("", false)
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
