package codegen

import (
	"fmt"
	"strings"
)

/* SeriesWindowExtractor converts Series to array window for runtime indicator calls */
type SeriesWindowExtractor struct{}

func NewSeriesWindowExtractor() *SeriesWindowExtractor {
	return &SeriesWindowExtractor{}
}

func (e *SeriesWindowExtractor) GenerateExtractionCode(
	sourceExpr string,
	windowSize string,
	indenter func() string,
) string {
	out := &strings.Builder{}
	ind := indenter

	out.WriteString(ind() + fmt.Sprintf("sourceWindow := make([]float64, %s)\n", windowSize))
	out.WriteString(ind() + fmt.Sprintf("for j := 0; j < %s; j++ {\n", windowSize))

	accessExpr := e.buildHistoricalAccess(sourceExpr, "j")
	out.WriteString(ind() + "\t" + fmt.Sprintf("sourceWindow[j] = %s\n", accessExpr))

	out.WriteString(ind() + "}\n\n")

	return out.String()
}

func (e *SeriesWindowExtractor) buildHistoricalAccess(sourceExpr, offsetVar string) string {
	if strings.HasSuffix(sourceExpr, "Series") {
		return fmt.Sprintf("%s.Get(%s)", sourceExpr, offsetVar)
	}

	if strings.Contains(sourceExpr, "Series.Get(0)") {
		return strings.Replace(sourceExpr, ".Get(0)", fmt.Sprintf(".Get(%s)", offsetVar), 1)
	}

	if strings.Contains(sourceExpr, "Series.GetCurrent()") {
		return strings.Replace(sourceExpr, ".GetCurrent()", fmt.Sprintf(".Get(%s)", offsetVar), 1)
	}

	if strings.HasPrefix(sourceExpr, "bar.") {
		field := strings.TrimPrefix(sourceExpr, "bar.")
		return fmt.Sprintf("ctx.Data[%s].%s", offsetVar, field)
	}

	return fmt.Sprintf("%s[%s]", sourceExpr, offsetVar)
}

func (e *SeriesWindowExtractor) ExtractHistoricalValue(sourceExpr, offsetExpr string) string {
	return e.buildHistoricalAccess(sourceExpr, offsetExpr)
}
