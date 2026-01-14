package codegen

import (
	"encoding/json"
	"fmt"
	"strconv"
)

// LiteralFormatter formats Pine Script literal values into Go code.
// Preserves full numeric precision for financial calculations.
type LiteralFormatter struct{}

func NewLiteralFormatter() *LiteralFormatter {
	return &LiteralFormatter{}
}

// FormatFloat converts float64 to Go literal with full precision.
// Uses %g format to preserve all significant digits while stripping trailing zeros.
//
// Examples:
//
//	0.001   → "0.001"    (preserves small precision values)
//	711.6   → "711.6"    (normal values)
//	2000000 → "2e+06"    (scientific notation for large numbers)
//	0.35    → "0.35"     (standard decimal)
func (f *LiteralFormatter) FormatFloat(value float64) string {
	return strconv.FormatFloat(value, 'g', -1, 64)
}

// FormatString converts string to Go quoted literal.
func (f *LiteralFormatter) FormatString(value string) string {
	return fmt.Sprintf("%q", value)
}

// FormatBool converts bool to Go literal.
func (f *LiteralFormatter) FormatBool(value bool) string {
	return fmt.Sprintf("%t", value)
}

// FormatGeneric handles arbitrary types via JSON marshaling.
func (f *LiteralFormatter) FormatGeneric(value interface{}) (string, error) {
	jsonBytes, err := json.Marshal(value)
	if err != nil {
		return "", fmt.Errorf("failed to marshal literal: %w", err)
	}
	return string(jsonBytes), nil
}
