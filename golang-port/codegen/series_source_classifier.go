package codegen

import (
	"regexp"

	"github.com/quant5-lab/runner/ast"
)

// SourceType distinguishes between user-defined Series and built-in OHLCV fields.
// This classification drives code generation strategy selection.
type SourceType int

const (
	SourceTypeUnknown SourceType = iota
	SourceTypeSeriesVariable // User variable: myVar, cagr5
	SourceTypeOHLCVField     // Built-in field: close, high, low, open, volume
)

// SourceInfo encapsulates classified source expression metadata for code generation.
type SourceInfo struct {
	Type         SourceType
	VariableName string
	FieldName    string
	OriginalExpr string
	BaseOffset   int // Historical lookback offset
}

// IsSeriesVariable returns true if the source is a user-defined Series variable.
func (s SourceInfo) IsSeriesVariable() bool {
	return s.Type == SourceTypeSeriesVariable
}

// IsOHLCVField returns true if the source is a built-in OHLCV field.
func (s SourceInfo) IsOHLCVField() bool {
	return s.Type == SourceTypeOHLCVField
}

// SeriesSourceClassifier determines source expression type from AST nodes.
// Distinguishes built-in OHLCV fields from user variables and extracts historical offsets.
type SeriesSourceClassifier struct {
	seriesVariablePattern *regexp.Regexp
}

// NewSeriesSourceClassifier creates classifier for analyzing source expressions.
func NewSeriesSourceClassifier() *SeriesSourceClassifier {
	return &SeriesSourceClassifier{
		seriesVariablePattern: regexp.MustCompile(`^([A-Za-z_][A-Za-z0-9_]*)Series\.Get(?:Current)?\(`),
	}
}

// ClassifyAST analyzes AST expression, extracting type and historical offset.
// Handles Identifier and MemberExpression nodes, falling back to Close for unknown expressions.
func (c *SeriesSourceClassifier) ClassifyAST(expr ast.Expression) SourceInfo {
	info := SourceInfo{}

	switch e := expr.(type) {
	case *ast.Identifier:
		if c.isBuiltinOHLCVField(e.Name) {
			info.Type = SourceTypeOHLCVField
			info.FieldName = c.capitalizeOHLCVField(e.Name)
			info.BaseOffset = 0
			return info
		}
		info.Type = SourceTypeSeriesVariable
		info.VariableName = e.Name
		info.BaseOffset = 0
		return info

	case *ast.MemberExpression:
		if obj, ok := e.Object.(*ast.Identifier); ok && e.Computed {
			offset := 0
			if lit, ok := e.Property.(*ast.Literal); ok {
				switch v := lit.Value.(type) {
				case float64:
					offset = int(v)
				case int:
					offset = v
				}
			}

			if c.isBuiltinOHLCVField(obj.Name) {
				info.Type = SourceTypeOHLCVField
				info.FieldName = c.capitalizeOHLCVField(obj.Name)
				info.BaseOffset = offset
				return info
			}
			info.Type = SourceTypeSeriesVariable
			info.VariableName = obj.Name
			info.BaseOffset = offset
			return info
		}
	}

	info.Type = SourceTypeOHLCVField
	info.FieldName = "Close"
	info.BaseOffset = 0
	return info
}

// isBuiltinOHLCVField checks if identifier is built-in OHLCV field.
func (c *SeriesSourceClassifier) isBuiltinOHLCVField(name string) bool {
	return name == "close" || name == "open" || name == "high" || name == "low" || name == "volume"
}

// capitalizeOHLCVField converts Pine field name to Go struct field name.
func (c *SeriesSourceClassifier) capitalizeOHLCVField(name string) string {
	switch name {
	case "close":
		return "Close"
	case "open":
		return "Open"
	case "high":
		return "High"
	case "low":
		return "Low"
	case "volume":
		return "Volume"
	default:
		return "Close"
	}
}

// Classify analyzes a source expression string and returns its classification.
// Deprecated: Use ClassifyAST for AST-based analysis to avoid code generation artifacts.
func (c *SeriesSourceClassifier) Classify(sourceExpr string) SourceInfo {
	info := SourceInfo{
		OriginalExpr: sourceExpr,
	}

	cleanExpr := sourceExpr
	for len(cleanExpr) > 0 && (cleanExpr[0] == '-' || cleanExpr[0] == '+' || cleanExpr[0] == '!') {
		cleanExpr = cleanExpr[1:]
	}

	if len(cleanExpr) > 2 && cleanExpr[0] == '(' && cleanExpr[len(cleanExpr)-1] == ')' {
		cleanExpr = cleanExpr[1 : len(cleanExpr)-1]
	}

	if varName := c.extractSeriesVariableName(cleanExpr); varName != "" {
		info.Type = SourceTypeSeriesVariable
		info.VariableName = varName
		return info
	}

	info.Type = SourceTypeOHLCVField
	info.FieldName = c.extractOHLCVFieldName(cleanExpr)
	return info
}

func (c *SeriesSourceClassifier) extractSeriesVariableName(expr string) string {
	matches := c.seriesVariablePattern.FindStringSubmatch(expr)
	if len(matches) == 2 {
		return matches[1]
	}
	return ""
}

func (c *SeriesSourceClassifier) extractOHLCVFieldName(expr string) string {
	if lastDotIndex := findLastDotIndex(expr); lastDotIndex >= 0 {
		return expr[lastDotIndex+1:]
	}
	return expr
}

func findLastDotIndex(s string) int {
	for i := len(s) - 1; i >= 0; i-- {
		if s[i] == '.' {
			return i
		}
	}
	return -1
}
