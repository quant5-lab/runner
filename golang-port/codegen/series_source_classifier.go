package codegen

import "regexp"

// SourceType represents the category of data source for technical analysis calculations.
type SourceType int

const (
	SourceTypeUnknown SourceType = iota
	SourceTypeSeriesVariable
	SourceTypeOHLCVField
)

// SourceInfo contains classification results for a source expression.
type SourceInfo struct {
	Type         SourceType
	VariableName string
	FieldName    string
	OriginalExpr string
}

// IsSeriesVariable returns true if the source is a user-defined Series variable.
func (s SourceInfo) IsSeriesVariable() bool {
	return s.Type == SourceTypeSeriesVariable
}

// IsOHLCVField returns true if the source is a built-in OHLCV field.
func (s SourceInfo) IsOHLCVField() bool {
	return s.Type == SourceTypeOHLCVField
}

// SeriesSourceClassifier analyzes source expressions to determine their type.
type SeriesSourceClassifier struct {
	seriesVariablePattern *regexp.Regexp
}

// NewSeriesSourceClassifier creates a classifier for series source expressions.
func NewSeriesSourceClassifier() *SeriesSourceClassifier {
	return &SeriesSourceClassifier{
		seriesVariablePattern: regexp.MustCompile(`^([A-Za-z_][A-Za-z0-9_]*)Series\.Get`),
	}
}

// Classify analyzes a source expression and returns its classification.
func (c *SeriesSourceClassifier) Classify(sourceExpr string) SourceInfo {
	info := SourceInfo{
		OriginalExpr: sourceExpr,
	}

	if varName := c.extractSeriesVariableName(sourceExpr); varName != "" {
		info.Type = SourceTypeSeriesVariable
		info.VariableName = varName
		return info
	}

	info.Type = SourceTypeOHLCVField
	info.FieldName = c.extractOHLCVFieldName(sourceExpr)
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
