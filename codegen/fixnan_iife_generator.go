package codegen

import "fmt"

type SelfReferencingIIFEGenerator interface {
	GenerateWithSelfReference(accessor AccessGenerator, targetSeriesVar string) string
}

type FixnanIIFEGenerator struct{}

func (g *FixnanIIFEGenerator) GenerateWithSelfReference(accessor AccessGenerator, targetSeriesVar string) string {
	body := "val := " + accessor.GenerateLoopValueAccess("0") + "; "
	body += "if math.IsNaN(val) { return 0.0 }; "
	body += "return val"

	return "func() float64 { " + body + " }()"
}

type FixnanCallExpressionAccessor struct {
	tempVarName string
	tempVarCode string
	exprCode    string // Expression code without variable assignment
}

func (a *FixnanCallExpressionAccessor) GenerateLoopValueAccess(loopVar string) string {
	// If exprCode is empty, fall back to temp variable (for backward compatibility)
	if a.exprCode == "" {
		return a.tempVarName
	}

	// If expression contains Series access, transform .GetCurrent() to .Get(loopVar)
	if containsGetCurrent(a.exprCode) {
		return transformSeriesAccess(a.exprCode, loopVar)
	}

	// Otherwise, return temp variable (expression doesn't need historical access)
	return a.tempVarName
}

func (a *FixnanCallExpressionAccessor) GenerateInitialValueAccess(period int) string {
	// If exprCode is empty, fall back to temp variable
	if a.exprCode == "" {
		return a.tempVarName
	}

	// If expression contains Series access, transform .GetCurrent() to .Get(period-1)
	// The initial value for RMA/EMA is at the oldest point in the period window
	if containsGetCurrent(a.exprCode) {
		offset := period - 1
		return transformSeriesAccess(a.exprCode, fmt.Sprintf("%d", offset))
	}

	// Otherwise, return temp variable (expression doesn't need historical access)
	return a.tempVarName
}

func (a *FixnanCallExpressionAccessor) GenerateCurrentValueAccess() string {
	// If exprCode is empty, fall back to temp variable
	if a.exprCode == "" {
		return a.tempVarName
	}

	// For current bar access, return the expression as-is (no transformation needed)
	return a.exprCode
}

func (a *FixnanCallExpressionAccessor) GetPreamble() string {
	// If expression contains Series access, don't generate preamble
	// The expression will be generated inline at each access point
	if a.exprCode != "" && containsGetCurrent(a.exprCode) {
		return ""
	}
	return a.tempVarCode
}

/* transformSeriesAccess replaces .GetCurrent() with .Get(offset) for historical Series access */
func transformSeriesAccess(exprCode, loopVar string) string {
	// Replace all occurrences of .GetCurrent() with .Get(loopVar)
	result := ""
	remaining := exprCode

	for {
		idx := indexOf(remaining, ".GetCurrent()")
		if idx == -1 {
			result += remaining
			break
		}
		result += remaining[:idx]
		result += ".Get(" + loopVar + ")"
		remaining = remaining[idx+len(".GetCurrent()"):]
	}

	return result
}

/* containsGetCurrent checks if expression contains Series .GetCurrent() calls */
func containsGetCurrent(exprCode string) bool {
	return indexOf(exprCode, ".GetCurrent()") != -1
}

/* indexOf returns the index of substr in s, or -1 if not found */
func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
