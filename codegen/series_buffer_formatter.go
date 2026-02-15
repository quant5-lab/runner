package codegen

import "fmt"

// formatSeriesGet generates code for reading from a series buffer
func formatSeriesGet(varName string, offset int) string {
	return fmt.Sprintf("%sSeries.Get(%d)", varName, offset)
}

// formatSeriesDynamicGet generates code for reading with a runtime-evaluated offset expression
func formatSeriesDynamicGet(varName string, offsetExpr string) string {
	return fmt.Sprintf("%sSeries.Get(%s)", varName, offsetExpr)
}

// formatSeriesSet generates code for writing to a series buffer
func formatSeriesSet(varName string, value string) string {
	return fmt.Sprintf("%sSeries.Set(%s)", varName, value)
}

// formatArrowSeriesGet generates code for reading from arrow context series
func formatArrowSeriesGet(varName string, offset int) string {
	return fmt.Sprintf("arrowCtx.GetOrCreateSeries(%q).Get(%d)", varName, offset)
}

// formatArrowSeriesDynamicGet generates code for reading from arrow context series with runtime offset
func formatArrowSeriesDynamicGet(varName string, offsetExpr string) string {
	return fmt.Sprintf("arrowCtx.GetOrCreateSeries(%q).Get(%s)", varName, offsetExpr)
}

// formatArrowSeriesSet generates code for writing to arrow context series
func formatArrowSeriesSet(varName string, value string) string {
	return fmt.Sprintf("arrowCtx.GetOrCreateSeries(%q).Set(%s)", varName, value)
}
